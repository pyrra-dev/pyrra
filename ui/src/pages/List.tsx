import React, { useEffect, useMemo, useReducer, useState } from 'react'
import { Badge, Col, Container, OverlayTrigger, Row, Spinner, Table, Tooltip as OverlayTooltip } from 'react-bootstrap'
import { Configuration, Objective, ObjectivesApi, ObjectiveStatus } from '../client'
import { formatDuration, PATH_PREFIX } from '../App'
import { Link, useHistory } from 'react-router-dom'
import Navbar from '../components/Navbar'
import { IconArrowDown, IconArrowUp, IconArrowUpDown } from '../components/Icons'
import { labelsString } from "../labels";

enum TableObjectiveState {
  Unknown,
  Success,
  NoData,
  Error,
}

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective extends Objective {
  lset: string
  groupingLabels: { [key: string]: string; };
  state: TableObjectiveState
  availability?: TableAvailability | null
  budget?: number | null
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
}

type TableAction =
  | { type: TableActionType.SetObjective, lset: string, objective: Objective }
  | { type: TableActionType.DeleteObjective, lset: string }
  | { type: TableActionType.SetStatus, lset: string, status: ObjectiveStatus }
  | { type: TableActionType.SetObjectiveWithStatus, lset: string, statusLabels: { [key: string]: string }, objective: Objective, status: ObjectiveStatus }
  | { type: TableActionType.SetStatusNone, lset: string }
  | { type: TableActionType.SetStatusError, lset: string }

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
            availability: undefined,
            budget: undefined
          }
        }
      }
    case TableActionType.DeleteObjective:
      const { [action.lset]: value, ...cleanedObjective } = state.objectives
      return {
        objectives: { ...cleanedObjective }
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
    case TableActionType.SetObjectiveWithStatus:
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
            availability: {
              errors: action.status.availability.errors,
              total: action.status.availability.total,
              percentage: action.status.availability.percentage
            },
            budget: action.status.budget?.remaining
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
}

enum TableSortOrder {Ascending, Descending}

interface TableSorting {
  type: TableSortType
  order: TableSortOrder
}

