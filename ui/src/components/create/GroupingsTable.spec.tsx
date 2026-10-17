import {create} from '@bufbuild/protobuf'
import {ConnectError} from '@connectrpc/connect'
import {QueryClient, QueryClientProvider} from '@tanstack/react-query'
import {cleanup, fireEvent, render, screen} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {ObjectiveSchema, QueriesSchema} from '../../proto/objectives/v1alpha1/objectives_pb'
import {
  QueryResponseSchema,
  SampleSchema,
  VectorSchema,
} from '../../proto/prometheus/v1/prometheus_pb'
import {usePrometheusQuery, type PrometheusQueryResponse} from '../../prometheus'
import GroupingsTable from './GroupingsTable'

vi.mock('../../prometheus', () => ({usePrometheusQuery: vi.fn()}))

const objective = create(ObjectiveSchema, {
  target: 0.99,
  queries: create(QueriesSchema, {countTotal: 'total-query', countErrors: 'errors-query'}),
})

const vector = (samples: Array<{metric: Record<string, string>; value: number}>) =>
  create(QueryResponseSchema, {
    options: {
      case: 'vector',
      value: create(VectorSchema, {
        samples: samples.map((sample) => create(SampleSchema, sample)),
      }),
    },
  })

const result = (
  response: PrometheusQueryResponse['response'],
  status: PrometheusQueryResponse['status'],
): PrometheusQueryResponse => ({
  response,
  status,
  error: status === 'error' ? new ConnectError('query failed') : (null as unknown as ConnectError),
})

let totalResult: PrometheusQueryResponse
let errorsResult: PrometheusQueryResponse

beforeEach(() => {
  totalResult = result(vector([{metric: {job: 'api'}, value: 100}]), 'success')
  errorsResult = result(vector([{metric: {job: 'api'}, value: 5}]), 'success')
  vi.mocked(usePrometheusQuery).mockImplementation((_client, query) =>
    query === 'total-query' ? totalResult : errorsResult,
  )
})

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  vi.clearAllMocks()
})

const renderTable = () => {
  const queryClient = new QueryClient()
  const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')
  render(
    <QueryClientProvider client={queryClient}>
      <GroupingsTable baseUrl="/" objective={objective} selected={null} onSelect={vi.fn()} />
    </QueryClientProvider>,
  )
  const queryTime = Math.floor(vi.mocked(usePrometheusQuery).mock.calls[0][2])
  return {invalidateQueries, queryTime}
}

describe('GroupingsTable query results', () => {
  it('shows an error and retries the errors query when only it fails', () => {
    const clock = vi.spyOn(Date, 'now').mockReturnValue(1_000_500)
    errorsResult = result(null, 'error')
    const {invalidateQueries, queryTime} = renderTable()

    expect(screen.getByText('Couldn’t load groupings')).toBeInTheDocument()
    expect(screen.getByText(/errors query failed/)).toBeInTheDocument()
    expect(screen.queryByText('100.000%')).not.toBeInTheDocument()

    clock.mockReturnValue(1_001_500)
    expect(queryTime).toBe(1000)
    expect(Math.floor(Date.now() / 1000)).toBe(1001)
    fireEvent.click(screen.getByRole('button', {name: 'Retry'}))
    expect(invalidateQueries).toHaveBeenCalledExactlyOnceWith({
      queryKey: ['query', 'errors-query', queryTime],
      exact: true,
    })
  })

  it('treats a successful empty errors vector as zero errors', () => {
    errorsResult = result(vector([]), 'success')
    renderTable()

    expect(screen.getByText('job=api')).toBeInTheDocument()
    expect(screen.getByText('100.000%')).toBeInTheDocument()
    expect(screen.getByText('0')).toBeInTheDocument()
  })

  it('shows an error and retries the total query when it fails', () => {
    totalResult = result(null, 'error')
    const {invalidateQueries, queryTime} = renderTable()

    expect(screen.getByText('Couldn’t load groupings')).toBeInTheDocument()
    expect(screen.getByText(/total query failed/)).toBeInTheDocument()
    expect(screen.queryByText('No groupings found')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', {name: 'Retry'}))
    expect(invalidateQueries).toHaveBeenCalledExactlyOnceWith({
      queryKey: ['query', 'total-query', queryTime],
      exact: true,
    })
  })

  it('keeps the loading state while queries are pending', () => {
    totalResult = result(null, 'pending')
    renderTable()

    expect(screen.getByText('Querying available groupings…')).toBeInTheDocument()
    expect(screen.queryByText('No groupings found')).not.toBeInTheDocument()
  })

  it('shows the empty state for a successful total query with no series', () => {
    totalResult = result(vector([]), 'success')
    errorsResult = result(vector([]), 'success')
    renderTable()

    expect(screen.getByText('No groupings found')).toBeInTheDocument()
  })
})
