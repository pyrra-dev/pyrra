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
import {API_BASEPATH, hasObjectiveType, latencyTarget, ObjectiveType} from '../App'
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
import {replaceInterval, usePrometheusQuery} from '../prometheus'
import {useObjectivesList} from '../objectives'
import {Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration, parseDuration} from '../duration'
import ObjectiveTile from '../components/tiles/ObjectiveTile'
import AvailabilityTile from '../components/tiles/AvailabilityTile'
import ErrorBudgetTile from '../components/tiles/ErrorBudgetTile'
import Tiles from '../components/tiles/Tiles'

const Detail = () => {
  const baseUrl = API_BASEPATH === undefined ? 'http://localhost:9099' : API_BASEPATH

  const client = useMemo(() => {
    return createPromiseClient(ObjectiveService, createConnectTransport({baseUrl}))
  }, [baseUrl])

  const promClient = useMemo(() => {
    return createPromiseClient(PrometheusService, createConnectTransport({baseUrl}))
  }, [baseUrl])

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

    let to: number = Date.now()
    const toQuery = query.get('to')
    if (toQuery !== null) {
      if (!toQuery.includes('now')) {
        to = parseInt(toQuery)
      }
    }

    let from: number = to - 60 * 60 * 1000
    const fromQuery = query.get('from')
    if (fromQuery !== null) {
      if (fromQuery.includes('now')) {
        const duration = parseDuration(fromQuery.substring(4)) // omit first 4 chars: `now-`
        if (duration !== null) {
          from = to - duration
        }
      } else {
        from = parseInt(fromQuery)
      }
    }

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
    (from: number, to: number, absolute: boolean) => {
      let fromStr = from.toString()
      let toStr = to.toString()
      if (!absolute) {
        fromStr = `now-${formatDuration(to - from)}`
        toStr = 'now'
      }
      navigate(
        `/objectives?expr=${expr}&grouping=${groupingExpr ?? ''}&from=${fromStr}&to=${toStr}`,
      )
    },
    [navigate, expr, groupingExpr],
  )

  const updateTimeRangeSelect = (min: number, max: number, absolute: boolean) => {
    // when selecting time ranges with the mouse we want to disable the auto refresh
    setAutoReload(false)
    updateTimeRange(min, max, absolute)
  }

  const duration = to - from
  const interval = intervalFromDuration(duration)

  useEffect(() => {
    if (autoReload) {
      const id = setInterval(() => {
        const newTo = Date.now()
        const newFrom = newTo - duration
        updateTimeRange(newFrom, newTo, false)
      }, interval)

      return () => {
        clearInterval(id)
      }
    }
  }, [updateTimeRange, autoReload, duration, interval])

  const handleTimeRangeClick = (t: number) => () => {
    const to = Date.now()
    const from = to - t
    updateTimeRange(from, to, false)
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
  const objectiveTypeLatency =
    objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative

  const loading: boolean =
    totalStatus === 'loading' ||
    totalStatus === 'idle' ||
    errorStatus === 'loading' ||
    errorStatus === 'idle'

  const success: boolean = totalStatus === 'success' && errorStatus === 'success'

  let errors: number = 0
  let total: number = 1
  if (totalResponse?.options.case === 'vector' && errorResponse?.options.case === 'vector') {
    if (errorResponse.options.value.samples.length > 0) {
      errors = errorResponse.options.value.samples[0].value
    }

    if (totalResponse.options.value.samples.length > 0) {
      total = totalResponse.options.value.samples[0].value
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
              <Tiles>
                <ObjectiveTile objective={objective} />
                <AvailabilityTile
                  objective={objective}
                  loading={loading}
                  success={success}
                  errors={errors}
                  total={total}
                />
                <ErrorBudgetTile
                  objective={objective}
                  loading={loading}
                  success={success}
                  errors={errors}
                  total={total}
                />
              </Tiles>
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
                  updateTimeRange={updateTimeRangeSelect}
                />
              ) : (
                <></>
              )}
            </Col>
          </Row>
          <Row>
            <Col
              xs={12}
              md={objectiveTypeLatency ? 12 : 6}
              className={objectiveTypeLatency ? 'col-xxxl-4' : ''}>
              {objective.queries?.graphRequests !== undefined ? (
                <RequestsGraph
                  client={promClient}
                  query={replaceInterval(objective.queries.graphRequests, from, to)}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  type={objectiveType}
                  updateTimeRange={updateTimeRangeSelect}
                />
              ) : (
                <></>
              )}
            </Col>
            <Col
              xs={12}
              md={objectiveTypeLatency ? 12 : 6}
              className={objectiveTypeLatency ? 'col-xxxl-4' : ''}>
              {objective.queries?.graphErrors !== undefined ? (
                <ErrorsGraph
                  client={promClient}
                  type={objectiveType}
                  query={replaceInterval(objective.queries.graphErrors, from, to)}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  updateTimeRange={updateTimeRangeSelect}
                />
              ) : (
                <></>
              )}
            </Col>
            {objectiveTypeLatency && (
              <Col xs={12} className="col-xxxl-4">
                <DurationGraph
                  client={client}
                  labels={labels}
                  grouping={groupingLabels}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  updateTimeRange={updateTimeRangeSelect}
                  target={objective.target}
                  latency={latencyTarget(objective)}
                />
              </Col>
            )}
          </Row>
          <Row>
            <Col>
              <h4>Multi Burn Rate Alerts</h4>
              <AlertsTable
                client={client}
                promClient={promClient}
                objective={objective}
                grouping={groupingLabels}
                from={from}
                to={to}
                uPlotCursor={uPlotCursor}
              />
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
