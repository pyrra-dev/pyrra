import { Link, RouteComponentProps, useHistory, useLocation } from 'react-router-dom'
import React, { useEffect, useState } from 'react'
import {
  Objective as APIObjective,
  ObjectiveStatus as APIObjectiveStatus,
  ObjectiveStatusAvailability,
  ObjectiveStatusBudget,
  QueryRange
} from '../client'
import { Button, ButtonGroup, Col, Container, Row, Spinner } from 'react-bootstrap'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, TooltipProps, XAxis, YAxis } from 'recharts'
import AlertsTable from '../components/AlertsTable'
import { APIObjectives, dateFormatter, dateFormatterFull, formatDuration, parseDuration, PROMETHEUS_URL } from '../App'

interface DetailRouteParams {
  name: string
  namespace: string
}

const Detail = (params: RouteComponentProps<DetailRouteParams>) => {
  const { namespace, name } = params.match.params;

  const history = useHistory()
  const query = new URLSearchParams(useLocation().search)

  const timeRangeQuery = query.get('timerange')
  const timeRangeParsed = timeRangeQuery != null ? parseDuration(timeRangeQuery) : null
  const timeRange: number = timeRangeParsed != null ? timeRangeParsed : 3600 * 1000

  const [objective, setObjective] = useState<APIObjective | null>(null);
  const [availability, setAvailability] = useState<ObjectiveStatusAvailability | null>(null);
  const [errorBudget, setErrorBudget] = useState<ObjectiveStatusBudget | null>(null);

  const [errorBudgetSamples, setErrorBudgetSamples] = useState<any[]>([]);
  const [errorBudgetSamplesOffset, setErrorBudgetSamplesOffset] = useState<number>(0)
  const [errorBudgetSamplesMin, setErrorBudgetSamplesMin] = useState<number>(-10000)
  const [errorBudgetSamplesMax, setErrorBudgetSamplesMax] = useState<number>(1)
  const [errorBudgetQuery, setErrorBudgetQuery] = useState<string>('')
  const [errorBudgetSamplesLoading, setErrorBudgetSamplesLoading] = useState<boolean>(true)

  const [requests, setRequests] = useState<any[]>([])
  const [requestsQuery, setRequestsQuery] = useState<string>('')
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [requestsLoading, setRequestsLoading] = useState<boolean>(true)
  const [errors, setErrors] = useState<any[]>([])
  const [errorsQuery, setErrorsQuery] = useState<string>('')
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])
  const [errorsLoading, setErrorsLoading] = useState<boolean>(true)

  useEffect(() => {
    const controller = new AbortController()

    document.title = `${name} - Pyrra`

    APIObjectives.getObjective({ namespace, name })
      .then((o: APIObjective) => setObjective(o))

    APIObjectives.getObjectiveStatus({ namespace, name })
      .then((s: APIObjectiveStatus) => {
        setAvailability(s.availability)
        setErrorBudget(s.budget)
      })

    setErrorBudgetSamplesLoading(true)

    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    APIObjectives.getObjectiveErrorBudget({ namespace, name, start, end })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i].v = v
            }
          })
        })

        const minRaw = Math.min(...data.map((o) => o.v))
        const maxRaw = Math.max(...data.map((o) => o.v))
        const diff = maxRaw - minRaw

        let roundBy = 1
        if (diff < 1) {
          roundBy = 10
        }
        if (diff < 0.1) {
          roundBy = 100
        }
        if (diff < 0.01) {
          roundBy = 1_000
        }

        // Calculate the offset to split the error budget into green and red areas
        const min = Math.floor(minRaw * roundBy) / roundBy;
        const max = Math.ceil(maxRaw * roundBy) / roundBy;

        setErrorBudgetSamplesMin(min === 1 ? 0 : min)
        setErrorBudgetSamplesMax(max)
        if (max <= 0) {
          setErrorBudgetSamplesOffset(0)
        } else if (min >= 1) {
          setErrorBudgetSamplesOffset(1)
        } else {
          setErrorBudgetSamplesOffset(maxRaw / (maxRaw - minRaw))
        }
        setErrorBudgetSamples(data)
        setErrorBudgetQuery(r.query)
      })
      .finally(() => setErrorBudgetSamplesLoading(false))

    setRequestsLoading(true)
    APIObjectives.getREDRequests({ namespace, name, start, end })
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
        setRequestsQuery(r.query)
        setRequests(data)
      }).finally(() => setRequestsLoading(false))

    setErrorsLoading(true)
    APIObjectives.getREDErrors({ namespace, name, start, end })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i][j - 1] = 100 * v
            }
          })
        })
        setErrorsLabels(r.labels)
        setErrorsQuery(r.query)
        setErrors(data)
      }).finally(() => setErrorsLoading(false))

    return () => {
      // cancel any pending requests.
      controller.abort()
    }
  }, [namespace, name, timeRange])

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
            <div key={i}>
              {requestsLabels[i]}: {(payload[0].payload[k]).toFixed(2)} req/s
            </div>
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
            <div key={i}>
              {errorsLabels[i]}: {(payload[0].payload[k]).toFixed(2)}%
            </div>
          ))}
        </div>
      )
    }
    return <></>
  }

  const timeRanges = [
    28 * 24 * 3600 * 1000, // 4w
    7 * 24 * 3600 * 1000, // 1w
    24 * 3600 * 1000, // 1d
    12 * 3600 * 1000, // 12h
    3600 * 1000 // 1h
  ]

  const handleTimeRangeClick = (t: number) => () => {
    history.push(`/objectives/${namespace}/${name}?timerange=${formatDuration(t)}`)
  }

  return (
    <div className="App">
      <Container>
        <Row>
          <Col>
            <Link to="/">⬅️ Overview</Link>
          </Col>
        </Row>
        <Row>
          <Col xs={12}>
            <h1>{objective.name}</h1>
          </Col>
          {objective.description !== undefined && objective.description !== '' ? (
              <Col xs={12} md={6}>
                <p>{objective.description}</p>
              </Col>
            )
            : (
              <></>
            )}
        </Row>
        <Row>
          <Col className="text-right">
            <ButtonGroup aria-label="Basic example">
              {timeRanges.map((t: number, i: number) => (
                <Button key={i} variant="light" onClick={handleTimeRangeClick(t)}
                        active={timeRange === t}>{formatDuration(t)}</Button>
              ))}
            </ButtonGroup>
          </Col>
        </Row>
        <Row>
          <Col xs={12} sm={12} md={4} className="metric">
            <div>
              <h2>{100 * objective.target}%</h2>
              <h6 className="text-muted">Objective in <strong>{formatDuration(objective.window)}</strong></h6>
            </div>
          </Col>
          <Col xs={12} sm={6} md={4} className="metric">
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
          <Col xs={12} sm={6} md={4} className="metric">
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
        <Row style={{ marginBottom: 0 }}>
          <Col>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <h4>
                Error Budget
                {errorBudgetSamplesLoading && errorBudgetSamples.length !== 0 ? (
                  <Spinner animation="border" style={{
                    marginLeft: '1rem',
                    marginBottom: '0.5rem',
                    width: '1rem',
                    height: '1rem',
                    borderWidth: '1px'
                  }}/>
                ) : <></>}
              </h4>
              {errorBudgetQuery !== '' ? (
                <a className="external-prometheus"
                   target="_blank"
                   rel="noreferrer"
                   href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(errorBudgetQuery)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
                  <PrometheusLogo/>
                </a>
              ) : <></>}
            </div>
            {errorBudgetSamplesLoading && errorBudgetSamples.length === 0 ?
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
                  <CartesianGrid strokeDasharray="3 3"/>
                  <Tooltip content={<DateTooltip/>}/>
                  <defs>
                    <linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
                      <stop offset={errorBudgetSamplesOffset} stopColor={`#${greens[0]}`} stopOpacity={1}/>
                      <stop offset={errorBudgetSamplesOffset} stopColor={`#${reds[0]}`} stopOpacity={1}/>
                    </linearGradient>
                  </defs>
                  <Area
                    dataKey="v"
                    type="monotone"
                    connectNulls={false}
                    animationDuration={250}
                    strokeWidth={0}
                    fill="url(#splitColor)"
                    fillOpacity={1}/>
                </AreaChart>
              </ResponsiveContainer>
            }
          </Col>
        </Row>
        <Row style={{ margin: 0 }}>
          <Col style={{ textAlign: 'right' }}>
            {availability != null ? (
              <>
                <small>Errors: {Math.floor(availability.errors).toLocaleString()}</small>&nbsp;
                <small>Total: {Math.floor(availability.total).toLocaleString()}</small>&nbsp;
              </>
            ) : (
              <></>
            )}
          </Col>
        </Row>
        <Row>
          <Col xs={12} sm={6}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <h4>
                Requests
                {requestsLoading ? <Spinner animation="border" style={{
                  marginLeft: '1rem',
                  marginBottom: '0.5rem',
                  width: '1rem',
                  height: '1rem',
                  borderWidth: '1px'
                }}/> : <></>}
              </h4>
              {requestsQuery !== '' ? (
                <a className="external-prometheus"
                   target="_blank"
                   rel="noreferrer"
                   href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(requestsQuery)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
                  <PrometheusLogo/>
                </a>
              ) : <></>}
            </div>
            {requests.length > 0 && requestsLabels.length > 0 ? (
              <ResponsiveContainer height={150}>
                <AreaChart height={150} data={requests}>
                  <CartesianGrid strokeDasharray="3 3"/>
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
                    if (label === undefined) {
                      return <></>
                    }
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
                      connectNulls={false}
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
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <h4>
                Errors
                {errorsLoading ? <Spinner animation="border" style={{
                  marginLeft: '1rem',
                  marginBottom: '0.5rem',
                  width: '1rem',
                  height: '1rem',
                  borderWidth: '1px'
                }}/> : <></>}
              </h4>
              {errorsQuery !== '' ? (
                <a className="external-prometheus"
                   target="_blank"
                   rel="noreferrer"
                   href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(errorsQuery)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
                  <PrometheusLogo/>
                </a>
              ) : <></>}
            </div>
            {errors.length > 0 && errorsLabels.length > 0 ? (
              <ResponsiveContainer height={150}>
                <AreaChart height={150} data={errors}>
                  <CartesianGrid strokeDasharray="3 3"/>
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
                      connectNulls={false}
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
        <Row>
          <Col>
            <h4>Multi Burn Rate Alerts</h4>
            <AlertsTable objective={objective}/>
          </Col>
        </Row>
        <Row>
          <Col>
            <h4>Config</h4>
            <pre style={{ padding: 20, backgroundColor: '#f8f9fa', borderRadius: 5 }}>
            <code>
              {objective.config}
            </code>
            </pre>
          </Col>
        </Row>
      </Container>
    </div>
  );
};

export default Detail

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

