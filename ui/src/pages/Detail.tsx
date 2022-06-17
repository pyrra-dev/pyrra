import {Link, useLocation, useNavigate} from 'react-router-dom'
import React, {useEffect, useMemo, useState} from 'react'
import {Badge, Button, ButtonGroup, Col, Container, Row, Spinner} from 'react-bootstrap'
import {
  Configuration,
  Objective,
  ObjectivesApi,
  ObjectiveStatus,
  ObjectiveStatusAvailability,
  ObjectiveStatusBudget
} from '../client'
import {API_BASEPATH, formatDuration, parseDuration} from '../App'
import Navbar from '../components/Navbar'
import {MetricName, parseLabels} from "../labels";
import ErrorBudgetGraph from "../components/graphs/ErrorBudgetGraph";
import RequestsGraph from "../components/graphs/RequestsGraph";
import ErrorsGraph from "../components/graphs/ErrorsGraph";
import AlertsTable from "../components/AlertsTable";

const Detail = () => {
  const api = useMemo(() => {
    return new ObjectivesApi(new Configuration({ basePath: API_BASEPATH }))
  }, []);

  const navigate = useNavigate()
  const {search} = useLocation()

  const {
    timeRange,
    expr,
    grouping,
    groupingExpr,
    groupingLabels,
    name,
    labels,
  } = useMemo(() => {
    const query = new URLSearchParams(search)

    const queryExpr = query.get('expr')
    const expr = queryExpr == null ? '' : queryExpr
    const labels = parseLabels(expr)

    const groupingExpr = query.get('grouping')
    const grouping = groupingExpr == null ? '' : groupingExpr
    const groupingLabels = parseLabels(grouping)

    const name: string = labels[MetricName]

    const timeRangeQuery = query.get('timerange')
    const timeRangeParsed = timeRangeQuery != null ? parseDuration(timeRangeQuery) : null
    const timeRange: number = timeRangeParsed != null ? timeRangeParsed : 3600 * 1000

    document.title = `${name} - Pyrra`

    return { timeRange, expr, grouping, groupingExpr, groupingLabels, name, labels };
  }, [search]);

  const [objective, setObjective] = useState<Objective | null>(null);
  const [objectiveError, setObjectiveError] = useState<string>('');

  enum StatusState {
    Unknown,
    Error,
    NoData,
    Success,
  }

  const [statusState, setStatusState] = useState<StatusState>(StatusState.Unknown);
  const [availability, setAvailability] = useState<ObjectiveStatusAvailability | null>(null);
  const [errorBudget, setErrorBudget] = useState<ObjectiveStatusBudget | null>(null);

  useEffect(() => {
    // const controller = new AbortController()

    api.listObjectives({ expr: expr })
      .then((os: Objective[]) => {
        if (os.length === 1) {
          if (os[0].config === objective?.config) {
            // Prevent the setState if the objective is the same
            return;
          }
          setObjective(os[0])
        } else {
          setObjective(null)
        }
      })
      .catch((resp) => {
        if (resp.status !== undefined) {
          resp.text().then((err: string) => (setObjectiveError(err)))
        } else {
          setObjectiveError(resp.message)
        }
      })

    api.getObjectiveStatus({ expr: expr, grouping: grouping })
      .then((s: ObjectiveStatus[]) => {
        if (s.length === 0) {
          setStatusState(StatusState.NoData)
        } else if (s.length === 1) {
          setAvailability(s[0].availability)
          setErrorBudget(s[0].budget)
          setStatusState(StatusState.Success)
        } else {
          setStatusState(StatusState.Error)
        }
      })
      .catch((resp) => {
        if (resp.status === 404) {
          setStatusState(StatusState.NoData)
        } else {
          setStatusState(StatusState.Error)
        }
      })

    // return () => {
    //     // cancel any pending requests.
    //     controller.abort()
    // }
  }, [api, name, expr, grouping, timeRange, StatusState.Error, StatusState.NoData, StatusState.Success, objective?.config])

  if (objectiveError !== '') {
    return (
      <>
        <Navbar/>
        <Container>
          <div className='header'>
            <h3>{objectiveError}</h3>
            <br/>
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
      <div style={{ marginTop: '50px', textAlign: 'center' }}>
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading...</span>
        </Spinner>
      </div>
    )
  }

  if (objective.labels === undefined) {
    return (
      <></>
    )
  }

  const timeRanges = [
    28 * 24 * 3600 * 1000, // 4w
    7 * 24 * 3600 * 1000, // 1w
    24 * 3600 * 1000, // 1d
    12 * 3600 * 1000, // 12h
    3600 * 1000 // 1h
  ]

  const handleTimeRangeClick = (t: number) => () => {
    navigate(`/objectives?expr=${expr}&grouping=${groupingExpr ?? ''}&timerange=${formatDuration(t)}`)
  }

  const renderAvailability = () => {
    const headline = (<h6>Availability</h6>)
    switch (statusState) {
      case StatusState.Unknown:
        return (
          <div>
            {headline}
            <Spinner animation={'border'} style={{
              width: 50,
              height: 50,
              padding: 0,
              borderRadius: 50,
              borderWidth: 2,
              opacity: 0.25
            }}/>
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
        if (availability === null) {
          return <></>
        }
        return (
          <div className={availability.percentage > objective.target ? 'good' : 'bad'}>
            {headline}
            <h2>{(100 * availability.percentage).toFixed(3)}%</h2>
          </div>
        )
    }
  }

  const renderErrorBudget = () => {
    const headline = (<h6>Error Budget</h6>)
    switch (statusState) {
      case StatusState.Unknown:
        return (
          <div>
            {headline}
            <Spinner animation={'border'} style={{
              width: 50,
              height: 50,
              padding: 0,
              borderRadius: 50,
              borderWidth: 2,
              opacity: 0.25
            }}/>
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
        if (errorBudget === null) {
          return <></>
        }
        return (
          <div className={errorBudget.remaining > 0 ? 'good' : 'bad'}>
            {headline}
            <h2>{(100 * errorBudget.remaining).toFixed(3)}%</h2>
          </div>
        )
    }
  }

  const labelBadges = Object.entries({ ...objective.labels, ...groupingLabels })
    .filter((l: [string, string]) => l[0] !== MetricName)
    .map((l: [string, string]) => (
      <Badge key={l[1]} bg='light' text='dark' className="fw-normal">{l[0]}={l[1]}</Badge>
    ))

  const uPlotCursor = {
    lock: true,
    sync: {
      key: 'detail'
    }
  };

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
            <Col xs={12} className='header'>
              <h3>{name}</h3>
              {labelBadges}
            </Col>
            {objective.description !== undefined && objective.description !== '' ? (
                <Col xs={12} md={6} style={{ marginTop: 12 }}>
                  <p>{objective.description}</p>
                </Col>
              )
              : (<></>)}
          </Row>
          <Row>
            <div className="metrics">
              <div>
                <h6>Objective in <strong>{formatDuration(objective.window)}</strong></h6>
                <h2>{(100 * objective.target).toFixed(3)}%</h2>
              </div>

              {renderAvailability()}

              {renderErrorBudget()}

            </div>
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
                      active={timeRange === t}
                    >{formatDuration(t)}</Button>
                  ))}
                </ButtonGroup>
              </div>
            </Col>
          </Row>
          <Row style={{ marginBottom: 0 }}>
            <Col>
              <ErrorBudgetGraph
                api={api}
                labels={labels}
                grouping={groupingLabels}
                timeRange={timeRange}
                uPlotCursor={uPlotCursor}
              />
            </Col>
          </Row>
          <Row>
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
            <Col xs={12} md={6}>
              <RequestsGraph
                api={api}
                labels={labels}
                grouping={groupingLabels}
                timeRange={timeRange}
                uPlotCursor={uPlotCursor}
              />
            </Col>
            <Col xs={12} md={6}>
              <ErrorsGraph
                api={api}
                labels={labels}
                grouping={groupingLabels}
                timeRange={timeRange}
                uPlotCursor={uPlotCursor}
              />
            </Col>
          </Row>
          <Row>
            <Col>
              <h4>Multi Burn Rate Alerts</h4>
              <AlertsTable
                api={api}
                objective={objective}
                grouping={groupingLabels}
              />
            </Col>
          </Row>
          <Row>
            <Col>
              <h4>Config</h4>
              <pre style={{ padding: 20, borderRadius: 4 }}>
                <code>{objective.config}</code>
              </pre>
            </Col>
          </Row>
        </Container>
      </div>
    </>
  );
};

export default Detail
