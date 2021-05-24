import React, { useEffect, useReducer, useRef, useState } from 'react';
import './App.css';
import { Col, Container, Row, Spinner, Table } from 'react-bootstrap';
import { BrowserRouter, Link, Route, RouteComponentProps, Switch, useHistory } from 'react-router-dom';

// @ts-ignore - this is passed from the HTML template.
const PUBLIC_API: string = window.PUBLIC_API;

interface Objective {
  name: string
  target: number
  window: string
}

interface Availability {
  percentage: number
  total: number
  errors: number
}

interface ErrorBudget {
  total: number
  max: number
  remaining: number
}

interface Valet {
  window: number
  volume: number
  errors: number
  availability: number
  budget: number
}

const App = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route exact={true} path="/" component={List}/>
        <Route path="/objectives/:name" component={Details}/>
      </Switch>
    </BrowserRouter>
  )
}

interface ObjectiveStatus {
  availability: number
  budget: number
}

const fetchObjectives = async (): Promise<Array<Objective>> => {
  const resp: Response = await fetch(`${PUBLIC_API}api/objectives`)
  const json = await resp.json()
  if (!resp.ok) {
    return Promise.reject(resp)
  }
  return json
}

const fetchObjectiveStatus = async (name: string, signal: AbortSignal): Promise<ObjectiveStatus> => {
  const resp: Response = await fetch(`${PUBLIC_API}api/objectives/${name}/status`, { signal })
  const json = await resp.json()
  if (!resp.ok) {
    return Promise.reject(resp)
  }
  return { availability: json.availability.percentage, budget: json.budget.remaining }
}

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective extends Objective {
  availability?: number
  budget?: number
}

interface TableState {
  objectives: { [key: string]: TableObjective }
}

enum TableActionType { SetObjective, SetStatus }

type TableAction =
  | { type: TableActionType.SetObjective, objective: Objective }
  | { type: TableActionType.SetStatus, name: string, status: ObjectiveStatus }

const tableReducer = (state: TableState, action: TableAction): TableState => {
  switch (action.type) {
    case TableActionType.SetObjective:
      return {
        objectives: {
          ...state.objectives,
          [action.objective.name]: {
            name: action.objective.name,
            window: action.objective.window,
            target: action.objective.target,
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
            availability: action.status.availability,
            budget: action.status.budget
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
  const [objectives, setObjectives] = useState<Array<Objective>>([])
  const initialTableState: TableState = { objectives: {} }
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)
  const [tableSortState, setTableSortState] = useState<TableSorting>({
    type: TableSortType.Budget,
    order: TableSortOrder.Ascending
  })

  useEffect(() => {
    fetchObjectives()
      .then((objectives: Objective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    const signal = controller.signal

    objectives
      .sort((a: Objective, b: Objective) => a.name.localeCompare(b.name))
      .forEach((o: Objective) => {
        dispatchTable({ type: TableActionType.SetObjective, objective: o })

        fetchObjectiveStatus(o.name, signal)
          .then((s: ObjectiveStatus) => {
            dispatchTable({ type: TableActionType.SetStatus, name: o.name, status: s })
          })
      })

    return () => {
      // cancel pending requests if necessary
      controller.abort()
    }
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
              return a.window.localeCompare(b.window)
            } else {
              return b.window.localeCompare(a.window)
            }
          case TableSortType.Objective:
            if (tableSortState.order === TableSortOrder.Ascending) {
              return a.target - b.target
            } else {
              return b.target - a.target
            }
          case TableSortType.Availability:
            if (a.availability !== undefined && b.availability !== undefined) {
              if (tableSortState.order === TableSortOrder.Ascending) {
                return a.availability - b.availability
              } else {
                return b.availability - a.availability
              }
            } else {
              return 1
            }
          case TableSortType.Budget:
            if (a.budget !== undefined && b.budget !== undefined) {
              if (tableSortState.order === TableSortOrder.Ascending) {
                return a.budget - b.budget
              } else {
                return b.budget - a.budget
              }
            } else {
              return 1
            }
        }
        return 0
      }
    )

  const upDownIcon = tableSortState.order === TableSortOrder.Ascending ? '⬆️' : '⬇️'

  const history = useHistory()

  const handleTableRowClick = (name: string) => () => {
    history.push(`/objectives/${name}`)
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
          <tr key={o.name} className="table-row-clickable" onClick={handleTableRowClick(o.name)}>
            <td>
              {o.name}
            </td>
            <td>{o.window}</td>
            <td>
              {(100 * o.target).toFixed(2)}%
            </td>
            <td>
              {o.availability !== undefined ?
                <span className={o.availability > o.target ? 'good' : 'bad'}>
                  {(100 * o.availability).toFixed(2)}%
                </span> :
                <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>}
            </td>
            <td>
              {o.budget !== undefined ?
                <span className={o.budget >= 0 ? 'good' : 'bad'}>
                  {(100 * o.budget).toFixed(2)}%
                </span> :
                <Spinner animation={'border'} style={{ width: 20, height: 20, borderWidth: 2, opacity: 0.1 }}/>}
            </td>
          </tr>
        ))}
        </tbody>
      </Table>
    </Container>
  )
}

