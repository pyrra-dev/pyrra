import React, {useEffect, useMemo, useReducer, useState} from 'react'
import {API_BASEPATH, latencyTarget} from '../App'
import {useNavigate} from 'react-router-dom'
import {useQueryState, parseAsString} from 'nuqs'
import Navbar from '../components/Navbar'
import {type Labels, labelsString, MetricName} from '../labels'
import {parseAsLabels} from '../searchParams'
import {createConnectTransport} from '@connectrpc/connect-web'
import {createClient, Code} from '@connectrpc/connect'
import {clone} from '@bufbuild/protobuf'
import {ObjectiveService, ObjectiveSchema} from '../proto/objectives/v1alpha1/objectives_pb'
import {
  type Alert as ObjectiveAlert,
  Alert_State,
  type GetAlertsResponse,
  type GetStatusResponse,
  type Objective,
  type ObjectiveStatus,
} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration} from '../duration'
import {
  type Cell,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  type Row as TableRowType,
  type SortingFnOption,
  type SortingState,
  useReactTable,
  type VisibilityState,
} from '@tanstack/react-table'
import {type Duration} from '@bufbuild/protobuf/wkt'
import {useObjectivesList} from '../objectives'
import {ArrowDown, ArrowUp, ArrowUpDown, Columns2, Search, TriangleAlert} from 'lucide-react'
import {Badge} from '@/components/ui/badge'
import {Alert, AlertDescription, AlertTitle} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Spinner} from '@/components/ui/spinner'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
} from '@/components/ui/dropdown-menu'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import {cn} from '@/lib/utils'

enum TableObjectiveState {
  Unknown,
  Success,
  NoData,
  Error,
}

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective {
  objective: Objective
  lset: string
  groupingLabels: Labels
  state: TableObjectiveState
  availability?: TableAvailability | null
  budget?: number | null
  severity: string | null
  latency: number | undefined
}

interface TableAvailability {
  errors: number
  total: number
  percentage: number
}

interface TableState {
  objectives: Record<string, TableObjective>
}

enum TableActionType {
  SetObjective,
  DeleteObjective,
  SetStatus,
  SetObjectiveWithStatus,
  SetStatusNone,
  SetStatusError,
  SetAlert,
}

type TableAction =
  | {type: TableActionType.SetObjective; lset: string; objective: Objective}
  | {type: TableActionType.DeleteObjective; lset: string}
  | {type: TableActionType.SetStatus; lset: string; status: ObjectiveStatus}
  | {
      type: TableActionType.SetObjectiveWithStatus
      lset: string
      statusLabels: Labels
      objective: Objective
      status: ObjectiveStatus
    }
  | {type: TableActionType.SetStatusNone; lset: string}
  | {type: TableActionType.SetStatusError; lset: string}
  | {type: TableActionType.SetAlert; labels: Labels; severity: string}

const tableReducer = (state: TableState, action: TableAction): TableState => {
  switch (action.type) {
    case TableActionType.SetObjective:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            objective: action.objective,
            lset: action.lset,
            groupingLabels: {},
            state: TableObjectiveState.Unknown,
            severity: null,
            availability: undefined,
            budget: undefined,
            latency: latencyTarget(action.objective),
          },
        },
      }
    case TableActionType.DeleteObjective: {
      const {[action.lset]: _, ...cleanedObjective} = state.objectives
      return {
        objectives: {...cleanedObjective},
      }
    }
    case TableActionType.SetStatus:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            ...state.objectives[action.lset],
            state: TableObjectiveState.Success,
            availability: {
              errors: action.status.availability?.errors ?? 0,
              total: action.status.availability?.total ?? 0,
              percentage: action.status.availability?.percentage ?? 0,
            },
            budget: action.status.budget?.remaining,
          },
        },
      }
    case TableActionType.SetObjectiveWithStatus: {
      // It is possible that we may need to merge some previous state
      let severity: string | null = null
      const o = state.objectives[action.lset]
      if (o !== undefined) {
        severity = o.severity
      }

      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            objective: action.objective,
            lset: action.lset,
            groupingLabels: action.statusLabels,
            state: TableObjectiveState.Success,
            severity,
            availability: {
              errors: action.status.availability?.errors ?? 0,
              total: action.status.availability?.total ?? 0,
              percentage: action.status.availability?.percentage ?? 0,
            },
            budget: action.status.budget?.remaining,
            latency: latencyTarget(action.objective),
          },
        },
      }
    }
    case TableActionType.SetStatusNone:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            ...state.objectives[action.lset],
            state: TableObjectiveState.NoData,
            availability: null,
            budget: null,
          },
        },
      }
    case TableActionType.SetStatusError:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            ...state.objectives[action.lset],
            state: TableObjectiveState.Error,
            availability: null,
            budget: null,
          },
        },
      }
    case TableActionType.SetAlert: {
      // Find the objective this alert's labels is the super set for.
      const result = Object.entries(state.objectives).find(([, o]) => {
        const allLabels: Labels = {...o.objective.labels, ...o.groupingLabels}

        let isSuperset = true
        Object.entries(action.labels).forEach(([k, v]) => {
          if (allLabels[k] === undefined) {
            return false
          }
          if (allLabels[k] !== v) {
            isSuperset = false
          }
        })
        return isSuperset
      })
      if (result === undefined) {
        return state
      }

      const [lset, objective] = result

      // If there is one multi burn rate with critical severity firing we want to only show that one.
      if (objective.severity === 'critical' && action.severity === 'warning') {
        return state
      }

      return {
        objectives: {
          ...state.objectives,
          [lset]: {
            ...state.objectives[lset],
            severity: action.severity,
          },
        },
      }
    }
    default:
      return state
  }
}

