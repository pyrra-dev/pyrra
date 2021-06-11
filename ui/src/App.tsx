import React, { useEffect, useReducer, useState } from 'react';
import './App.css';
import { Col, Container, Row, Spinner, Table } from 'react-bootstrap';
import { BrowserRouter, Link, Route, RouteComponentProps, Switch, useHistory } from 'react-router-dom';
import { Area, AreaChart, ResponsiveContainer, Tooltip, TooltipProps, XAxis, YAxis } from 'recharts'
import {
  Configuration,
  ErrorBudget as APIErrorBudget,
  ErrorBudgetPair,
  Objective as APIObjective,
  ObjectivesApi,
  ObjectiveStatus as APIObjectiveStatus,
  ObjectiveStatusAvailability,
  ObjectiveStatusBudget,
  QueryRange
} from './client'
import AlertsTable from './components/AlertsTable'

// @ts-ignore - this is passed from the HTML template.
export const PUBLIC_API: string = window.PUBLIC_API;
const APIConfiguration = new Configuration({ basePath: `${PUBLIC_API}api/v1` })
export const APIObjectives = new ObjectivesApi(APIConfiguration)

export interface Objective {
  name: string
  target: number
  window: string
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

// TableObjective extends Objective to add some more additional (async) properties
interface TableObjective extends APIObjective {
  availability?: number
  budget?: number
}

interface TableState {
  objectives: { [key: string]: TableObjective }
}

enum TableActionType { SetObjective, SetStatus }

type TableAction =
  | { type: TableActionType.SetObjective, objective: APIObjective }
  | { type: TableActionType.SetStatus, name: string, status: APIObjectiveStatus }

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
            availability: action.status.availability?.percentage,
            budget: action.status.budget?.remaining
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
  const [objectives, setObjectives] = useState<Array<APIObjective>>([])
  const initialTableState: TableState = { objectives: {} }
  const [table, dispatchTable] = useReducer(tableReducer, initialTableState)
  const [tableSortState, setTableSortState] = useState<TableSorting>({
    type: TableSortType.Budget,
    order: TableSortOrder.Ascending
  })

  useEffect(() => {
    APIObjectives.listObjectives()
      .then((objectives: APIObjective[]) => setObjectives(objectives))
      .catch((err) => console.log(err))
  }, [])