interface DetailsRouteParams {
  name: string
}

const Details = (params: RouteComponentProps<DetailsRouteParams>) => {
  const name = params.match.params.name;
  const [objective, setObjective] = useState<Objective | null>(null);
  const [availability, setAvailability] = useState<Availability | null>(null);
  const [errorBudget, setErrorBudget] = useState<ErrorBudget | null>(null);
  const [valets, setValets] = useState<Array<Valet>>([]);

  const [errorBudgetSVGLoading, setErrorBudgetSVGLoading] = useState<boolean>(true)
  const errorBudgetSVG = useRef<SVGSVGElement>(null)

  useEffect(() => {
    const controller = new AbortController()

    fetch(`${PUBLIC_API}api/objectives/${name}`, { signal: controller.signal })
      .then((resp: Response) => resp.json())
      .then((data) => setObjective(data))

    fetch(`${PUBLIC_API}api/objectives/${name}/status`, { signal: controller.signal })
      .then((resp: Response) => resp.json())
      .then((data) => {
        setAvailability(data.availability)
        setErrorBudget(data.budget)
      })

    fetch(`${PUBLIC_API}api/objectives/${name}/valet`, { signal: controller.signal })
      .then((resp: Response) => resp.json())
      .then((data) => setValets(data))

    setErrorBudgetSVGLoading(true)

    fetch(`${PUBLIC_API}api/objectives/${name}/errorbudget.svg`, { signal: controller.signal })
      .then((resp: Response) => resp.text())
      .then((raw: string) => {
        const doc: Document = new DOMParser().parseFromString(raw, 'image/svg+xml')
        const svg: SVGSVGElement | null = doc.querySelector('svg')
        if (errorBudgetSVG.current != null && svg != null) {
          errorBudgetSVG.current.replaceWith(svg)
        }
      }).finally(() => setErrorBudgetSVGLoading(false))

    return () => {
      // cancel any pending requests.
      controller.abort()
    }
  }, [name])

  if (objective == null) {
    return (
      <div style={{ marginTop: '50px', textAlign: 'center' }}>
        <Spinner animation="border" role="status">
          <span className="sr-only">Loading...</span>
        </Spinner>
      </div>
    )
  }

  return (
    <div className="App">
      <Container>
        <Row style={{ marginBottom: '2em' }}>
          <Col>
            <Link to="/">⬅️ Overview</Link>
          </Col>
        </Row>
        <Row style={{ marginBottom: '5em' }}>
          <Col>
            <h1>{objective?.name}</h1>
          </Col>
        </Row>
        <Row>
          <Col className="metric">
            <div>
              <h2>{100 * objective.target}%</h2>
              <h6 className="text-muted">Objective in {objective.window}</h6>
            </div>
          </Col>
          <Col className="metric">
            <div>
              {availability != null ? (
                <h2 className={availability.percentage > 0 ? 'good' : 'bad'}>
                  {(100 * availability.percentage).toFixed(3)}%
                </h2>
              ) : (
                <h2>&nbsp;</h2>
              )
              }
              <h6 className="text-muted">Availability</h6>
            </div>
          </Col>
          <Col className="metric">
            <div>
              {errorBudget != null ? (
                <h2 className={errorBudget.remaining > 0 ? 'good' : 'bad'}>
                  {(100 * errorBudget.remaining).toFixed(3)}%
                </h2>
              ) : (
                <h2>&nbsp;</h2>
              )}
              <h6 className="text-muted">Error Budget</h6>
            </div>
          </Col>
        </Row>
        <br/>
        <Row>
          {errorBudgetSVGLoading ?
            <div style={{
              width: '100%',
              height: 230,
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center'
            }}>
              <Spinner animation="border" style={{ margin: '0 auto' }}/>
            </div>
            : <></>}
          <svg ref={errorBudgetSVG} style={{ display: errorBudgetSVGLoading ? 'none' : 'block' }}/>
          {valets.map((v: Valet) => (
            <Col style={{ textAlign: 'right' }}>
              <small>Volume: {Math.floor(v.volume).toLocaleString()}</small>&nbsp;
              <small>Errors: {Math.floor(v.errors).toLocaleString()}</small>
            </Col>
          ))}
        </Row>
        <br/><br/>
        {/*
        <Row>
          <Col xs={12} sm={6} md={4}>
            <img src={`${PUBLIC_API}api/objectives/${name}/red/requests.svg`}
                 alt=""
                 style={{ maxWidth: '100%' }}/>
          </Col>
          <Col xs={12} sm={6} md={4}/>
          <Col xs={12} sm={6} md={4}/>
        </Row>
        */}
        <br/><br/><br/><br/><br/>
      </Container>
    </div>
  );
};

export default App;
