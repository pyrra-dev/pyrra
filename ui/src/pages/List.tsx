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
  Tooltip as OverlayTooltip
} from 'react-bootstrap'
import {
  Configuration,
  MultiBurnrateAlert,
  MultiBurnrateAlertStateEnum,
  Objective,
  ObjectivesApi,
  ObjectiveStatus
} from '../client'
import {API_BASEPATH, formatDuration} from '../App'
import {Link, useLocation, useNavigate} from 'react-router-dom'
import Navbar from '../components/Navbar'
import {IconArrowDown, IconArrowUp, IconArrowUpDown, IconWarning} from '../components/Icons'
import {Labels, labelsString, MetricName, parseLabels} from "../labels";

enum TableObjectiveState {
  Unknown,
  Success,
  NoData,
  Error,
}

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective extends Objective {
  lset: string
  groupingLabels: Labels
  state: TableObjectiveState
  availability?: TableAvailability | null
  budget?: number | null
  severity: string | null
}

interface TableAvailability {
  errors: number
  total: number
  percentage: number
}

interface TableState {
  objectives: { [key: string]: TableObjective }
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
  | { type: TableActionType.SetObjective, lset: string, objective: Objective }
  | { type: TableActionType.DeleteObjective, lset: string }
  | { type: TableActionType.SetStatus, lset: string, status: ObjectiveStatus }
  | { type: TableActionType.SetObjectiveWithStatus, lset: string, statusLabels: Labels, objective: Objective, status: ObjectiveStatus }
  | { type: TableActionType.SetStatusNone, lset: string }
  | { type: TableActionType.SetStatusError, lset: string }
  | { type: TableActionType.SetAlert, labels: Labels, severity: string }

const tableReducer = (state: TableState, action: TableAction): TableState => {
  switch (action.type) {
    case TableActionType.SetObjective:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            lset: action.lset,
            labels: action.objective.labels,
            groupingLabels: {},
            description: action.objective.description,
            window: action.objective.window,
            target: action.objective.target,
            config: action.objective.config,
            state: TableObjectiveState.Unknown,
            severity: null,
            availability: undefined,
            budget: undefined
          }
        }
      }
    case TableActionType.DeleteObjective: {
      const {[action.lset]: _, ...cleanedObjective} = state.objectives
      return {
        objectives: {...cleanedObjective}
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
              errors: action.status.availability.errors,
              total: action.status.availability.total,
              percentage: action.status.availability.percentage
            },
            budget: action.status.budget?.remaining
          }
        }
      }
    case TableActionType.SetObjectiveWithStatus: {
      // It is possible that we may need to merge some previous state
      let severity: string | null = null;
      const o = state.objectives[action.lset]
      if (o !== undefined) {
        severity = o.severity
      }

      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            lset: action.lset,
            labels: action.objective.labels,
            groupingLabels: action.statusLabels,
            description: action.objective.description,
            window: action.objective.window,
            target: action.objective.target,
            config: action.objective.config,
            state: TableObjectiveState.Success,
            severity: severity,
            availability: {
              errors: action.status.availability.errors,
              total: action.status.availability.total,
              percentage: action.status.availability.percentage
            },
            budget: action.status.budget?.remaining
          }
        }
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
            budget: null
          }
        }
      }
    case TableActionType.SetStatusError:
      return {
        objectives: {
          ...state.objectives,
          [action.lset]: {
            ...state.objectives[action.lset],
            state: TableObjectiveState.Error,
            availability: null,
            budget: null
          }
        }
      }
    case TableActionType.SetAlert: {
      // Find the objective this alert's labels is the super set for.
      const result = Object.entries(state.objectives).find(([, o]) => {
        const allLabels: Labels = {...o.labels, ...o.groupingLabels}

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
            severity: action.severity
          }
        }
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
  Availability,
  Budget,
  Alerts,
}

enum TableSortOrder {Ascending, Descending}

interface TableSorting {
  type: TableSortType
  order: TableSortOrder
}

