import React, {useEffect, useMemo, useReducer, useState} from 'react'
import {
  Alert,
  Badge,
  Button,
  Col,
  Container,
  Dropdown,
  OverlayTrigger,
  Row,
  Table,
  Tooltip as OverlayTooltip,
} from 'react-bootstrap'
import {API_BASEPATH, latencyTarget} from '../App'
import {useLocation, useNavigate} from 'react-router-dom'
import Navbar from '../components/Navbar'
import {Labels, labelsString, MetricName, parseLabels} from '../labels'
import {createConnectTransport} from '@bufbuild/connect-web'
import {createPromiseClient} from '@connectrpc/connect'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connect'
import {
  Alert as ObjectiveAlert,
  Alert_State,
  GetAlertsResponse,
  GetStatusResponse,
  Objective,
  ObjectiveStatus,
} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration} from '../duration'
import {
  Cell,
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  Row as TableRow,
  SortingFnOption,
  SortingState,
  useReactTable,
  VisibilityState,
} from '@tanstack/react-table'
import {Duration} from '@bufbuild/protobuf'
import {useObjectivesList} from '../objectives'
import {
  IconArrowDown,
  IconArrowUp,
  IconArrowUpDown,
  IconMagnifyingGlass,
  IconTableColumns,
  IconWarning,
} from '../components/Icons'
import AwesomeDebouncePromise from 'awesome-debounce-promise'
import useConstant from 'use-constant'
import {useAsync} from 'react-async-hook'

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
  objectives: {[key: string]: TableObjective}
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
            severity: severity,
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
  rowA: TableRow<row>,
  rowB: TableRow<row>,
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
  }),
  columnHelper.accessor('window', {
    id: 'window',
    header: 'Window',
    cell: (props) => {
      return formatDuration(Number(props.getValue()?.seconds) * 1000)
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
          <span className={v > target ? 'good' : 'bad'}>{(100 * v).toFixed(2)}%</span>
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
      return <span className={v >= 0 ? 'good' : 'bad'}>{(100 * v).toFixed(2)}%</span>
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
      return <span className="severity">{v}</span>
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
    <OverlayTrigger
      key={`${id}-warning`}
      overlay={
        <OverlayTooltip id={`tooltip-${id}-warning`}>
          Too few requests!
          <br />
          Adjust your objective or wait for events.
        </OverlayTooltip>
      }>
      <span className="volume-warning">
        <IconWarning width={20} height={20} fill="#b10d0d" />
      </span>
    </OverlayTrigger>
  )
}

const emptyData: row[] = []

const List = () => {
  console.log('render List')

  document.title = 'Objectives - Pyrra'
  const navigate = useNavigate()
  const {search} = useLocation()

  const client = useMemo(() => {
    const baseUrl = API_BASEPATH === undefined ? 'http://localhost:9099' : API_BASEPATH
    return createPromiseClient(ObjectiveService, createConnectTransport({baseUrl}))
  }, [])

  const [filterSearch] = useMemo((): [string] => {
    const query = new URLSearchParams(search)
    const querySearch = query.get('search')
    return [querySearch ?? '']
  }, [search])

  const [filterLabels, filterError] = useMemo((): [Labels, boolean] => {
    const query = new URLSearchParams(search)
    const queryFilter = query.get('filter')
    try {
      if (queryFilter !== null) {
        if (queryFilter.indexOf('=') > 0) {
          return [parseLabels(queryFilter), false]
        } else {
          filterLabels[MetricName] = queryFilter
          return [filterLabels, false]
        }
      }
    } catch (e) {
      console.log(e)
      return [{}, true]
    }
    return [{}, false]
  }, [search])

  // TODO: Pass in the search to the useObjectivesList hook
  const {
    response: objectiveResponse,
    // error: objectiveError,
    status: objectiveStatus,
  } = useObjectivesList(client, labelsString(filterLabels), '')

  const navigateFilterURL = (search: string, labels: Labels) => {
    // TODO: use something like URLState?

    const hasSearch = search.length > 0
    const hasLabels = Object.keys(labels).length > 0
    console.log('hasSearch', hasSearch, 'hasLabels', hasLabels, labels)

    if (!hasSearch && !hasLabels) {
      navigate('?')
      return
    }
    if (hasSearch && !hasLabels) {
      navigate(`?search=${encodeURI(search)}`)
      return
    }
    if (!hasSearch && hasLabels) {
      navigate(`?filter=${encodeURI(labelsString(labels))}`)
      return
    }

    navigate(`?search=${encodeURI(search)}&filter=${encodeURI(labelsString(labels))}`)
  }

  const initialTableState: TableState = {objectives: {}}
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)

  const [sorting, setSorting] = React.useState<SortingState>([{id: 'budget', desc: false}])
  const [searchInput, setSearchInput] = React.useState<string>(filterSearch)

  // Only update the URL when the user stops typing for 1 second
  const debouncedSearchFunction = useConstant(() =>
    AwesomeDebouncePromise((search: string, labels: Labels) => {
      navigateFilterURL(search, labels)
    }, 1000),
  )

  // The async callback is run each time the text changes,
  // but as the search function is debounced, it does not
  // fire a new request on each keystroke
  useAsync(async () => {
    return debouncedSearchFunction(searchInput, filterLabels)
  }, [debouncedSearchFunction, searchInput, filterLabels])

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
    // Copy existing filterLabels (from router) and add/overwrite k-v-pairs
    const updatedFilter: Labels = {...filterLabels}
    for (const l in lset) {
      updatedFilter[l] = lset[l]
    }
    navigateFilterURL(searchInput, updatedFilter)
  }

  const removeFilterLabel = (k: string) => {
    const updatedFilter: Labels = {}
    for (const name in filterLabels) {
      if (name !== k) {
        updatedFilter[name] = filterLabels[name]
      }
    }
    navigateFilterURL(searchInput, updatedFilter)
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
        .catch((err) => console.log(err))

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
                const so = o.clone()
                // Identify by the combined labels
                const sLabels: Labels = s.labels !== undefined ? s.labels : {}
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
          return k.toLowerCase().includes(searchInput.toLowerCase())
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
            alerts: o.severity !== null ? o.severity : '',
          }
          return r
        }) ?? null
    )
  }, [table.objectives, searchInput, filterLabels])

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

  return (
    <>
      <Navbar />
      <Container className="content list">
        <Row>
          <Col>
            <h3>Service Level Objectives</h3>
          </Col>
        </Row>
        <Row className="align-items-center">
          <Col xs={12} md={6} lg={4} className="my-2">
            <div className="position-relative">
              <div className="position-absolute" style={{top: 5, left: 14}}>
                <IconMagnifyingGlass width={16} height={16} />
              </div>
              <input
                type="search"
                className="form-control"
                placeholder="Search name"
                aria-label="Search"
                style={{paddingLeft: 40}}
                value={searchInput}
                onChange={(e) => {
                  setSearchInput(e.target.value)
                }}
              />
            </div>
          </Col>
          <Col className="my-2 order-md-2 order-lg-1 flex-lg-fill" xs={12} lg={'auto'}>
            {Object.keys(filterLabels)
              .sort((a, b) => a.localeCompare(b))
              .map((k: string) => (
                <Button
                  key={k}
                  variant="light"
                  size="sm"
                  className="filter-close"
                  onClick={() => removeFilterLabel(k)}>
                  {`${k}=${filterLabels[k]}`}
                  <span className="btn-close"></span>
                </Button>
              ))}
            <Alert show={filterError} variant="danger">
              Your SLO filter is broken. Please reset the filter.
            </Alert>
          </Col>
          <Col
            xs={12}
            md={6}
            lg={'auto'}
            className="my-2 order-md-1 order-lg-2"
            style={{textAlign: 'right'}}>
            <Dropdown>
              <Dropdown.Toggle
                variant="outline-light"
                id="dropdown-basic"
                role="checkbox"
                style={{color: 'var(--bs-body-color)', justifyContent: 'center'}}>
                <IconTableColumns width={16} height={16} />
                <span style={{marginLeft: 8}}>Columns</span>
              </Dropdown.Toggle>
              <Dropdown.Menu align="end">
                <ul style={{listStyle: 'none', padding: '0 10px'}}>
                  {columns.map((c) => {
                    const id = c.id ?? ''
                    const header = c.header as string
                    return (
                      <li key={id} style={{padding: '4px 0'}}>
                        <input
                          type="checkbox"
                          checked={columnVisibility[id]}
                          id={id}
                          onChange={() =>
                            setColumnVisibility({
                              ...columnVisibility,
                              [id]: !columnVisibility[id],
                            })
                          }
                        />
                        <label htmlFor={id} style={{marginLeft: 8}}>
                          {header}
                        </label>
                      </li>
                    )
                  })}
                </ul>
              </Dropdown.Menu>
            </Dropdown>
          </Col>
        </Row>
        <Row>
          <div className="table-responsive">
            <Table hover={true}>
              <thead>
                {reactTable.getHeaderGroups().map((headerGroup) => (
                  <tr key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <th
                        key={header.id}
                        className={header.column.getIsSorted() !== false ? 'active' : ''}
                        onClick={header.column.getToggleSortingHandler()}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getCanSort() && header.column.getIsSorted() === false ? (
                          <IconArrowUpDown />
                        ) : (
                          ''
                        )}
                        {header.column.getIsSorted() === 'asc' ? <IconArrowUp /> : ''}
                        {header.column.getIsSorted() === 'desc' ? <IconArrowDown /> : ''}
                      </th>
                    ))}
                  </tr>
                ))}
              </thead>
              <tbody>
                {reactTable.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    onClick={() => {
                      const labels: {lset: Labels; grouping: Labels} = row.getValue('lset')
                      navigate(objectivePage(labels.lset, labels.grouping))
                    }}
                    className={
                      row.getValue('alerts') !== ''
                        ? 'table-row-clickable firing'
                        : 'table-row-clickable'
                    }>
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id}>
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
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </Row>
        <Row>
          <Col>
            <small>
              All availabilities and error budgets are calculated across the entire time window of
              the objective.
            </small>
          </Col>
        </Row>
      </Container>
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
        bg="light"
        text="dark"
        className="fw-normal"
        style={{marginRight: 5}}
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