  useEffect(() => {
    // const controller = new AbortController()
    // const signal = controller.signal // TODO: Pass this to the generated fetch client?

    objectives
      .sort((a: APIObjective, b: APIObjective) => a.name.localeCompare(b.name))
      .forEach((o: APIObjective) => {
        dispatchTable({ type: TableActionType.SetObjective, objective: o })

        APIObjectives.getObjectiveStatus({ name: o.name })
          .then((s: APIObjectiveStatus) => {
            dispatchTable({ type: TableActionType.SetStatus, name: o.name, status: s })
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
            <td>{formatDuration(o.window)}</td>
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

interface SamplePair {
  t: number
  v: number
}

const Details = (params: RouteComponentProps<DetailsRouteParams>) => {
  const name = params.match.params.name;
  const [objective, setObjective] = useState<APIObjective | null>(null);
  const [availability, setAvailability] = useState<ObjectiveStatusAvailability | null>(null);
  const [errorBudget, setErrorBudget] = useState<ObjectiveStatusBudget | null>(null);

  const [errorBudgetSamples, setErrorBudgetSamples] = useState<ErrorBudgetPair[]>([]);
  const [errorBudgetSamplesOffset, setErrorBudgetSamplesOffset] = useState<number>(0)
  const [errorBudgetSamplesMin, setErrorBudgetSamplesMin] = useState<number>(-10000)
  const [errorBudgetSamplesMax, setErrorBudgetSamplesMax] = useState<number>(1)
  const [errorBudgetSamplesLoading, setErrorBudgetSamplesLoading] = useState<boolean>(true)

  const [requests, setRequests] = useState<SamplePair[]>([])
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [errors, setErrors] = useState<SamplePair[]>([])
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])

  useEffect(() => {
    const controller = new AbortController()

    APIObjectives.getObjective({ name })
      .then((o: APIObjective) => setObjective(o))

    APIObjectives.getObjectiveStatus({ name })
      .then((s: APIObjectiveStatus) => {
        setAvailability(s.availability)
        setErrorBudget(s.budget)
      })

    setErrorBudgetSamplesLoading(true)

    // TODO: Pass as query parameters
    // const end = Math.ceil(Date.now() / 1000)
    // const start = end - 3600

    APIObjectives.getObjectiveErrorBudget({ name })
      .then((b: APIErrorBudget) => {
        setErrorBudgetSamples(b.pair)

        // Calculate the offset to split the errorbudget into green and red areas
        const min = Math.floor(10000 * Math.min(...b.pair.map((o: ErrorBudgetPair) => o.v))) / 10000;
        const max = Math.ceil(10000 * Math.max(...b.pair.map((o: ErrorBudgetPair) => o.v))) / 10000;
        setErrorBudgetSamplesMin(min === 1 ? 0 : min)
        setErrorBudgetSamplesMax(max)
        if (max <= 0) {
          setErrorBudgetSamplesOffset(0)
        } else if (min >= 1) {
          setErrorBudgetSamplesOffset(1)
        } else {
          setErrorBudgetSamplesOffset(max / (max - min))
        }
      })
      .finally(() => setErrorBudgetSamplesLoading(false))

    APIObjectives.getREDRequests({ name })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i][j - 1] = v
            }
          })
        })
        setRequestsLabels(r.labels)
        setRequests(data)
      })

    APIObjectives.getREDErrors({ name })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            }else {
              data[i][j - 1] = v
            }
          })
        })
        setErrorsLabels(r.labels)
        setErrors(data)
      })

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


  const RequestTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
    const style = {
      padding: 10,
      paddingTop: 5,
      paddingBottom: 5,
      backgroundColor: 'white',
      border: '1px solid #666',
      borderRadius: 3
    }
    if (payload !== undefined && payload?.length > 0) {
      return (
        <div className="area-chart-tooltip" style={style}>
          Date: {dateFormatterFull(payload[0].payload.t)}<br/>
          {Object.keys(payload[0].payload).filter((k) => k !== 't').map((k: string, i: number) => (
            <>
              <span>{requestsLabels[i]}: {(payload[0].payload[k]).toFixed(2)} req/s</span><br/>
            </>
          ))}
        </div>
      )
    }
    return <></>
  }

  const ErrorsTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
    const style = {
      padding: 10,
      paddingTop: 5,
      paddingBottom: 5,
      backgroundColor: 'white',
      border: '1px solid #666',
      borderRadius: 3
    }
    if (payload !== undefined && payload?.length > 0) {
      return (
        <div className="area-chart-tooltip" style={style}>
          Date: {dateFormatterFull(payload[0].payload.t)}<br/>
          {Object.keys(payload[0].payload).filter((k) => k !== 't').map((k: string, i: number) => (
            <>
              <span>{errorsLabels[i]}: {(payload[0].payload[k]).toFixed(2)}%</span><br/>
            </>
          ))}
        </div>
      )
    }
    return <></>
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
              <h6 className="text-muted">Objective in {formatDuration(objective.window)}</h6>
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
              )}
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
          <Col>

          </Col>
        </Row>
        <Row>
          {errorBudgetSamplesLoading ?
            <div style={{
              width: '100%',
              height: 230,
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center'
            }}>
              <Spinner animation="border" style={{ margin: '0 auto' }}/>
            </div>
            : <ResponsiveContainer height={300}>
              <AreaChart height={300} data={errorBudgetSamples}>
                <XAxis
                  type="number"
                  dataKey="t"
                  tickCount={7}
                  tickFormatter={dateFormatter}
                  domain={[errorBudgetSamples[0].t, errorBudgetSamples[errorBudgetSamples.length - 1].t]}
                />
                <YAxis
                  tickCount={5}
                  unit="%"
                  tickFormatter={(v: number) => (100 * v).toFixed(2)}
                  domain={[errorBudgetSamplesMin, errorBudgetSamplesMax]}
                />
                <Tooltip content={<DateTooltip/>}/>
                <defs>
                  <linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
                    <stop offset={errorBudgetSamplesOffset} stopColor={`#${greens[0]}`} stopOpacity={1}/>
                    <stop offset={errorBudgetSamplesOffset} stopColor={`#${reds[0]}`} stopOpacity={1}/>
                  </linearGradient>
                </defs>
                <Area dataKey="v" type="monotone" animationDuration={250} strokeWidth={0} fill="url(#splitColor)" fillOpacity={1}/>
              </AreaChart>
            </ResponsiveContainer>
          }
          <Col style={{ textAlign: 'right' }}>
            {availability != null ? (
              <>
                <small>Errors: {Math.floor(availability.errors).toLocaleString()}</small>&nbsp;
                <small>Volume: {Math.floor(availability.total).toLocaleString()}</small>&nbsp;
              </>
            ) : (
              <></>
            )}
          </Col>
        </Row>
        <br/><br/>
        <Row>
          <Col xs={12} sm={6}>
            <h4>Requests</h4>
            {requests.length > 0 && requestsLabels.length > 0 ? (
              <ResponsiveContainer height={150}>
                <AreaChart height={150} data={requests}>
                  <XAxis
                    type="number"
                    dataKey="t"
                    tickCount={3}
                    tickFormatter={dateFormatter}
                    domain={[requests[0].t, requests[requests.length - 1].t]}
                  />
                  <YAxis
                    tickCount={3}
                    // tickFormatter={(v: number) => (100 * v).toFixed(2)}
                    // domain={[0, 10]}
                  />
                  {Object.keys(requests[0]).filter((k: string) => k !== 't').map((k: string, i: number) => {
                    const label = requestsLabels[parseInt(k)]
                    console.log(label)
                    let color = ''
                    if (label.match(/"(2\d{2}|OK)"/) != null) {
                      color = greens[i]
                    }
                    if (label.match(/"(3\d{2})"/) != null) {
                      color = yellows[i]
                    }
                    if (label.match(/"(4\d{2}|Canceled|InvalidArgument|NotFound|AlreadyExists|PermissionDenied|Unauthenticated|ResourceExhausted|FailedPrecondition|Aborted|OutOfRange)"/) != null) {
                      color = blues[i]
                    }
                    if (label.match(/"(5\d{2}|Unknown|DeadlineExceeded|Unimplemented|Internal|Unavailable|DataLoss)"/) != null) {
                      color = reds[i]
                    }

                    return <Area
                      key={k}
                      type="monotone"
                      animationDuration={250}
                      dataKey={k}
                      stackId={1}
                      strokeWidth={0}
                      fill={`#${color}`}
                      fillOpacity={1}/>
                  })}
                  <Tooltip content={RequestTooltip}/>
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <></>
            )}
          </Col>
          <Col xs={12} sm={6}>
            <h4>Errors</h4>
            {errors.length > 0 && errorsLabels.length > 0 ? (
              <ResponsiveContainer height={150}>
                <AreaChart height={150} data={errors}>
                  <XAxis
                    type="number"
                    dataKey="t"
                    tickCount={3}
                    tickFormatter={dateFormatter}
                    domain={[errors[0].t, errors[errors.length - 1].t]}
                  />
                  <YAxis
                    tickCount={3}
                    unit="%"
                    // tickFormatter={(v: number) => (100 * v).toFixed(2)}
                    // domain={[0, 10]}
                  />
                  {Object.keys(errors[0]).filter((k: string) => k !== 't').map((k: string, i: number) => {
                    return <Area
                      key={k}
                      type="monotone"
                      animationDuration={250}
                      dataKey={k}
                      stackId={1}
                      strokeWidth={0}
                      fill={`#${reds[i]}`}
                      fillOpacity={1}/>
                  })}
                  <Tooltip content={ErrorsTooltip}/>
                </AreaChart>
              </ResponsiveContainer>
            ) : (
              <></>
            )}
          </Col>
        </Row>
        <br/><br/><br/>
        <Row>
          <Col>
            <h4>Multi Burn Rate Alerts</h4>
            <AlertsTable objective={objective}/>
          </Col>
        </Row>
      </Container>
    </div>
  );
};