const List = () => {
  const api = useMemo(() => {
    return new ObjectivesApi(new Configuration({ basePath: `./api/v1` }))
  }, [])

  const history = useHistory()
  const [objectives, setObjectives] = useState<Array<Objective>>([])
  const initialTableState: TableState = { objectives: {} }
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)
  const [tableSortState, setTableSortState] = useState<TableSorting>({
    type: TableSortType.Budget,
    order: TableSortOrder.Ascending
  })

  useEffect(() => {
    document.title = 'Objectives - Pyrra'

    api.listObjectives({ expr: '' })
      .then((objectives: Objective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [api])

  useEffect(() => {
    // const controller = new AbortController()
    // const signal = controller.signal // TODO: Pass this to the generated fetch client?

    objectives
      .sort((a: Objective, b: Objective) => labelsString(a.labels).localeCompare(labelsString(b.labels)))
      .forEach((o: Objective) => {
        dispatchTable({ type: TableActionType.SetObjective, lset: labelsString(o.labels), objective: o })

        api.getObjectiveStatus({ expr: labelsString(o.labels) })
          .then((s: ObjectiveStatus[]) => {
            if (s.length === 0) {
              dispatchTable({ type: TableActionType.SetStatusNone, lset: labelsString(o.labels) })
            } else if (s.length === 1) {
              dispatchTable({ type: TableActionType.SetStatus, lset: labelsString(o.labels), status: s[0] })
            } else {
              dispatchTable({ type: TableActionType.DeleteObjective, lset: labelsString(o.labels) })

              s.forEach((s: ObjectiveStatus) => {
                // Copy the objective
                const so = { ...o }
                // Identify by the combined labels
                const sLabels = s.labels !== undefined ? s.labels : {}
                const soLabels = { ...o.labels, ...sLabels }

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
            dispatchTable({ type: TableActionType.SetStatusError, lset: labelsString(o.labels) })
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
      setTableSortState({ type: type, order: order })
    } else {
      setTableSortState({ type: type, order: TableSortOrder.Ascending })
    }
  }

  const tableList = Object.keys(table.objectives)
    .map((k: string) => table.objectives[k])
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
        }
        return 0
      }
    )

  const upDownIcon = tableSortState.order === TableSortOrder.Ascending ? <IconArrowUp/> : <IconArrowDown/>

  const objectivePage = (
    labels: { [key: string]: string },
    grouping: { [key: string]: string }
  ) => {
    return `/objectives?expr=${labelsString(labels)}&grouping=${labelsString(grouping)}`
  }

  const handleTableRowClick = (
    labels: { [key: string]: string },
    grouping: { [key: string]: string }
  ) => () => {
    history.push(objectivePage(labels, grouping))
  }

  const renderAvailability = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>
        )
      case TableObjectiveState.NoData:
        return <>No data</>
      case TableObjectiveState.Error:
        return <span className="error">Error</span>
      case TableObjectiveState.Success:
        if (o.availability === null || o.availability === undefined) {
          return <></>
        }
        return (
          <OverlayTrigger
            key={labelsString(o.labels)}
            overlay={
              <OverlayTooltip id={`tooltip-${labelsString(o.labels)}`}>
                Errors: {Math.floor(o.availability.errors).toLocaleString()}<br/>
                Total: {Math.floor(o.availability.total).toLocaleString()}
              </OverlayTooltip>
            }>
          <span className={o.availability.percentage > o.target ? 'good' : 'bad'}>
            {(100 * o.availability.percentage).toFixed(2)}%
          </span>
          </OverlayTrigger>
        )
    }
  }

  const renderErrorBudget = (o: TableObjective) => {
    switch (o.state) {
      case TableObjectiveState.Unknown:
        return (
          <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>
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
          <div className="table-responsive">
            <Table hover={true} striped={false}>
              <thead>
              <tr>
                <th
                  className={tableSortState.type === TableSortType.Name ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Name)}
                >Name {tableSortState.type === TableSortType.Name ? upDownIcon : <IconArrowUpDown/>}</th>
                <th
                  className={tableSortState.type === TableSortType.Window ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Window)}
                >Time Window {tableSortState.type === TableSortType.Window ? upDownIcon : <IconArrowUpDown/>}</th>
                <th
                  className={tableSortState.type === TableSortType.Objective ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Objective)}
                >Objective {tableSortState.type === TableSortType.Objective ? upDownIcon : <IconArrowUpDown/>}</th>
                <th
                  className={tableSortState.type === TableSortType.Availability ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Availability)}
                >Availability {tableSortState.type === TableSortType.Availability ? upDownIcon :
                  <IconArrowUpDown/>}</th>
                <th
                  className={tableSortState.type === TableSortType.Budget ? 'active' : ''}
                  onClick={() => handleTableSort(TableSortType.Budget)}
                >Error Budget {tableSortState.type === TableSortType.Budget ? upDownIcon : <IconArrowUpDown/>}</th>
              </tr>
              </thead>
              <tbody>
              {tableList.map((o: TableObjective, i: number) => {
                const name = o.labels['__name__']
                const labelBadges = Object.entries({ ...o.labels, ...o.groupingLabels })
                  .filter((l: [string, string]) => l[0] !== '__name__')
                  .map((l: [string, string]) => (
                    <Badge variant={"light"}>{l[0]}={l[1]}</Badge>
                  ))

                return (
                  <tr key={i} className="table-row-clickable" onClick={handleTableRowClick(o.labels, o.groupingLabels)}>
                    <td>
                      <Link to={objectivePage(o.labels, o.groupingLabels)} className="text-reset" style={{ marginRight: 5 }}>
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
