import Navbar from '../components/Navbar'
import React, {useMemo, useState} from 'react'
import {Button, Col, Container, Form, InputGroup, Row, Table} from 'react-bootstrap'
import {useForm} from 'react-hook-form'

import AvailabilityTile from '../components/tiles/AvailabilityTile'
import ObjectiveTile from '../components/tiles/ObjectiveTile'
import Tiles from '../components/tiles/Tiles'
import ErrorBudgetTile from '../components/tiles/ErrorBudgetTile'
import {createConnectTransport, createPromiseClient} from '@bufbuild/connect-web'
import {API_BASEPATH, ObjectiveType} from '../App'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_connectweb'
import {QueryResponse, Sample} from '../proto/prometheus/v1/prometheus_pb'
import {labelsString, MetricName, removeMetricName} from '../labels'
import ErrorBudgetGraph from '../components/graphs/ErrorBudgetGraph'
import RequestsGraph from '../components/graphs/RequestsGraph'
import ErrorsGraph from '../components/graphs/ErrorsGraph'
import uPlot from 'uplot'
import {parseDuration} from '../duration'
import {Objective} from '../crd/servicelevelobjectives'
import {stringify} from 'yaml'

const formFieldName = 'name'
const formFieldSLITotal = 'sli-total'
const formFieldSLIErrors = 'sli-error'
const formFieldObjective = 'objective'
const formFieldWindow = 'window'

