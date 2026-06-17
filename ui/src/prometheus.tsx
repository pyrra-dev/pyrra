import {type PrometheusService, type QueryRangeResponse, type QueryResponse} from './proto/prometheus/v1/prometheus_pb'
import {type ConnectError, type Client} from '@connectrpc/connect'
import {type QueryStatus} from '@tanstack/react-query'
import {type QueryOptions, useConnectQuery} from './query'
import {formatDuration} from './duration'

export interface PrometheusQueryResponse {
  response: QueryResponse | null
  error: ConnectError
  status: QueryStatus
}

export const usePrometheusQuery = (
  client: Client<typeof PrometheusService>,
  query: string,
  time: number,
  options?: QueryOptions,
): PrometheusQueryResponse => {
  time = Math.floor(time)
  const {data, error, status} = useConnectQuery<QueryResponse>({
    key: ['query', query, time],
    func: async () => {
      return await client.query({query, time: BigInt(time)})
    },
    options,
  })

  return {response: data ?? null, error: error as ConnectError, status}
}

export interface PrometheusQueryRangeResponse {
  response: QueryRangeResponse | null
  error: ConnectError
  status: QueryStatus
}

export const usePrometheusQueryRange = (
  client: Client<typeof PrometheusService>,
  query: string,
  start: number,
  end: number,
  step: number,
  options?: QueryOptions,
): PrometheusQueryRangeResponse => {
  start = Math.floor(start)
  end = Math.floor(end)
  step = Math.floor(step) !== 0 ? Math.floor(step) : 1
  const {data, error, status} = useConnectQuery<QueryRangeResponse>({
    key: ['queryRange', query, start / 10, end / 10, step],
    func: async () => {
      return await client.queryRange({
        query,
        start: BigInt(start),
        end: BigInt(end),
        step: BigInt(step),
      })
    },
    options,
  })

  return {response: data ?? null, error: error as ConnectError, status}
}

export const replaceInterval = (query: string, from: number, to: number): string => {
  const duration: number = (to - from) / 1000

  const minute = 60
  const hour = 60 * minute
  const day = 24 * hour
  const week = 7 * day
  const month = 4 * week

  let rateInterval: number = 5 * 60
  if (duration >= month) {
    rateInterval = 3 * hour
  } else if (duration >= week) {
    rateInterval = hour
  } else if (duration >= day) {
    rateInterval = 30 * minute
  } else if (duration >= day / 2) {
    rateInterval = 15 * minute
  }

  const rateIntervalStr = formatDuration(rateInterval * 1000, 1)

  return query.replaceAll(/\[(1s)\]/g, `[${rateIntervalStr}]`)
}

// vectorErrorsTotal pulls the errors/total sample values out of two instant-vector
// query responses. total defaults to 1 so callers can divide without guarding.
export const vectorErrorsTotal = (
  totalResponse: QueryResponse | null,
  errorResponse: QueryResponse | null,
): {errors: number; total: number} => {
  let errors = 0
  let total = 1
  if (totalResponse?.options.case === 'vector' && errorResponse?.options.case === 'vector') {
    if (errorResponse.options.value.samples.length > 0) {
      errors = errorResponse.options.value.samples[0].value
    }
    if (totalResponse.options.value.samples.length > 0) {
      total = totalResponse.options.value.samples[0].value
    }
  }
  return {errors, total}
}
