import { Link, RouteComponentProps, useHistory, useLocation } from 'react-router-dom'
import React, { useEffect, useMemo, useState } from 'react'
import {
  Configuration,
  Objective as APIObjective,
  ObjectivesApi,
  ObjectiveStatus as APIObjectiveStatus,
  ObjectiveStatusAvailability,
  ObjectiveStatusBudget
} from '../client'
import { Button, ButtonGroup, Col, Container, Row, Spinner } from 'react-bootstrap'
import { formatDuration, parseDuration, PUBLIC_API } from '../App'
import AlertsTable from '../components/AlertsTable'
import ErrorBudgetGraph from '../components/graphs/ErrorBudgetGraph'
import RequestsGraph from '../components/graphs/RequestsGraph'
import ErrorsGraph from '../components/graphs/ErrorsGraph'

interface DetailRouteParams {
  name: string
  namespace: string
}

const Detail = (params: RouteComponentProps<DetailRouteParams>) => {
  const { namespace, name } = params.match.params;

  const history = useHistory()
  const query = new URLSearchParams(useLocation().search)

  const api = useMemo(() => {
    return new ObjectivesApi(new Configuration({ basePath: `${PUBLIC_API}api/v1` }))
  }, [])

  const timeRangeQuery = query.get('timerange')
  const timeRangeParsed = timeRangeQuery != null ? parseDuration(timeRangeQuery) : null
  const timeRange: number = timeRangeParsed != null ? timeRangeParsed : 3600 * 1000

  const [objective, setObjective] = useState<APIObjective | null>(null);
  const [availability, setAvailability] = useState<ObjectiveStatusAvailability | null>(null);
  const [errorBudget, setErrorBudget] = useState<ObjectiveStatusBudget | null>(null);


  useEffect(() => {
    const controller = new AbortController()

    document.title = `${name} - Pyrra`

    api.getObjective({ namespace, name })
      .then((o: APIObjective) => setObjective(o))

    api.getObjectiveStatus({ namespace, name })
      .then((s: APIObjectiveStatus) => {
        setAvailability(s.availability)
        setErrorBudget(s.budget)
      })

    return () => {
      // cancel any pending requests.
      controller.abort()
    }
  }, [api, namespace, name, timeRange])

  if (objective == null) {
    return (
      <div style={{ marginTop: '50px', textAlign: 'center' }}>
        <Spinner animation="border" role="status">
          <span className="sr-only">Loading...</span>
        </Spinner>
      </div>
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
        <Row>
          <Col className="text-right">
            <ButtonGroup aria-label="Basic example">
              {timeRanges.map((t: number, i: number) => (
                <Button
                  key={i}
                  variant="light"
                  onClick={handleTimeRangeClick(t)}
                  active={timeRange === t}
                >{formatDuration(t)}</Button>
              ))}
            </ButtonGroup>
          </Col>
        </Row>
        <Row style={{ marginBottom: 0 }}>
          <Col>
            <ErrorBudgetGraph
              api={api}
              namespace={namespace}
              name={name}
              timeRange={timeRange}
            />
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
            <RequestsGraph
              api={api}
              namespace={namespace}
              name={name}
              timeRange={timeRange}
            />
          </Col>
          <Col xs={12} sm={6}>
            <ErrorsGraph
              api={api}
              namespace={namespace}
              name={name}
              timeRange={timeRange}
            />
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