interface row {
  lset: {lset: Labels; grouping: Labels}
  window: Duration | undefined
  objective: number
  latency: number | undefined
  errors: number | undefined
  total: number | undefined
  availability: number | undefined
  budget: number | undefined | null
  alerts: string
}

const columnHelper = createColumnHelper<row>()
const sortingNumberNull: SortingFnOption<row> = (
  rowA: TableRowType<row>,
  rowB: TableRowType<row>,
  columnId: string,
): number => {
  const av: number | null = rowA.getValue(columnId)
  const bv: number | null = rowB.getValue(columnId)
  if (av !== null && bv !== null) {
    return av - bv
  }
  if (av === null && bv !== null) {
    return 1
  }
  if (av !== null && bv === null) {
    return -1
  }
  return 0
}

const columns = [
  columnHelper.accessor('lset', {
    id: 'lset',
    header: 'Name',
    sortingFn: (a: TableRowType<row>, b: TableRowType<row>, columnId: string): number => {
      const av: {lset: Labels; grouping: Labels} = a.getValue(columnId)
      const bv: {lset: Labels; grouping: Labels} = b.getValue(columnId)

      // First compare by the metric name
      const aName = av.lset[MetricName] ?? ''
      const bName = bv.lset[MetricName] ?? ''
      const nameComparison = aName.localeCompare(bName)
      if (nameComparison !== 0) {
        return nameComparison
      }

      // If names are equal, compare by grouping labels only
      const aGrouping = labelsString(av.grouping)
      const bGrouping = labelsString(bv.grouping)
      return aGrouping.localeCompare(bGrouping)
    },
  }),
  columnHelper.accessor('window', {
    id: 'window',
    header: 'Window',
    cell: (props) => {
      const window = props.getValue()
      if (Number(window?.seconds) === 0) {
        return <span className="text-destructive">Error: Invalid SLO configuration</span>
      }
      return formatDuration(Number(window?.seconds) * 1000)
    },
    sortingFn: (a, b, column): number => {
      const av: Duration = a.getValue(column)
      const bv: Duration = b.getValue(column)
      return Number(av.seconds - bv.seconds)
    },
  }),
  columnHelper.accessor('objective', {
    id: 'objective',
    header: 'Objective',
    cell: (props) => `${(100 * props.getValue()).toFixed(2)}%`,
  }),
  columnHelper.accessor('latency', {
    id: 'latency',
    header: 'Latency',
    cell: (props): string => {
      const v = props.getValue()
      if (v === undefined) {
        return ''
      }
      return formatDuration(v)
    },
  }),
  columnHelper.accessor('errors', {
    id: 'errors',
    header: 'Errors',
    cell: (props) => {
      const v = props.getValue()
      if (v === undefined) {
        return 'No data'
      }
      return `${Math.floor(v).toLocaleString()}`
    },
    sortingFn: sortingNumberNull,
  }),
  columnHelper.accessor('total', {
    id: 'total',
    header: 'Total',
    cell: (props) => {
      const v = props.getValue()
      if (v === undefined) {
        return 'No data'
      }

      const objective: number = props.row.getValue('objective')
      return (
        <>
          {Math.floor(v).toLocaleString()}
          <VolumeWarningTooltip id={props.cell.id} objective={objective} total={v} />
        </>
      )
    },
    sortingFn: sortingNumberNull,
  }),
  columnHelper.accessor('availability', {
    id: 'availability',
    header: 'Availability',
    cell: (props) => {
      const v = props.getValue()
      const target = props.row.getValue<number>('objective') ?? 0
      if (v === undefined) {
        return 'No data'
      }

      const objective: number = props.row.getValue('objective')
      const total: number = props.row.getValue('total')
      const totalVisible = props.table.getColumn('total')?.getIsVisible() ?? false

      return (
        <>
          <span className={cn(v <= target && 'text-[#b10d0d]')}>
            {(100 * v).toFixed(2)}%
          </span>
          {!totalVisible && (
            <VolumeWarningTooltip id={props.cell.id} objective={objective} total={total} />
          )}
        </>
      )
    },
    sortingFn: sortingNumberNull,
  }),
  columnHelper.accessor('budget', {
    id: 'budget',
    header: 'Budget',
    cell: (props) => {
      const v = props.getValue()
      if (v === undefined || v === null) {
        return 'No data'
      }
      return (
        <span className={cn(v < 0 && 'text-[#b10d0d]')}>
          {(100 * v).toFixed(2)}%
        </span>
      )
    },
    sortingFn: sortingNumberNull,
  }),
  columnHelper.accessor('alerts', {
    id: 'alerts',
    header: 'Alerts',
    cell: (props) => {
      const v = props.getValue()
      if (v === '') {
        return
      }
      return <span className="text-destructive font-semibold">{v}</span>
    },
  }),
]

