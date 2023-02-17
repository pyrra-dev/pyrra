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
import {createConnectTransport, createPromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connectweb'
import AlertsTable from '../components/AlertsTable'
import Toggle from '../components/Toggle'
import DurationGraph from '../components/graphs/DurationGraph'
import uPlot from 'uplot'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_connectweb'
import {usePrometheusQuery} from '../prometheus'
import {useObjectivesList} from '../objectives'
import {Objective} from '../proto/objectives/v1alpha1/objectives_pb'

const Detail = () => {
  const client = useMemo(() => {
    return createPromiseClient(
      ObjectiveService,
      createConnectTransport({
        baseUrl: API_BASEPATH,
      }),
    )
  }, [])

  const promClient = useMemo(() => {
    return createPromiseClient(
      PrometheusService,
      // createGrpcWebTransport({ TODO: Use grpcWeb in production for efficiency?
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

  const [autoReload, setAutoReload] = useState<boolean>(true)

  const {
    response: objectiveResponse,
    error: objectiveError,
    status: objectiveStatus,
  } = useObjectivesList(client, expr, grouping)

  const objective: Objective | null = objectiveResponse?.objectives[0] ?? null

  const {response: totalResponse, status: totalStatus} = usePrometheusQuery(
    promClient,
    objective?.queries?.countTotal ?? '',
    to / 1000,
    {enabled: objectiveStatus === 'success' && objective?.queries?.countTotal !== undefined},
  )

  const {response: errorResponse, status: errorStatus} = usePrometheusQuery(
    promClient,
    objective?.queries?.countErrors ?? '',
    to / 1000,
    {enabled: objectiveStatus === 'success' && objective?.queries?.countTotal !== undefined},
  )

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

  if (objectiveError !== null) {
    return (
      <>
        <Navbar />
        <Container>
          <div className="header">
            <h3></h3>
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
      case ObjectiveType.BoolGauge:
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
    if (
      totalStatus === 'loading' ||
      totalStatus === 'idle' ||
      errorStatus === 'loading' ||
      errorStatus === 'idle'
    ) {
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
    }

    if (totalStatus === 'success' && errorStatus === 'success') {
      if (totalResponse?.options.case === 'vector' && errorResponse?.options.case === 'vector') {
        let errors = 0
        if (errorResponse.options.value.samples.length > 0) {
          errors = errorResponse.options.value.samples[0].value
        }

        let total = 1
        if (totalResponse.options.value.samples.length > 0) {
          total = totalResponse.options.value.samples[0].value
        }

        const percentage = 1 - errors / total

        return (
          <div className={percentage > objective.target ? 'good' : 'bad'}>
            {headline}
            <h2 className="metric">{(100 * percentage).toFixed(3)}%</h2>
            <table className="details">
              <tbody>
                <tr>
                  <td>{objectiveType === ObjectiveType.Latency ? 'Slow:' : 'Errors:'}</td>
                  <td>{Math.floor(errors).toLocaleString()}</td>
                </tr>
                <tr>
                  <td>Total:</td>
                  <td>{Math.floor(total).toLocaleString()}</td>
                </tr>
              </tbody>
            </table>
          </div>
        )
      } else {
        return (
          <div>
            {headline}
            <h2>No data</h2>
          </div>
        )
      }
    }

    return (
      <div>
        <>
          {headline}
          <h2 className="error">Error</h2>
        </>
      </div>
    )
  }

  const renderErrorBudget = () => {
    const headline = <h6 className="headline">Error Budget</h6>

    if (
      totalStatus === 'loading' ||
      totalStatus === 'idle' ||
      errorStatus === 'loading' ||
      errorStatus === 'idle'
    ) {
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
    }
    if (totalStatus === 'success' && errorStatus === 'success') {
      if (totalResponse?.options.case === 'vector' && errorResponse?.options.case === 'vector') {
        let errors = 0
        if (errorResponse.options.value.samples.length > 0) {
          errors = errorResponse.options.value.samples[0].value
        }

        let total = 1
        if (totalResponse.options.value.samples.length > 0) {
          total = totalResponse.options.value.samples[0].value
        }

        const budget = 1 - objective.target
        const unavailability = errors / total
        const availableBudget = (budget - unavailability) / budget

        return (
          <div className={availableBudget > 0 ? 'good' : 'bad'}>
            {headline}
            <h2 className="metric">{(100 * availableBudget).toFixed(3)}%</h2>
          </div>
        )
      } else {
        return (
          <div>
            {headline}
            <h2>No data</h2>
          </div>
        )
      }
    }
    return (
      <div>
        {headline}
        <h2 className="error">Error</h2>
      </div>
    )
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
              <Col
                xs={12}
                md={6}
                style={{marginTop: labelBadges.length > 0 ? 12 : 0}}
                className="col-xxxl-5 offset-xxxl-1">
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
              {objective.queries?.graphErrorBudget !== undefined ? (
                <ErrorBudgetGraph
                  client={promClient}
                  query={objective.queries.graphErrorBudget}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                />
              ) : (
                <></>
              )}
            </Col>
          </Row>
          <Row>
            <Col
              xs={12}
              md={objectiveType === ObjectiveType.Latency ? 12 : 6}
              className={objectiveType === ObjectiveType.Latency ? 'col-xxxl-4' : ''}>
              {objective.queries?.graphRequests !== undefined ? (
                <RequestsGraph
                  client={promClient}
                  query={objective.queries.graphRequests}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  type={objectiveType}
                />
              ) : (
                <></>
              )}
            </Col>
            <Col
              xs={12}
              md={objectiveType === ObjectiveType.Latency ? 12 : 6}
              className={objectiveType === ObjectiveType.Latency ? 'col-xxxl-4' : ''}>
              {objective.queries?.graphErrors !== undefined ? (
                <ErrorsGraph
                  client={promClient}
                  type={objectiveType}
                  query={objective.queries.graphErrors}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                />
              ) : (
                <></>
              )}
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
            <Col>
              <h4>Multi Burn Rate Alerts</h4>
              <AlertsTable client={client} objective={objective} grouping={groupingLabels} />
            </Col>
          </Row>
          <Row>
            <Col>
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
