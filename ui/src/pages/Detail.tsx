import {Link, useLocation, useNavigate} from 'react-router-dom'
import React, {useCallback, useEffect, useMemo, useState} from 'react'
import {
  Badge,
  Button,
  ButtonGroup,
  Col,
  Container,
  OverlayTrigger,
  Row,
  Spinner,
  Tooltip as OverlayTooltip,
} from 'react-bootstrap'
import {
  API_BASEPATH,
  formatDuration,
  hasObjectiveType,
  latencyTarget,
  ObjectiveType,
  renderLatencyTarget,
} from '../App'
import Navbar from '../components/Navbar'
import {MetricName, parseLabels} from '../labels'
import ErrorBudgetGraph from '../components/graphs/ErrorBudgetGraph'
import RequestsGraph from '../components/graphs/RequestsGraph'
import ErrorsGraph from '../components/graphs/ErrorsGraph'
import {
  Code,
  ConnectError,
  createConnectTransport,
  createPromiseClient,
} from '@bufbuild/connect-web'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connectweb'
import {
  Availability,
  Budget,
  GetStatusResponse,
  ListResponse,
  Objective,
} from '../proto/objectives/v1alpha1/objectives_pb'
import AlertsTable from '../components/AlertsTable'
import Toggle from '../components/Toggle'
import {Timestamp} from '@bufbuild/protobuf'
import DurationGraph from '../components/graphs/DurationGraph'
import uPlot from 'uplot'

