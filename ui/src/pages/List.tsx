import React, {useEffect, useMemo, useReducer, useState} from 'react'
import {
  Alert,
  Badge,
  Button,
  Col,
  Container,
  OverlayTrigger,
  Row,
  Spinner,
  Table,
  Tooltip as OverlayTooltip,
} from 'react-bootstrap'
import {
  API_BASEPATH,
  hasObjectiveType,
  latencyTarget,
  ObjectiveType,
  renderLatencyTarget,
} from '../App'
import {Link, useLocation, useNavigate} from 'react-router-dom'
import Navbar from '../components/Navbar'
import {IconArrowDown, IconArrowUp, IconArrowUpDown, IconWarning} from '../components/Icons'
import {Labels, labelsString, MetricName, parseLabels} from '../labels'
import {createConnectTransport, createPromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connectweb'
import {
  Alert as ObjectiveAlert,
  Alert_State,
  GetAlertsResponse,
  GetStatusResponse,
  ListResponse,
  Objective,
  ObjectiveStatus,
} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration} from '../duration'

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

enum TableSortType {
  Name,
  Window,
  Objective,
  Latency,
  Availability,
  Budget,
  Alerts,
}

enum TableSortOrder {
  Ascending,
  Descending,
}

interface TableSorting {
  type: TableSortType
  order: TableSortOrder
}