const dateFormatter = (t: number): string => {
  const date = new Date(t * 1000)
  const year = date.getUTCFullYear()
  const month = date.getUTCMonth() + 1
  const day = date.getUTCDate()

  const monthLeading = month > 9 ? month : `0${month}`
  const dayLeading = day > 9 ? day : `0${day}`

  return `${year}-${monthLeading}-${dayLeading}`
}

const dateFormatterFull = (t: number): string => {
  const date = new Date(t * 1000)
  const year = date.getUTCFullYear()
  const month = date.getUTCMonth() + 1
  const day = date.getUTCDate()
  const hour = date.getUTCHours()
  const minute = date.getUTCMinutes()

  const monthLeading = month > 9 ? month : `0${month}`
  const dayLeading = day > 9 ? day : `0${day}`
  const hourLeading = hour > 9 ? hour : `0${hour}`
  const minuteLeading = minute > 9 ? minute : `0${minute}`

  return `${year}-${monthLeading}-${dayLeading} ${hourLeading}:${minuteLeading}`
}

const DateTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
  const style = {
    padding: 10,
    paddingTop: 5,
    paddingBottom: 5,
    backgroundColor: 'white',
    border: '1px solid #666',
    borderRadius: 3
  }
  if (payload !== undefined && payload?.length > 0) {
    return (
      <div className="area-chart-tooltip" style={style}>
        Date: {dateFormatterFull(payload[0].payload.t)}<br/>
        Value: {(100 * payload[0].payload.v).toFixed(3)}%
      </div>
    )
  }
  return <></>
}