const VolumeWarningTooltip = ({
  id,
  objective,
  total,
}: {
  id: string
  objective: number
  total: number
}): React.JSX.Element => {
  const show = (1 - objective) * total < 1

  if (!show) return <></>

  return (
    <span
      className="ml-2 inline-flex align-middle"
      title="Too few requests! Adjust your objective or wait for events.">
      <TriangleAlert size={20} color="#b10d0d" />
    </span>
  )
}

const emptyData: row[] = []

const List = () => {
  console.log('render List')

  document.title = 'Objectives - Pyrra'
  const navigate = useNavigate()

  const client = useMemo(() => {
    const baseUrl = API_BASEPATH ?? 'http://localhost:9099'
    return createClient(ObjectiveService, createConnectTransport({baseUrl}))
  }, [])

  const [filterSearch, setFilterSearch] = useQueryState('search', parseAsString.withDefault(''))
  const [filterLabels, setFilterLabels] = useQueryState('filter', parseAsLabels.withDefault({}))
  const [filterError] = useState<boolean>(false)

  // TODO: Pass in the search to the useObjectivesList hook
  const {
    response: objectiveResponse,
    error: objectiveError,
    status: objectiveStatus,
  } = useObjectivesList(client, labelsString(filterLabels), '')

  const initialTableState: TableState = {objectives: {}}
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)

  const [sorting, setSorting] = React.useState<SortingState>([{id: 'budget', desc: false}])
  // TODO: Persist the column visibility in the browser's state
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({
    lset: true,
    window: true,
    objective: true,
    latency: true,
    errors: false,
    total: false,
    availability: true,
    budget: true,
    alerts: true,
  })

  const updateFilter = (lset: Labels) => {
    const updated: Labels = {...filterLabels, ...lset}
    void setFilterLabels(Object.keys(updated).length > 0 ? updated : null)
  }

  const removeFilterLabel = (k: string) => {
    const {[k]: _, ...rest} = filterLabels
    void setFilterLabels(Object.keys(rest).length > 0 ? rest : null)
  }

  useEffect(() => {
    if (objectiveStatus === 'success' && objectiveResponse !== null) {
      client
        .getAlerts({
          expr: '',
          inactive: false,
          current: false,
        })
        .then((resp: GetAlertsResponse) => {
          resp.alerts.forEach((a: ObjectiveAlert) => {
            if (a.state === Alert_State.firing) {
              dispatchTable({
                type: TableActionType.SetAlert,
                labels: a.labels,
                severity: a.severity,
              })
            }
          })
        })
        .catch((err) => { console.log(err); })

      objectiveResponse.objectives.forEach((o: Objective) => {
        dispatchTable({
          type: TableActionType.SetObjective,
          lset: labelsString(o.labels),
          objective: o,
        })

        client
          .getStatus({expr: labelsString(o.labels)})
          .then((resp: GetStatusResponse) => {
            if (resp.status.length === 0) {
              dispatchTable({type: TableActionType.SetStatusNone, lset: labelsString(o.labels)})
            } else if (resp.status.length === 1) {
              dispatchTable({
                type: TableActionType.SetStatus,
                lset: labelsString(o.labels),
                status: resp.status[0],
              })
            } else {
              dispatchTable({type: TableActionType.DeleteObjective, lset: labelsString(o.labels)})

              resp.status.forEach((s: ObjectiveStatus) => {
                const so = clone(ObjectiveSchema, o)
                // Identify by the combined labels
                const sLabels: Labels = s.labels ?? {}
                const soLabels: Labels = {...o.labels, ...sLabels}

                dispatchTable({
                  type: TableActionType.SetObjectiveWithStatus,
                  lset: labelsString(soLabels),
                  statusLabels: sLabels,
                  objective: so,
                  status: s,
                })
              })
            }
          })
          .catch((err) => {
            console.log(err)
            dispatchTable({type: TableActionType.SetStatusError, lset: labelsString(o.labels)})
          })
      })
    }
  }, [client, objectiveResponse, objectiveStatus])

  const reactTableData = useMemo(() => {
    if (Object.keys(table.objectives).length === 0) {
      return emptyData
    }
    return (
      Object.keys(table.objectives)
        // TODO: Replace with a react-table filter.
        // I couldn't make it work...
        .filter((k): boolean => {
          // first filter for labels filtered for
          const o = table.objectives[k]
          const labels = {...o.objective.labels, ...o.groupingLabels}
          for (const k in filterLabels) {
            // if label doesn't exist by key or if values differ filter out.
            if (labels[k] === undefined || labels[k] !== filterLabels[k]) {
              return false
            }
          }
          // if all labels match, filter for the search string
          // TODO: Use fuzzy search
          return k.toLowerCase().includes(filterSearch.toLowerCase())
        })
        .map((k: string) => {
          const o = table.objectives[k]
          const r: row = {
            lset: {lset: o.objective.labels, grouping: o.groupingLabels},
            window: o.objective.window,
            objective: o.objective.target,
            latency: o.latency,
            errors: o.availability?.errors,
            total: o.availability?.total,
            availability: o.availability?.percentage,
            budget: o.budget,
            alerts: o.severity ?? '',
          }
          return r
        }) ?? null
    )
  }, [table.objectives, filterSearch, filterLabels])

  const reactTable = useReactTable({
    data: reactTableData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    state: {
      sorting,
      columnVisibility,
    },
    // sorting
    getSortedRowModel: getSortedRowModel(),
    enableSortingRemoval: false,
    onSortingChange: setSorting,
    sortingFns: {
      sortingNumberNull,
    },
    // debugTable: true,
    // debugHeaders: true,
    // debugColumns: true,
  })

  const objectivePage = (labels: Labels, grouping: Labels) => {
    return `/objectives?expr=${encodeURI(labelsString(labels))}&grouping=${encodeURI(
      labelsString(grouping),
    )}`
  }

  if (objectiveStatus === 'pending') {
    return (
      <>
        <Navbar />
        <div className="container-responsive mt-[100px]">
          <div className="mt-3 flex justify-center">
            <div className="text-center">
              <Spinner />
              <p className="mt-3">Loading objectives...</p>
            </div>
          </div>
        </div>
      </>
    )
  }

  if (objectiveError !== null && objectiveError !== undefined) {
    return (
      <>
        <Navbar />
        <div className="container-responsive mt-[100px]">
          <div className="mt-3">
            <div>
              {objectiveError.code === Code.Unavailable && (
                <Alert variant="destructive">
                  <AlertTitle>Backend connection failed</AlertTitle>
                  <AlertDescription>
                    Cannot reach the backend service. Ensure the <b>filesystem</b> or{' '}
                    <b>Kubernetes</b> backend is running.
                  </AlertDescription>
                </Alert>
              )}
              {objectiveError.code !== Code.NotFound &&
                objectiveError.code !== Code.Unavailable && (
                  <Alert variant="destructive">
                    <AlertTitle>Error loading objectives</AlertTitle>
                    <AlertDescription>{objectiveError.message}</AlertDescription>
                  </Alert>
                )}
            </div>
          </div>
        </div>
      </>
    )
  }

  return (
    <>
      <Navbar />
      <div className="container-responsive mt-[100px]">
        <div>
          <div>
            <h3 className="mb-8">Service Level Objectives</h3>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-y-2">
          <div className="my-2 w-full md:w-1/2 lg:w-1/3">
            <div className="relative">
              <div className="absolute top-1/2 left-3 -translate-y-1/2">
                <Search size={16} />
              </div>
              <input
                type="search"
                className="h-10 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm"
                placeholder="Search name"
                aria-label="Search"
                value={filterSearch}
                onChange={(e) => {
                  void setFilterSearch(e.target.value !== '' ? e.target.value : null)
                }}
              />
            </div>
          </div>
          <div className="my-2 order-2 lg:order-1 lg:flex-1">
            {Object.keys(filterLabels)
              .sort((a, b) => a.localeCompare(b))
              .map((k: string) => (
                <Button
                  key={k}
                  variant="secondary"
                  size="sm"
                  className="mr-2 mb-2 md:mb-0"
                  onClick={() => { removeFilterLabel(k); }}>
                  {`${k}=${filterLabels[k]}`}
                  <span className="ml-1 text-xs">✕</span>
                </Button>
              ))}
            {filterError && (
              <Alert variant="destructive">
                <AlertDescription>
                  Your SLO filter is broken. Please reset the filter.
                </AlertDescription>
              </Alert>
            )}
          </div>
          <div className="my-2 order-1 md:w-1/2 lg:order-2 lg:w-auto text-right">
            <DropdownMenu>
              <DropdownMenuTrigger
                className="inline-flex shrink-0 items-center justify-center rounded-md border border-border bg-background px-2.5 py-1 text-sm font-medium hover:bg-muted">
                <Columns2 size={16} />
                <span className="ml-2">Columns</span>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {columns.map((c) => {
                  const id = c.id ?? ''
                  const header = c.header as string
                  return (
                    <DropdownMenuCheckboxItem
                      key={id}
                      checked={columnVisibility[id]}
                      onCheckedChange={() => {
                        setColumnVisibility({
                          ...columnVisibility,
                          [id]: !columnVisibility[id],
                        })
                      }}>
                      {header}
                    </DropdownMenuCheckboxItem>
                  )
                })}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        <div className="mt-2">
          <Table>
            <TableHeader>
              {reactTable.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead
                      key={header.id}
                      className={cn(
                        'cursor-pointer select-none hover:text-foreground [&_svg]:h-4 [&_svg]:w-4 [&_svg]:-mt-1 [&_svg]:inline',
                        header.column.getIsSorted() !== false && 'text-foreground [&_svg_path]:stroke-foreground',
                      )}
                      onClick={header.column.getToggleSortingHandler()}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.header, header.getContext())}
                      {header.column.getCanSort() && header.column.getIsSorted() === false ? (
                        <ArrowUpDown />
                      ) : (
                        ''
                      )}
                      {header.column.getIsSorted() === 'asc' ? <ArrowUp /> : ''}
                      {header.column.getIsSorted() === 'desc' ? <ArrowDown /> : ''}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {reactTable.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className={cn(
                    'cursor-pointer',
                    row.getValue('alerts') !== '' &&
                      'bg-destructive/20 border-l-4 border-l-destructive',
                  )}
                  onClick={() => {
                    const window: Duration | undefined = row.getValue('window')
                    if (Number(window?.seconds) === 0) {
                      // Don't navigate for invalid SLOs
                      return
                    }
                    const labels: {lset: Labels; grouping: Labels} = row.getValue('lset')
                    void navigate(objectivePage(labels.lset, labels.grouping))
                  }}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      <>
                        {cell.column.id === 'lset' ? (
                          <NameCell
                            cell={cell}
                            onFilter={(lset) => {
                              updateFilter(lset)
                            }}
                          />
                        ) : (
                          flexRender(cell.column.columnDef.cell, cell.getContext())
                        )}
                      </>
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <div>
          <div>
            <small>
              All availabilities and error budgets are calculated across the entire time window of
              the objective.
            </small>
          </div>
        </div>
      </div>
    </>
  )
}

interface NameCellProps {
  cell: Cell<row, any>
  onFilter: (lset: Labels) => void
}

const NameCell = ({cell, onFilter}: NameCellProps): React.JSX.Element => {
  const v: {lset: Labels; grouping: Labels} = cell.getValue()
  const name = v.lset[MetricName]
  const labelBadges = Object.entries({...v.lset, ...v.grouping})
    .filter((l: [string, string]) => l[0] !== MetricName)
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map((l: [string, string]) => (
      <Badge
        key={l[0]}
        variant="secondary"
        className="mr-1 cursor-pointer font-normal"
        onClick={(event) => {
          event.stopPropagation()
          const lset: Labels = {}
          lset[l[0]] = l[1]
          console.log('filter', lset)
          onFilter(lset)
        }}>
        <a>
          {l[0]}={l[1]}
        </a>
      </Badge>
    ))

  return (
    <>
      <a>{name}</a> {labelBadges}
    </>
  )
}

export default List