const List = () => {
  const client = useMemo(() => {
    const baseUrl = API_BASEPATH === undefined ? 'http://localhost:9099' : API_BASEPATH
    return createPromiseClient(ObjectiveService, createConnectTransport({baseUrl}))
  }, [])

  const navigate = useNavigate()
  const {search} = useLocation()

  const [objectives, setObjectives] = useState<Objective[]>([])
  const initialTableState: TableState = {objectives: {}}
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)
  const [tableSortState, setTableSortState] = useState<TableSorting>({
    type: TableSortType.Budget,
    order: TableSortOrder.Ascending,
  })

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

  const updateFilter = (lset: Labels) => {
    // Copy existing filterLabels (from router) and add/overwrite k-v-pairs
    const updatedFilter: Labels = {...filterLabels}
    for (const l in lset) {
      updatedFilter[l] = lset[l]
    }
    navigate(`?filter=${encodeURI(labelsString(updatedFilter))}`)
  }

  const removeFilterLabel = (k: string) => {
    const updatedFilter: Labels = {}
    for (const name in filterLabels) {
      if (name !== k) {
        updatedFilter[name] = filterLabels[name]
      }
    }

    if (Object.keys(updatedFilter).length === 0) {
      navigate(`?`)
      return
    }

    navigate(`?filter=${encodeURI(labelsString(updatedFilter))}`)
  }

  useEffect(() => {
    document.title = 'Objectives - Pyrra'

    client
      .list({expr: labelsString(filterLabels)})
      .then((resp: ListResponse) => setObjectives(resp.objectives))
      .catch((err) => console.log(err))
  }, [client, filterLabels])

  useEffect(() => {
    // const controller = new AbortController()
    // const signal = controller.signal // TODO: Pass this to the generated fetch client?

    // First we need to make sure to have objectives before we try fetching alerts for them.
    // If we were to do this concurrently it may end up in a situation
    // where we try to add an alert for a not yet existing objective.
    // TODO: This is prone to a concurrency race with updates of status that have additional groupings...
    // One solution would be to store this in a separate array and reconcile against that array after every status update.
    if (objectives.length > 0) {
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
    }

    objectives
      .sort((a: Objective, b: Objective) =>
        labelsString(a.labels).localeCompare(labelsString(b.labels)),
      )
      .forEach((o: Objective) => {
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

    // return () => {
    //   // cancel pending requests if necessary
    //   controller.abort()
    // }
  }, [client, objectives])

  const handleTableSort = (e: any, type: TableSortType): void => {
    e.preventDefault()

    if (tableSortState.type === type) {
      const order =
        tableSortState.order === TableSortOrder.Ascending
          ? TableSortOrder.Descending
          : TableSortOrder.Ascending
      setTableSortState({type: type, order: order})
    } else {
      setTableSortState({type: type, order: TableSortOrder.Ascending})
    }
  }

  // Indicates whether a latency column is needed or not.
  let tableLatency = false

  const tableList = Object.keys(table.objectives)
    .map((k: string) => table.objectives[k])
    .filter((o: TableObjective) => {
      const labels = {...o.objective.labels, ...o.groupingLabels}
      for (const k in filterLabels) {
        // if label doesn't exist by key or if values differ filter out.
        if (labels[k] === undefined || labels[k] !== filterLabels[k]) {
          return false
        }
      }
      return true
    })
    .sort((a: TableObjective, b: TableObjective) => {
      if (a.latency !== undefined || b.latency !== undefined) {
        tableLatency = true
      }

      // TODO: Make higher order function returning the sort function itself.
      switch (tableSortState.type) {
        case TableSortType.Name:
          if (tableSortState.order === TableSortOrder.Ascending) {
            return a.lset.localeCompare(b.lset)
          } else {
            return b.lset.localeCompare(a.lset)
          }
        case TableSortType.Window:
          if (tableSortState.order === TableSortOrder.Ascending) {
            return Number(a.objective.window?.seconds) - Number(b.objective.window?.seconds)
          } else {
            return Number(b.objective.window?.seconds) - Number(a.objective.window?.seconds)
          }
        case TableSortType.Objective:
          if (tableSortState.order === TableSortOrder.Ascending) {
            return a.objective.target - b.objective.target
          } else {
            return b.objective.target - a.objective.target
          }
        case TableSortType.Latency:
          if (a.latency === undefined && b.latency !== undefined) {
            return 1
          }
          if (a.latency !== undefined && b.latency === undefined) {
            return -1
          }
          if (a.latency !== undefined && b.latency !== undefined) {
            if (tableSortState.order === TableSortOrder.Ascending) {
              return a.latency - b.latency
            } else {
              return b.latency - a.latency
            }
          }

          return 0
        case TableSortType.Availability:
          if (a.availability == null && b.availability != null) {
            return 1
          }
          if (a.availability != null && b.availability == null) {
            return -1
          }
          if (
            a.availability !== undefined &&
            a.availability != null &&
            b.availability !== undefined &&
            b.availability != null
          ) {
            if (tableSortState.order === TableSortOrder.Ascending) {
              return a.availability.percentage - b.availability.percentage
            } else {
              return b.availability.percentage - a.availability.percentage
            }
          } else {
            return 0
          }
        case TableSortType.Budget:
          if (a.budget == null && b.budget != null) {
            return 1
          }
          if (a.budget != null && b.budget == null) {
            return -1
          }
          if (
            a.budget !== undefined &&
            a.budget != null &&
            b.budget !== undefined &&
            b.budget != null
          ) {
            if (tableSortState.order === TableSortOrder.Ascending) {
              return a.budget - b.budget
            } else {
              return b.budget - a.budget
            }
          } else {
            return 0
          }
        case TableSortType.Alerts:
          if (a.severity === null && b.severity === null) {
            return 0
          }
          if (a.severity === null && b.severity !== null) {
            return 1
          }
          if (a.severity !== null && b.severity === null) {
            return -1
          }
          if (tableSortState.order === TableSortOrder.Ascending) {
            if (a.severity === 'critical' && b.severity === 'warning') {
              return -1
            } else {
              return 1
            }
          } else {
            if (a.severity === 'critical' && b.severity === 'warning') {
              return 1
            } else {
              return 1
            }
          }
      }
      return 0
    })

  const upDownIcon =
    tableSortState.order === TableSortOrder.Ascending ? <IconArrowUp /> : <IconArrowDown />

  const objectivePage = (labels: Labels, grouping: Labels) => {
    return `/objectives?expr=${encodeURI(labelsString(labels))}&grouping=${encodeURI(
      labelsString(grouping),
    )}`
  }

  const renderLatency = (o: TableObjective) => {
    switch (hasObjectiveType(o.objective)) {
      case ObjectiveType.Ratio:
        return <></>
      case ObjectiveType.BoolGauge:
        return <></>
      case ObjectiveType.Latency:
      case ObjectiveType.LatencyNative:
        return renderLatencyTarget(o.objective)
    }
  }

  const renderAvailability = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner
            animation={'border'}
            style={{width: 20, height: 20, borderWidth: 2, opacity: 0.1}}
          />
        )
      case TableObjectiveState.NoData:
        return <>No data</>
      case TableObjectiveState.Error:
        return <span className="error">Error</span>
      case TableObjectiveState.Success: {
        if (o.availability === null || o.availability === undefined) {
          return <></>
        }

        const volumeWarning = (1 - o.objective.target) * o.availability.total

        const ls = labelsString(Object.assign({}, o.objective.labels, o.groupingLabels))
        return (
          <>
            <OverlayTrigger
              key={ls}
              overlay={
                <OverlayTooltip id={`tooltip-${ls}`}>
                  Errors: {Math.floor(o.availability.errors).toLocaleString()}
                  <br />
                  Total: {Math.floor(o.availability.total).toLocaleString()}
                </OverlayTooltip>
              }>
              <span className={o.availability.percentage > o.objective.target ? 'good' : 'bad'}>
                {(100 * o.availability.percentage).toFixed(2)}%
              </span>
            </OverlayTrigger>
            {volumeWarning < 1 ? (
              <>
                <OverlayTrigger
                  key={`${ls}-warning`}
                  overlay={
                    <OverlayTooltip id={`tooltip-${ls}-warning`}>
                      Too few requests!
                      <br />
                      Adjust your objective or wait for events.
                    </OverlayTooltip>
                  }>
                  <span className="volume-warning">
                    <IconWarning width={20} height={20} fill="#b10d0d" />
                  </span>
                </OverlayTrigger>
              </>
            ) : (
              <></>
            )}
          </>
        )
      }
    }
  }

  const renderErrorBudget = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner
            animation={'border'}
            style={{width: 20, height: 20, borderWidth: 2, opacity: 0.1}}
          />
        )
      case TableObjectiveState.NoData:
        return <>No data</>
      case TableObjectiveState.Error:
        return <span className="error">Error</span>
      case TableObjectiveState.Success:
        if (o.budget === null || o.budget === undefined) {
          return <></>
        }
        return (
          <span className={o.budget >= 0 ? 'good' : 'bad'}>{(100 * o.budget).toFixed(2)}%</span>
        )
    }
  }

  return (
    <>
      <Navbar />
      <Container className="content list">
        <Row>
          <Col>
            <h3>Objectives</h3>
          </Col>
        </Row>
        <Row>
          <Col>
            {Object.keys(filterLabels).map((k: string) => (
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
        </Row>
        <Row>
          <div className="table-responsive">
            <Table hover={true}>
              <thead>
                <tr>
                  <th
                    className={tableSortState.type === TableSortType.Name ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Name)}>
                    <a>Name </a>
                    {tableSortState.type === TableSortType.Name ? upDownIcon : <IconArrowUpDown />}
                  </th>
                  <th
                    className={tableSortState.type === TableSortType.Window ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Window)}>
                    <a>Time Window </a>
                    {tableSortState.type === TableSortType.Window ? (
                      upDownIcon
                    ) : (
                      <IconArrowUpDown />
                    )}
                  </th>
                  <th
                    className={tableSortState.type === TableSortType.Objective ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Objective)}>
                    <a>Objective </a>
                    {tableSortState.type === TableSortType.Objective ? (
                      upDownIcon
                    ) : (
                      <IconArrowUpDown />
                    )}
                  </th>
                  {tableLatency ? (
                    <th
                      className={tableSortState.type === TableSortType.Latency ? 'active' : ''}
                      onClick={(e) => handleTableSort(e, TableSortType.Latency)}>
                      <a>Latency </a>
                      {tableSortState.type === TableSortType.Latency ? (
                        upDownIcon
                      ) : (
                        <IconArrowUpDown />
                      )}
                    </th>
                  ) : (
                    <></>
                  )}
                  <th
                    className={tableSortState.type === TableSortType.Availability ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Availability)}>
                    <a>Availability </a>
                    {tableSortState.type === TableSortType.Availability ? (
                      upDownIcon
                    ) : (
                      <IconArrowUpDown />
                    )}
                  </th>
                  <th
                    className={tableSortState.type === TableSortType.Budget ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Budget)}>
                    <a>Error Budget </a>
                    {tableSortState.type === TableSortType.Budget ? (
                      upDownIcon
                    ) : (
                      <IconArrowUpDown />
                    )}
                  </th>
                  <th
                    className={tableSortState.type === TableSortType.Alerts ? 'active' : ''}
                    onClick={(e) => handleTableSort(e, TableSortType.Alerts)}>
                    <a>Alerts </a>
                    {tableSortState.type === TableSortType.Alerts ? (
                      upDownIcon
                    ) : (
                      <IconArrowUpDown />
                    )}
                  </th>
                </tr>
              </thead>
              <tbody>
                {tableList.map((o: TableObjective) => {
                  const name = o.objective.labels[MetricName]
                  const labelBadges = Object.entries({...o.objective.labels, ...o.groupingLabels})
                    .filter((l: [string, string]) => l[0] !== MetricName)
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
                          updateFilter(lset)
                        }}>
                        <a>
                          {l[0]}={l[1]}
                        </a>
                      </Badge>
                    ))

                  const classes =
                    o.severity !== null
                      ? ['table-row-clickable', 'firing']
                      : ['table-row-clickable']

                  return (
                    <tr
                      key={o.lset}
                      className={classes.join(' ')}
                      onClick={() => {
                        navigate(objectivePage(o.objective.labels, o.groupingLabels))
                      }}>
                      <td>
                        <Link
                          to={objectivePage(o.objective.labels, o.groupingLabels)}
                          className="text-reset"
                          style={{marginRight: 5}}>
                          {name}
                        </Link>
                        {labelBadges}
                      </td>
                      <td>{formatDuration(Number(o.objective.window?.seconds) * 1000)}</td>
                      <td>{(100 * o.objective.target).toFixed(2)}%</td>
                      {tableLatency ? <td>{renderLatency(o)}</td> : <></>}
                      <td>{renderAvailability(o)}</td>
                      <td>{renderErrorBudget(o)}</td>
                      <td>
                        <span className="severity">{o.severity !== null ? o.severity : ''}</span>
                      </td>
                    </tr>
                  )
                })}
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

export default List