const List = () => {
  const api = useMemo(() => {
    return new ObjectivesApi(new Configuration({basePath: API_BASEPATH}))
  }, [])

  const navigate = useNavigate()
  const {search} = useLocation()

  const [objectives, setObjectives] = useState<Objective[]>([])
  const initialTableState: TableState = {objectives: {}}
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)
  const [tableSortState, setTableSortState] = useState<TableSorting>({
    type: TableSortType.Budget,
    order: TableSortOrder.Ascending
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

    api.listObjectives({ expr: labelsString(filterLabels) })
      .then((objectives: Objective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [api, filterLabels])

  useEffect(() => {
    // const controller = new AbortController()
    // const signal = controller.signal // TODO: Pass this to the generated fetch client?

    // First we need to make sure to have objectives before we try fetching alerts for them.
    // If we were to do this concurrently it may end up in a situation
    // where we try to add an alert for a not yet existing objective.
    // TODO: This is prone to a concurrency race with updates of status that have additional groupings...
    // One solution would be to store this in a separate array and reconcile against that array after every status update.
    if (objectives.length > 0) {
      api.getMultiBurnrateAlerts({expr: '', inactive: false, current: false})
        .then((alerts: MultiBurnrateAlert[]) => {
          alerts.forEach((alert: MultiBurnrateAlert) => {
            if (alert.state === MultiBurnrateAlertStateEnum.Firing) {
              dispatchTable({type: TableActionType.SetAlert, labels: alert.labels, severity: alert.severity})
            }
          })
        })
        .catch((err) => console.log(err))
    }

    objectives
      .sort((a: Objective, b: Objective) => labelsString(a.labels).localeCompare(labelsString(b.labels)))
      .forEach((o: Objective) => {
        dispatchTable({type: TableActionType.SetObjective, lset: labelsString(o.labels), objective: o})

        api.getObjectiveStatus({expr: labelsString(o.labels)})
          .then((s: ObjectiveStatus[]) => {
            if (s.length === 0) {
              dispatchTable({type: TableActionType.SetStatusNone, lset: labelsString(o.labels)})
            } else if (s.length === 1) {
              dispatchTable({type: TableActionType.SetStatus, lset: labelsString(o.labels), status: s[0]})
            } else {
              dispatchTable({type: TableActionType.DeleteObjective, lset: labelsString(o.labels)})

              s.forEach((s: ObjectiveStatus) => {
                // Copy the objective
                const so = {...o}
                // Identify by the combined labels
                const sLabels: Labels = s.labels !== undefined ? s.labels : {}
                const soLabels: Labels = {...o.labels, ...sLabels}

                dispatchTable({
                  type: TableActionType.SetObjectiveWithStatus,
                  lset: labelsString(soLabels),
                  statusLabels: sLabels,
                  objective: so,
                  status: s
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
  }, [api, objectives])

  const handleTableSort = (type: TableSortType): void => {
    if (tableSortState.type === type) {
      const order = tableSortState.order === TableSortOrder.Ascending ? TableSortOrder.Descending : TableSortOrder.Ascending
      setTableSortState({type: type, order: order})
    } else {
      setTableSortState({type: type, order: TableSortOrder.Ascending})
    }
  }

  const tableList = Object.keys(table.objectives)
    .map((k: string) => table.objectives[k])
    .filter((o: TableObjective) => {
      const labels = { ...o.labels, ...o.groupingLabels }
      for (const k in filterLabels) {
        // if label doesn't exist by key or if values differ filter out.
        if (labels[k] === undefined || labels[k] !== filterLabels[k]) {
          return false
        }
      }
      return true
    })
    .sort((a: TableObjective, b: TableObjective) => {
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
              return a.window - b.window
            } else {
              return b.window - a.window
            }
          case TableSortType.Objective:
            if (tableSortState.order === TableSortOrder.Ascending) {
              return a.target - b.target
            } else {
              return b.target - a.target
            }
          case TableSortType.Availability:
            if (a.availability == null && b.availability != null) {
              return 1
            }
            if (a.availability != null && b.availability == null) {
              return -1
            }
            if (a.availability !== undefined && a.availability != null && b.availability !== undefined && b.availability != null) {
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
            if (a.budget !== undefined && a.budget != null && b.budget !== undefined && b.budget != null) {
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
      }
    )

  const upDownIcon = tableSortState.order === TableSortOrder.Ascending ? <IconArrowUp/> : <IconArrowDown/>

  const objectivePage = (
    labels: Labels,
    grouping: Labels,
  ) => {
    return `/objectives?expr=${encodeURI(labelsString(labels))}&grouping=${encodeURI(labelsString(grouping))}`
  }

  const renderAvailability = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner animation={'border'} style={{width: 20, height: 20, borderWidth: 2, opacity: 0.1}}/>
        )
      case TableObjectiveState.NoData:
        return <>No data</>
      case TableObjectiveState.Error:
        return <span className="error">Error</span>
      case TableObjectiveState.Success: {
        if (o.availability === null || o.availability === undefined) {
          return <></>
        }

        const volumeWarning = ((1 - o.target) * o.availability.total)

        const ls = labelsString(Object.assign({}, o.labels, o.groupingLabels))
        return (
          <>
            <OverlayTrigger
              key={ls}
              overlay={
                <OverlayTooltip id={`tooltip-${ls}`}>
                  Errors: {Math.floor(o.availability.errors).toLocaleString()}<br/>
                  Total: {Math.floor(o.availability.total).toLocaleString()}
                </OverlayTooltip>
              }>
              <span className={o.availability.percentage > o.target ? 'good' : 'bad'}>
                {(100 * o.availability.percentage).toFixed(2)}%
              </span>
            </OverlayTrigger>
            {volumeWarning < 1 ? <>
              <OverlayTrigger
                key={`${ls}-warning`}
                overlay={
                  <OverlayTooltip id={`tooltip-${ls}-warning`}>
                    Too few requests!<br/>Adjust your objective or wait for events.
                  </OverlayTooltip>
                }>
                <span className="volume-warning">
                  <IconWarning width={20} height={20} fill="#b10d0d"/>
                </span>
              </OverlayTrigger>
            </> : <></>}
          </>
        )
      }
    }
  }

  const renderErrorBudget = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner animation={'border'} style={{width: 20, height: 20, borderWidth: 2, opacity: 0.1}}/>
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
          <span className={o.budget >= 0 ? 'good' : 'bad'}>
            {(100 * o.budget).toFixed(2)}%
          </span>
        )
    }
  }

  return (
    <>
      <Navbar/>
      <Container className="content list">
        <Row>
          <Col>
            <h3>Objectives</h3>
          </Col>
        </Row>
        <Row>
          <Col>
            {Object.keys(filterLabels).map((k: string) => (
              <Button variant="light" size="sm" className='filter-close' onClick={() => removeFilterLabel(k)}>
                {`${k}=${filterLabels[k]}`}
                <span className='btn-close'></span>
              </Button>
            ))}
            <Alert show={filterError} variant="danger">Your SLO filter is broken. Please reset the filter.</Alert>
          </Col>
        </Row>
        <Row>
          <div className="table-responsive">
            <Table hover={true}>
              <thead>
              <tr>
                <th
                  className={tableSortState.type === TableSortType.Name ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Name)}>
                  Name {tableSortState.type === TableSortType.Name ? upDownIcon : <IconArrowUpDown/>}
                </th>
                <th
                  className={tableSortState.type === TableSortType.Window ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Window)}>
                  Time Window {tableSortState.type === TableSortType.Window ? upDownIcon : <IconArrowUpDown/>}
                </th>
                <th
                  className={tableSortState.type === TableSortType.Objective ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Objective)}>
                  Objective {tableSortState.type === TableSortType.Objective ? upDownIcon : <IconArrowUpDown/>}
                </th>
                <th
                  className={tableSortState.type === TableSortType.Availability ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Availability)}>
                  Availability {tableSortState.type === TableSortType.Availability ? upDownIcon : <IconArrowUpDown/>}
                </th>
                <th
                  className={tableSortState.type === TableSortType.Budget ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Budget)}>
                  Error Budget {tableSortState.type === TableSortType.Budget ? upDownIcon : <IconArrowUpDown/>}
                </th>
                <th
                  className={tableSortState.type === TableSortType.Alerts ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Alerts)}>
                  Alerts {tableSortState.type === TableSortType.Alerts ? upDownIcon : <IconArrowUpDown/>}
                </th>
              </tr>
              </thead>
              <tbody>
              {tableList.map((o: TableObjective) => {
                const name = o.labels[MetricName]
                const labelBadges = Object.entries({...o.labels, ...o.groupingLabels})
                  .filter((l: [string, string]) => l[0] !== MetricName)
                  .map((l: [string, string]) => (
                    <Badge key={l[0]} bg="light" text="dark" className="fw-normal"
                           style={{ marginRight: 5 }}
                           onClick={(event) => {
                             event.stopPropagation()

                             const lset: Labels = {}
                             lset[l[0]] = l[1]
                             updateFilter(lset)
                           }}>
                      {l[0]}={l[1]}
                    </Badge>
                  ))

                const classes = o.severity !== null ? ['table-row-clickable', 'firing'] : ['table-row-clickable']

                return (
                  <tr key={o.lset} className={classes.join(' ')} onClick={() => {
                    navigate(objectivePage(o.labels, o.groupingLabels))
                  }}>
                    <td>
                      <Link to={objectivePage(o.labels, o.groupingLabels)} className="text-reset"
                            style={{marginRight: 5}}>
                        {name}
                      </Link>
                      {labelBadges}
                    </td>
                    <td>{formatDuration(o.window)}</td>
                    <td>
                      {(100 * o.target).toFixed(2)}%
                    </td>
                    <td>
                      {renderAvailability(o)}
                    </td>
                    <td>
                      {renderErrorBudget(o)}
                    </td>
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
            <small>All availabilities and error budgets are calculated across the entire time window of the
              objective.</small>
          </Col>
        </Row>
      </Container>
    </>
  )
}

export default List
