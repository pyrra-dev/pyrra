import React, { useEffect, useReducer, useState } from 'react'
import { Col, Container, OverlayTrigger, Row, Spinner, Table, Tooltip as OverlayTooltip } from 'react-bootstrap'
import { Configuration, Objective, ObjectivesApi, ObjectiveStatus } from '../client'
import { formatDuration, PUBLIC_API } from '../App'
import { Link, useHistory } from 'react-router-dom'

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective extends Objective {
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

enum TableActionType { SetObjective, SetStatus, SetStatusNone }

type TableAction =
  | { type: TableActionType.SetObjective, objective: Objective }
  | { type: TableActionType.SetStatus, name: string, status: ObjectiveStatus }
  | { type: TableActionType.SetStatusNone, name: string }

const tableReducer = (state: TableState, action: TableAction): TableState => {
  switch (action.type) {
    case TableActionType.SetObjective:
      return {
        objectives: {
          ...state.objectives,
          [action.objective.name]: {
            name: action.objective.name,
            namespace: action.objective.namespace,
            description: action.objective.description,
            window: action.objective.window,
            target: action.objective.target,
            config: action.objective.config,
            availability: undefined,
            budget: undefined
          }
        }
      }
    case TableActionType.SetStatus:
      return {
        objectives: {
          ...state.objectives,
          [action.name]: {
            ...state.objectives[action.name],
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
          [action.name]: {
            ...state.objectives[action.name],
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
  const APIConfiguration = new Configuration({ basePath: `${PUBLIC_API}api/v1` })
  const APIObjectives = new ObjectivesApi(APIConfiguration)

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

    APIObjectives.listObjectives()
      .then((objectives: Objective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [])

  useEffect(() => {
    // const controller = new AbortController()
    // const signal = controller.signal // TODO: Pass this to the generated fetch client?

    objectives
      .sort((a: Objective, b: Objective) => a.name.localeCompare(b.name))
      .forEach((o: Objective) => {
        dispatchTable({ type: TableActionType.SetObjective, objective: o })

        APIObjectives.getObjectiveStatus({ namespace: o.namespace, name: o.name })
          .then((s: ObjectiveStatus) => {
            dispatchTable({ type: TableActionType.SetStatus, name: o.name, status: s })
          })
          .catch((err) => {
            if (err.status === 404) {
              dispatchTable({ type: TableActionType.SetStatusNone, name: o.name })
            }
          })
      })

    // return () => {
    //   // cancel pending requests if necessary
    //   controller.abort()
    // }
  }, [objectives])

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
              return a.name.localeCompare(b.name)
            } else {
              return b.name.localeCompare(a.name)
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

  const upDownIcon = tableSortState.order === TableSortOrder.Ascending ? '⬆️' : '⬇️'

  const objectivePage = (namespace: string, name: string) => `/objectives/${namespace}/${name}`

  const handleTableRowClick = (namespace: string, name: string) => () => {
    history.push(objectivePage(namespace, name))
  }

  return (
    <Container className="App">
      <Row>
        <Col>
          <h1>Objectives</h1>
        </Col>
      </Row>
      <Table hover={true} striped={false}>
        <thead>
        <tr>
          <th onClick={() => handleTableSort(TableSortType.Name)}>
            Name {tableSortState.type === TableSortType.Name ? upDownIcon : '↕️'}
          </th>
          <th onClick={() => handleTableSort(TableSortType.Window)}>
            Time Window {tableSortState.type === TableSortType.Window ? upDownIcon : '↕️'}
          </th>
          <th onClick={() => handleTableSort(TableSortType.Objective)}>
            Objective {tableSortState.type === TableSortType.Objective ? upDownIcon : '↕️'}
          </th>
          <th onClick={() => handleTableSort(TableSortType.Availability)}>
            Availability {tableSortState.type === TableSortType.Availability ? upDownIcon : '↕️'}
          </th>
          <th onClick={() => handleTableSort(TableSortType.Budget)}>
            Error Budget {tableSortState.type === TableSortType.Budget ? upDownIcon : '↕️'}
          </th>
        </tr>
        </thead>
        <tbody>
        {tableList.map((o: TableObjective) => (
          <tr key={o.name} className="table-row-clickable" onClick={handleTableRowClick(o.namespace, o.name)}>
            <td>
              <Link to={objectivePage(o.namespace, o.name)}>
                {o.name}
              </Link>
            </td>
            <td>{formatDuration(o.window)}</td>
            <td>
              {(100 * o.target).toFixed(2)}%
            </td>
            <td>
              {o.availability === undefined ? (
                <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>
              ) : <></>}
              {o.availability === null ? <>-</> : <></>}
              {o.availability !== undefined && o.availability != null ?
                <OverlayTrigger
                  key={o.name}
                  overlay={
                    <OverlayTooltip id={`tooltip-${o.name}`}>
                      Errors: {Math.floor(o.availability.errors).toLocaleString()}<br/>
                      Total: {Math.floor(o.availability.total).toLocaleString()}
                    </OverlayTooltip>
                  }>
                  <span className={o.availability.percentage > o.target ? 'good' : 'bad'}>
                    {(100 * o.availability.percentage).toFixed(2)}%
                  </span>
                </OverlayTrigger>
                : <></>}
            </td>
            <td>
              {o.budget === undefined ? (
                <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>
              ) : <></>}
              {o.budget === null ? <>-</> : <></>}
              {o.budget !== undefined && o.budget != null ?
                <span className={o.budget >= 0 ? 'good' : 'bad'}>
                  {(100 * o.budget).toFixed(2)}%
                </span> : <></>}
            </td>
          </tr>
        ))}
        </tbody>
      </Table>
      <Row>
        <Col>
          <small>All availabilities and error budgets are calculated across the entire time window of the
            objective.</small>
        </Col>
      </Row>
    </Container>
  )
}


export default List