export default App;

const greens = [
  '1B5E20',
  '2E7D32',
  '388E3C',
  '43A047',
  '4CAF50',
  '66BB6A',
  '81C784',
  'A5D6A7',
  'C8E6C9'
]
const blues = [
  "0D47A1",
  "1565C0",
  "1976D2",
  "1E88E5",
  "2196F3",
  "42A5F5",
  "64B5F6",
  "90CAF9",
  "BBDEFB"
]
const reds = [
  "B71C1C",
  "C62828",
  "D32F2F",
  "E53935",
  "F44336",
  "EF5350",
  "E57373",
  "EF9A9A",
  "FFCDD2"
]
const yellows = [
  "F57F17",
  "F9A825",
  "FBC02D",
  "FDD835",
  "FFEB3B",
  "FFEE58",
  "FFF176",
  "FFF59D",
  "FFF9C4"
]

// From prometheus/prometheus

export const formatDuration = (d: number): string => {
  let ms = d;
  let r = '';
  if (ms === 0) {
    return '0s';
  }

  const f = (unit: string, mult: number, exact: boolean) => {
    if (exact && ms % mult !== 0) {
      return;
    }
    const v = Math.floor(ms / mult);
    if (v > 0) {
      r += `${v}${unit}`;
      ms -= v * mult;
    }
  };

  // Only format years and weeks if the remainder is zero, as it is often
  // easier to read 90d than 12w6d.
  f('y', 1000 * 60 * 60 * 24 * 365, true);
  f('w', 1000 * 60 * 60 * 24 * 7, true);

  f('d', 1000 * 60 * 60 * 24, false);
  f('h', 1000 * 60 * 60, false);
  f('m', 1000 * 60, false);
  f('s', 1000, false);
  f('ms', 1, false);

  return r;
};