const Create = () => {
  const baseUrl = API_BASEPATH === undefined ? 'http://localhost:9099' : API_BASEPATH

  const promClient = useMemo(() => {
    return createPromiseClient(PrometheusService, createConnectTransport({baseUrl}))
  }, [baseUrl])

  const [objectiveTotal, setObjectiveTotal] = useState<number | undefined>()
  const [objectiveErrors, setObjectiveErrors] = useState<number | undefined>()
  const [objectiveTotalSamples, setObjectiveTotalSamples] = useState<Sample[]>([])
  const [objectiveErrorsSamples, setObjectiveErrorsSamples] = useState<Sample[]>([])
  const [enableGraphs, setEnableGraphs] = useState<boolean>(false)

  const {
    register,
    handleSubmit,
    getValues,
    watch,
    formState: {errors, touchedFields, isValid},
  } = useForm({
    mode: 'onTouched',
    reValidateMode: 'onChange',
    shouldFocusError: true,
  })

  const now = Date.now()
  const from = now - 24 * 60 * 60 * 1000

  const queryTotal = () => {
    promClient
      .query({
        query: increaseQuery(getValues(formFieldSLITotal), getValues(formFieldWindow)),
        time: BigInt(Math.floor(now / 1000)),
      })
      .then((resp: QueryResponse) => {
        if (resp.options.case === 'vector') {
          setObjectiveTotalSamples(resp.options.value.samples)
          setObjectiveTotal(
            // sum up all values for the total number
            resp.options.value.samples
              .map((s: Sample): number => s.value)
              .reduce(
                (previousValue: number, currentValue: number): number =>
                  previousValue + currentValue,
              ),
          )
        }
      })
      .catch((err) => console.warn(err))
  }

  const queryErrors = () => {
    promClient
      .query({
        query: increaseQuery(getValues(formFieldSLIErrors), getValues(formFieldWindow)),
        time: BigInt(Math.floor(now / 1000)),
      })
      .then((resp: QueryResponse) => {
        if (resp.options.case === 'vector') {
          if (resp.options.value.samples.length === 0) {
            setObjectiveErrorsSamples([])
            setObjectiveErrors(0)
            return
          }

          setObjectiveErrorsSamples(resp.options.value.samples)
          setObjectiveErrors(
            // sum up all values for the total error number
            resp.options.value.samples
              .map((s: Sample): number => s.value)
              .reduce(
                (previousValue: number, currentValue: number): number =>
                  previousValue + currentValue,
              ),
          )
        }
      })
      .catch((err) => console.log(err))
  }

  const uPlotCursor: uPlot.Cursor = {
    y: false,
    lock: true,
    sync: {
      key: 'detail',
    },
  }

  // TODO: Validate against generated types
  const objective: Objective = {
    apiVersion: 'pyrra.dev/v1alpha1',
    kind: 'ServiceLevelObjective',
    metadata: {
      name: getValues(formFieldName),
      namespace: 'default',
    },
    spec: {
      description: getValues('description'),
      target: getValues(formFieldObjective),
      window: getValues(formFieldWindow),
      indicator: {
        ratio: {
          errors: {
            metric: getValues(formFieldSLIErrors),
          },
          total: {
            metric: getValues(formFieldSLITotal),
          },
        },
      },
    },
  }

  return (
    <>
      <Navbar />
      <Container className="content list">
        <Row>
          <Col>
            <h3>Create a SLO</h3>
          </Col>
        </Row>
        <Row>
          <Col>
            <Form
              onSubmit={handleSubmit((data) => {
                console.log('submit data', data)
              })}
              validated={isValid}>
              <Form.Group as={Col} controlId="name">
                <Form.Label>Name</Form.Label>
                <InputGroup hasValidation>
                  <Form.Control
                    type="text"
                    placeholder="prometheus-http-errors"
                    {...register(formFieldName, {required: 'You must provide a name for the SLO.'})}
                    required
                  />
                  <Form.Control.Feedback type="invalid">
                    Please choose a name for your SLO.
                  </Form.Control.Feedback>
                </InputGroup>
                <Form.Text className="text-muted">
                  Give this SLO a name that has meaning for others on your team.
                </Form.Text>
              </Form.Group>

              <Form.Group className="mb-3" controlId="description">
                <Form.Label>Description</Form.Label>
                <Form.Control
                  as="textarea"
                  type="password"
                  placeholder="Describe in a few sentences the impact that this SLOs firing might have."
                  {...register('description')}
                />
              </Form.Group>
              <Row>
                <Col>
                  <Form.Group controlId="objective">
                    <Form.Label>Objective</Form.Label>
                    <Form.Control
                      type="number"
                      min={0}
                      max={100}
                      step={0.001}
                      required
                      placeholder="99"
                      {...register(formFieldObjective)}
                    />
                  </Form.Group>
                </Col>
                <Col>
                  <Form.Group>
                    <Form.Label>Window</Form.Label>
                    <InputGroup>
                      <Form.Control
                        placeholder="4w"
                        isValid={
                          errors[formFieldWindow] === undefined && touchedFields[formFieldWindow]
                        }
                        isInvalid={errors[formFieldWindow]?.message !== undefined}
                        {...register(formFieldWindow, {
                          required: true,
                          pattern: /^([0-9]+)[y|w|d|h|m|s|ms]$/,
                        })}
                      />
                    </InputGroup>
                  </Form.Group>
                </Col>
              </Row>
              <br />
              <Row>
                <Col>
                  <h4>Indicator</h4>
                </Col>
              </Row>
              <Row>
                <Form.Group controlId="objective">
                  <Form.Label>
                    <h5>Total</h5>
                  </Form.Label>
                  <InputGroup>
                    <Form.Control
                      placeholder='http_requests_total{handler=~"/api.*"}'
                      required
                      {...register(formFieldSLITotal)}
                      onSubmit={() => console.log('query the shit')}
                    />
                    <Button
                      variant={touchedFields[formFieldSLITotal] !== undefined ? 'light' : 'primary'}
                      onClick={queryTotal}>
                      Query
                    </Button>
                  </InputGroup>
                  <Form.Text className="text-muted">
                    Summing up the values of all selected series results in the total number of
                    requests in 4w.
                  </Form.Text>
                  <Table hover={true}>
                    <tbody>
                      {objectiveTotalSamples
                        .sort((a: Sample, b: Sample) => b.value - a.value)
                        .map((s: Sample) => (
                          <tr>
                            <td>
                              <small>
                                {s.metric.__name__}
                                {labelsString(removeMetricName(s.metric))}
                              </small>
                            </td>
                            <td>{Math.floor(s.value)}</td>
                          </tr>
                        ))}
                      {objectiveTotalSamples.length > 0 ? (
                        <tr>
                          <td>Total</td>
                          <td>
                            <strong>{Math.floor(objectiveTotal ?? 0)}</strong>
                          </td>
                        </tr>
                      ) : (
                        <></>
                      )}
                    </tbody>
                  </Table>
                </Form.Group>
              </Row>
              <Row>
                <Form.Group controlId="objective">
                  <Form.Label>Errors</Form.Label>
                  <InputGroup>
                    <Form.Control
                      placeholder='http_requests_total{handler=~"/api.*",code=~"5.."}'
                      required
                      {...register(formFieldSLIErrors)}
                    />
                    <Button
                      variant={
                        touchedFields[formFieldSLIErrors] !== undefined ? 'light' : 'primary'
                      }
                      onClick={queryErrors}
                      disabled={
                        !(
                          touchedFields[formFieldSLIErrors] !== undefined &&
                          touchedFields[formFieldWindow] !== undefined &&
                          errors[formFieldSLIErrors] === undefined &&
                          errors[formFieldWindow] === undefined
                        )
                      }>
                      Query
                    </Button>
                  </InputGroup>
                  <Form.Text className="text-muted">
                    Summing up the values of all selected series results in the total number of
                    errors in 4w.
                  </Form.Text>
                  <Table hover={true}>
                    <tbody>
                      {objectiveErrorsSamples
                        .sort((a: Sample, b: Sample) => b.value - a.value)
                        .map((s: Sample) => (
                          <tr>
                            <td>
                              <small>
                                {s.metric[MetricName]}
                                {labelsString(removeMetricName(s.metric))}
                              </small>
                            </td>
                            <td>{Math.floor(s.value)}</td>
                          </tr>
                        ))}
                      {objectiveErrorsSamples.length > 0 ? (
                        <tr>
                          <td>Total</td>
                          <td>
                            <strong>{Math.floor(objectiveErrors ?? 0)}</strong>
                          </td>
                        </tr>
                      ) : (
                        <></>
                      )}
                    </tbody>
                  </Table>
                </Form.Group>
              </Row>
              <br />
              <br />
              <Row>
                <Col></Col>
              </Row>
              <br />
              <Row>
                <Col>
                  <Tiles>
                    {isValid && (
                      <ObjectiveTile
                        window={(parseDuration(watch(formFieldWindow) ?? '') ?? 0) / 1000}
                        target={watch(formFieldObjective) / 100}
                        objectiveType={ObjectiveType.Ratio} // TODO
                      />
                    )}
                    {objectiveErrors !== undefined && objectiveTotal !== undefined ? (
                      <>
                        <AvailabilityTile
                          target={watch(formFieldObjective) / 100}
                          objectiveType={ObjectiveType.Ratio} // TODO
                          total={{count: objectiveTotal, status: 'success'}}
                          errors={{count: objectiveErrors, status: 'success'}}
                        />
                        <ErrorBudgetTile
                          target={watch(formFieldObjective) / 100}
                          total={{count: objectiveTotal, status: 'success'}}
                          errors={{count: objectiveErrors, status: 'success'}}
                        />
                      </>
                    ) : (
                      <></>
                    )}
                  </Tiles>
                </Col>
              </Row>
              <Row>
                <Col></Col>
              </Row>
              <br />
              <Row>
                <Col>
                  <Button variant="light" onClick={() => setEnableGraphs(true)}>
                    Query Graphs
                  </Button>
                </Col>
                <Col></Col>
              </Row>
              {enableGraphs ? (
                <>
                  <Row>
                    <ErrorBudgetGraph
                      client={promClient}
                      from={from}
                      to={now}
                      query={watch(formFieldSLITotal)}
                      uPlotCursor={uPlotCursor}
                      updateTimeRange={() => {}}
                    />
                  </Row>
                  <Row>
                    <Col xs={12} md={6}>
                      <RequestsGraph
                        client={promClient}
                        query={`sum(rate(${String(watch(formFieldSLITotal))}[5m]))`}
                        from={from}
                        to={now}
                        type={ObjectiveType.Ratio}
                        uPlotCursor={uPlotCursor}
                        updateTimeRange={() => {}}
                      />
                    </Col>
                    <Col xs={12} md={6}>
                      <ErrorsGraph
                        client={promClient}
                        query={errorsRangeQuery(
                          watch(formFieldSLIErrors),
                          watch(formFieldSLITotal),
                        )}
                        from={from}
                        to={now}
                        type={ObjectiveType.Ratio} // TODO
                        uPlotCursor={uPlotCursor}
                        updateTimeRange={() => {}}
                      />
                    </Col>
                  </Row>
                </>
              ) : (
                <></>
              )}
            </Form>
          </Col>
        </Row>
        <Row>
          <Col>
            <h4>Config</h4>
            <pre style={{padding: 20, borderRadius: 4}}>
              <code>{stringify(objective, null, 2)}</code>
            </pre>
          </Col>
        </Row>
      </Container>
    </>
  )
}

export default Create

const increaseQuery = (matchers: string, window: string): string =>
  `increase(${matchers}[${window}]) > 0`

const errorsRangeQuery = (errors: string, total: string): string =>
  `sum(rate(${errors}[5m])) / sum(rate(${total}[5m]))`