const Detail = () => {
  const client = useMemo(() => {
    return createPromiseClient(
      ObjectiveService,
      createConnectTransport({
        baseUrl: API_BASEPATH,
      }),
    )
  }, [])

  const navigate = useNavigate()
  const {search} = useLocation()

  const {from, to, expr, grouping, groupingExpr, groupingLabels, name, labels} = useMemo(() => {
    const query = new URLSearchParams(search)

    const queryExpr = query.get('expr')
    const expr = queryExpr == null ? '' : queryExpr
    const labels = parseLabels(expr)

    const groupingExpr = query.get('grouping')
    const grouping = groupingExpr == null ? '' : groupingExpr
    const groupingLabels = parseLabels(grouping)

    const name: string = labels[MetricName]

    const toQuery = query.get('to')
    const to = toQuery != null ? parseInt(toQuery) : Date.now()

    const fromQuery = query.get('from')
    const from = fromQuery != null ? parseInt(fromQuery) : to - 3600 * 1000

    document.title = `${name} - Pyrra`

    return {from, to, expr, grouping, groupingExpr, groupingLabels, name, labels}
  }, [search])

  const [objective, setObjective] = useState<Objective | null>(null)
  const [objectiveError, setObjectiveError] = useState<string>('')
  const [autoReload, setAutoReload] = useState<boolean>(true)

  enum StatusState {
    Unknown,
    Error,
    NoData,
    Success,
  }

  const [statusState, setStatusState] = useState<StatusState>(StatusState.Unknown)
  const [availability, setAvailability] = useState<Availability | undefined>(undefined)
  const [errorBudget, setErrorBudget] = useState<Budget | undefined>(undefined)

  const getObjective = useCallback(() => {
    client
      .list({expr})
      .then((resp: ListResponse) => {
        if (resp.objectives.length === 1) {
          if (resp.objectives[0].config === objective?.config) {
            // Prevent the setState if the objective is the same
            return
          }
          setObjective(resp.objectives[0])
        } else {
          setObjective(null)
        }
      })
      .catch((err: ConnectError) => {
        console.log(err)
        setObjectiveError(err.rawMessage)
      })
  }, [client, expr, objective?.config])

  const getObjectiveStatus = useCallback(() => {
    client
      .getStatus({expr, grouping, time: Timestamp.fromDate(new Date(to))})
      .then((resp: GetStatusResponse) => {
        if (resp.status.length === 0) {
          if (statusState !== StatusState.NoData) {
            setStatusState(StatusState.NoData)
          }
        } else if (resp.status.length === 1) {
          setAvailability(resp.status[0].availability)
          setErrorBudget(resp.status[0].budget)
          setStatusState(StatusState.Success)
        } else {
          if (statusState !== StatusState.Error) {
            setStatusState(StatusState.Error)
          }
        }
      })
      .catch((err: ConnectError) => {
        if (err.code === Code.NotFound) {
          setStatusState(StatusState.NoData)
        } else {
          console.log(err)
          setStatusState(StatusState.Error)
        }
      })
  }, [
    client,
    expr,
    grouping,
    to,
    statusState,
    StatusState.NoData,
    StatusState.Success,
    StatusState.Error,
  ])

  useEffect(() => {
    getObjective()
    getObjectiveStatus()
  }, [getObjective, getObjectiveStatus])

  const updateTimeRange = useCallback(
    (from: number, to: number) => {
      navigate(`/objectives?expr=${expr}&grouping=${groupingExpr ?? ''}&from=${from}&to=${to}`)
    },
    [navigate, expr, groupingExpr],
  )

  const duration = to - from
  const interval = intervalFromDuration(duration)

  useEffect(() => {
    if (autoReload) {
      const id = setInterval(() => {
        const newTo = Date.now()
        const newFrom = newTo - duration
        updateTimeRange(newFrom, newTo)
      }, interval)

      return () => {
        clearInterval(id)
      }
    }
  }, [updateTimeRange, autoReload, duration, interval])

  const handleTimeRangeClick = (t: number) => () => {
    const to = Date.now()
    const from = to - t
    updateTimeRange(from, to)
  }

  if (objectiveError !== '') {
    return (
      <>
        <Navbar />
        <Container>
          <div className="header">
            <h3>{objectiveError}</h3>
            <br />
            <Link to="/" className="btn btn-light">
              Go Back
            </Link>
          </div>
        </Container>
      </>
    )
  }

  if (objective == null) {
    return (
      <div style={{marginTop: '50px', textAlign: 'center'}}>
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading...</span>
        </Spinner>
      </div>
    )
  }

  if (objective.labels === undefined) {
    return <></>
  }

  const timeRanges = [
    28 * 24 * 3600 * 1000, // 4w
    7 * 24 * 3600 * 1000, // 1w
    24 * 3600 * 1000, // 1d
    12 * 3600 * 1000, // 12h
    3600 * 1000, // 1h
  ]

  const objectiveType = hasObjectiveType(objective)

  const renderObjective = () => {
    switch (objectiveType) {
      case ObjectiveType.Ratio:
        return (
          <div>
            <h6 className="headline">Objective</h6>
            <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
            <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
          </div>
        )
      case ObjectiveType.Latency:
        return (
          <div>
            <h6 className="headline">Objective</h6>
            <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
            <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
            <br />
            <p className="details">faster than {renderLatencyTarget(objective)}</p>
          </div>
        )
      default:
        return <div></div>
    }
  }

  const renderAvailability = () => {
    const headline = <h6 className="headline">Availability</h6>
    switch (statusState) {
      case StatusState.Unknown:
        return (
          <div>
            {headline}
            <Spinner
              animation={'border'}
              style={{
                width: 50,
                height: 50,
                padding: 0,
                borderRadius: 50,
                borderWidth: 2,
                opacity: 0.25,
              }}
            />
          </div>
        )
      case StatusState.Error:
        return (
          <div>
            {headline}
            <h2 className="error">Error</h2>
          </div>
        )
      case StatusState.NoData:
        return (
          <div>
            {headline}
            <h2>No data</h2>
          </div>
        )
      case StatusState.Success:
        if (availability === undefined) {
          return <></>
        }
        return (
          <div className={availability.percentage > objective.target ? 'good' : 'bad'}>
            {headline}
            <h2 className="metric">{(100 * availability.percentage).toFixed(3)}%</h2>
            <table className="details">
              <tr>
                <td>{objectiveType === ObjectiveType.Latency ? 'Slow:' : 'Errors:'}</td>
                <td>{Math.floor(availability.errors).toLocaleString()}</td>
              </tr>
              <tr>
                <td>Total:</td>
                <td>{Math.floor(availability.total).toLocaleString()}</td>
              </tr>
            </table>
          </div>
        )
    }
  }

  const renderErrorBudget = () => {
    const headline = <h6 className="headline">Error Budget</h6>
    switch (statusState) {
      case StatusState.Unknown:
        return (
          <div>
            {headline}
            <Spinner
              animation={'border'}
              style={{
                width: 50,
                height: 50,
                padding: 0,
                borderRadius: 50,
                borderWidth: 2,
                opacity: 0.25,
              }}
            />
          </div>
        )
      case StatusState.Error:
        return (
          <div>
            {headline}
            <h2 className="error">Error</h2>
          </div>
        )
      case StatusState.NoData:
        return (
          <div>
            {headline}
            <h2>No data</h2>
          </div>
        )
      case StatusState.Success:
        if (errorBudget === undefined) {
          return <></>
        }
        return (
          <div className={errorBudget.remaining > 0 ? 'good' : 'bad'}>
            {headline}
            <h2 className="metric">{(100 * errorBudget.remaining).toFixed(3)}%</h2>
          </div>
        )
    }
  }

  const labelBadges = Object.entries({...objective.labels, ...groupingLabels})
    .filter((l: [string, string]) => l[0] !== MetricName)
    .map((l: [string, string]) => (
      <Badge key={l[0]} bg="light" text="dark" className="fw-normal">
        {l[0]}={l[1]}
      </Badge>
    ))

  const uPlotCursor: uPlot.Cursor = {
    y: false,
    lock: true,
    sync: {
      key: 'detail',
    },
  }

  return (
    <>
      <Navbar>
        <div>
          <Link to="/">Objectives</Link> &gt; <span>{name}</span>
        </div>
      </Navbar>

      <div className="content detail">
        <Container>
          <Row>
            <Col xs={12} className="col-xxxl-10 offset-xxxl-1 header">
              <h3>{name}</h3>
              {labelBadges}
            </Col>
            {objective.description !== undefined && objective.description !== '' ? (
              <Col xs={12} md={6} style={{marginTop: 12}} className="col-xxxl-10 offset-xxxl-1">
                <p>{objective.description}</p>
              </Col>
            ) : (
              <></>
            )}
          </Row>
          <Row>
            <Col className="col-xxxl-10 offset-xxxl-1">
              <div className="metrics">
                {renderObjective()}
                {renderAvailability()}
                {renderErrorBudget()}
              </div>
            </Col>
          </Row>
          <Row>
            <Col className="text-center timerange">
              <div className="inner">
                <ButtonGroup aria-label="Time Range">
                  {timeRanges.map((t: number) => (
                    <Button
                      key={t}
                      variant="light"
                      onClick={handleTimeRangeClick(t)}
                      active={to - from === t}>
                      {formatDuration(t)}
                    </Button>
                  ))}
                </ButtonGroup>
                &nbsp; &nbsp; &nbsp;
                <OverlayTrigger
                  key="auto-reload"
                  overlay={
                    <OverlayTooltip id={`tooltip-auto-reload`}>Automatically reload</OverlayTooltip>
                  }>
                  <span>
                    <Toggle
                      checked={autoReload}
                      onChange={() => setAutoReload(!autoReload)}
                      onText={formatDuration(interval)}
                    />
                  </span>
                </OverlayTrigger>
              </div>
            </Col>
          </Row>
          <Row>
            <Col>
              <ErrorBudgetGraph
                client={client}
                labels={labels}
                grouping={groupingLabels}
                from={from}
                to={to}
                uPlotCursor={uPlotCursor}
              />
            </Col>
          </Row>
          <Row>
            <Col
              xs={12}
              md={objectiveType === ObjectiveType.Latency ? 12 : 6}
              className={objectiveType === ObjectiveType.Latency ? 'col-xxxl-4' : ''}>
              <RequestsGraph
                client={client}
                labels={labels}
                grouping={groupingLabels}
                from={from}
                to={to}
                uPlotCursor={uPlotCursor}
              />
            </Col>
            <Col
              xs={12}
              md={objectiveType === ObjectiveType.Latency ? 12 : 6}
              className={objectiveType === ObjectiveType.Latency ? 'col-xxxl-4' : ''}>
              <ErrorsGraph
                client={client}
                type={objectiveType}
                labels={labels}
                grouping={groupingLabels}
                from={from}
                to={to}
                uPlotCursor={uPlotCursor}
              />
            </Col>
            {objectiveType === ObjectiveType.Latency ? (
              <Col xs={12} className="col-xxxl-4">
                <DurationGraph
                  client={client}
                  labels={labels}
                  grouping={groupingLabels}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  target={objective.target}
                  latency={latencyTarget(objective)}
                />
              </Col>
            ) : (
              <></>
            )}
          </Row>
          <Row>
            <Col className="col-xxxl-10 offset-xxxl-1">
              <h4>Multi Burn Rate Alerts</h4>
              <AlertsTable client={client} objective={objective} grouping={groupingLabels} />
            </Col>
          </Row>
          <Row>
            <Col className="col-xxxl-10 offset-xxxl-1">
              <h4>Config</h4>
              <pre style={{padding: 20, borderRadius: 4}}>
                <code>{objective.config}</code>
              </pre>
            </Col>
          </Row>
        </Container>
      </div>
    </>
  )
}

const intervalFromDuration = (duration: number): number => {
  // map some preset duration to nicer looking intervals
  switch (duration) {
    case 60 * 60 * 1000: // 1h => 10s
      return 10 * 1000
    case 12 * 60 * 60 * 1000: // 12h => 30s
      return 30 * 1000
    case 24 * 60 * 60 * 1000: // 12h => 30s
      return 90 * 1000
  }

  if (duration < 10 * 1000 * 1000) {
    return 10 * 1000
  }
  if (duration < 10 * 60 * 1000 * 1000) {
    return Math.floor(duration / 1000 / 1000) * 1000 // round to seconds
  }

  return Math.floor(duration / 60 / 1000 / 1000) * 60 * 1000
}

export default Detail
