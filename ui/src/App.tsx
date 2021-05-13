import React, { useEffect, useReducer, useState } from 'react';
import './App.css';
import { Col, Container, Row, Spinner, Table } from 'react-bootstrap';
import { BrowserRouter, Link, Route, RouteComponentProps, Switch } from 'react-router-dom';

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
        <Route exact={true} path="/">
          <List/>
        </Route>
        <Route path="/objective/:name" component={Details}/>
      </Switch>
    </BrowserRouter>
  )
}

interface ObjectiveStatus {
  availability: number
  budget: number
}

const fetchObjectives = async (): Promise<Array<Objective>> => {
  const resp: Response = await fetch(`/api/objectives`)
  const json = await resp.json()
  if (!resp.ok) {
    return Promise.reject(resp)
  }
  return json
}

const fetchObjectiveStatus = async (name: string): Promise<ObjectiveStatus> => {
  const resp: Response = await fetch(`/api/objectives/${name}/status`)
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

enum TableActionType { SetObjective}

type TableAction =
  | { type: TableActionType.SetObjective, objective: Objective }

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
    default:
      return state
  }
}

const List = () => {
  const [objectives, setObjectives] = useState<Array<Objective>>([])
  const initialTableState: TableState = { objectives: {} }
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)

  useEffect(() => {
    fetchObjectives()
      .then((objectives: Objective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [])

  useEffect(() => {
    objectives
      .sort((a: Objective, b: Objective) => a.name.localeCompare(b.name))
      .forEach((o: Objective) => {
        dispatchTable({ type: TableActionType.SetObjective, objective: o })
      })

  }, [objectives])

  const tableList = Object.keys(table.objectives)
    .map((k: string) => table.objectives[k])
    .sort((a: TableObjective, b: TableObjective) => a.name.localeCompare(b.name))

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
          <th>Name</th>
          <th>Time Window</th>
          <th>Objective</th>
          <th>Availability</th>
          <th>Error Budget</th>
        </tr>
        </thead>
        <tbody>
        {tableList.map((o: TableObjective) => (
          <tr key={o.name}>
            <td>
              <Link to={`/objective/${o.name}`}>{o.name}</Link>
            </td>
            <td>{o.window}</td>
            <td>
              {(100 * o.target).toFixed(1)}%
            </td>
            <td>
              {o.availability !== undefined ?
                o.availability :
                <Spinner animation={'border'} style={{ opacity: 0.1 }}/>}
            </td>
            <td>
              {o.budget !== undefined ?
                o.budget :
                <Spinner animation={'border'} style={{ opacity: 0.1 }}/>}
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

  useEffect(() => {
    fetch(`/api/objectives/${name}`)
      .then((resp: Response) => resp.json())
      .then((data) => setObjective(data))
  }, [name])

  useEffect(() => {
    fetch(`/api/objectives/${name}/status`)
      .then((resp: Response) => resp.json())
      .then((data) => {
        setAvailability(data.availability)
        setErrorBudget(data.budget)
      })
  }, [name])

  useEffect(() => {
    fetch(`/api/objectives/${name}/valet`)
      .then((resp: Response) => resp.json())
      .then((data) => setValets(data))
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
        <Row>
          <Col>
            <h1>{objective?.name}</h1>
          </Col>
        </Row>
        <br/><br/><br/>
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
          <img src={`/api/objectives/${name}/errorbudget.svg`}
               alt=""
               style={{ maxWidth: '100%' }}/>
        </Row>
        <br/><br/>
        <Row>
          <table className="table">
            <thead>
            <tr>
              <th scope="col">Window</th>
              <th scope="col">Volume</th>
              <th scope="col">Errors</th>
              <th scope="col">Availability</th>
              <th scope="col">Error Budget</th>
            </tr>
            </thead>
            <tbody>
            {valets.map((v: Valet) => (
              <tr key={v.window}>
                <td>{v.window}</td>
                <td>{Math.floor(v.volume)}</td>
                <td>{Math.floor(v.errors)}</td>
                <td>{(100 * v.availability).toFixed(3)}%</td>
                <td>{(100 * v.budget).toFixed(3)}%</td>
              </tr>
            ))}
            </tbody>
          </table>
        </Row>
      </Container>
    </div>
  );
};

export default App;
