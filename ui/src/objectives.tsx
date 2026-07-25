import {type ConnectError, type Client} from '@connectrpc/connect'
import {type QueryStatus} from '@tanstack/react-query'
import {type GetStatusResponse, type ListResponse, type ObjectiveService, type Timeseries} from './proto/objectives/v1alpha1/objectives_pb'
import {timestampFromDate} from '@bufbuild/protobuf/wkt'
import {type QueryOptions, useConnectQuery} from './query'

export interface ObjectivesListResponse {
  response: ListResponse | null
  error: ConnectError
  status: QueryStatus
}

export const useObjectivesList = (
  client: Client<typeof ObjectiveService>,
  expr: string,
  grouping: string,
  options?: QueryOptions,
): ObjectivesListResponse => {
  const {data, error, status} = useConnectQuery({
    key: [expr, grouping],
    func: async () => {
      return await client.list({expr, grouping})
    },
    options,
  })

  return {response: data ?? null, error: error as ConnectError, status}
}

export interface ObjectivesQueryResponse {
  response: GetStatusResponse | null
  error: ConnectError
  status: QueryStatus
}

// TODO: Probably not needed anymore with PrometheusService's existence now.
export const useObjectivesStatus = (
  client: Client<typeof ObjectiveService>,
  expr: string,
  grouping: string,
  to: number,
  options?: QueryOptions,
): ObjectivesQueryResponse => {
  const {data, error, status} = useConnectQuery({
    key: ['status', expr, grouping],
    func: async () => {
      return await client.getStatus({expr, grouping, time: timestampFromDate(new Date(to))})
    },
    options,
  })

  return {response: data ?? null, error: error as ConnectError, status}
}

export interface DurationTimeseriesResponse {
  timeseries: Timeseries[]
  error: ConnectError
  status: QueryStatus
}

// useGraphDuration fetches the duration percentiles of a stored SLO, looked up
// by the same expr the list and detail views use.
export const useGraphDuration = (
  client: Client<typeof ObjectiveService>,
  expr: string,
  grouping: string,
  from: number,
  to: number,
  options?: QueryOptions,
): DurationTimeseriesResponse => {
  const {data, error, status} = useConnectQuery({
    key: ['graphDuration', expr, grouping, from, to],
    func: async () => {
      return await client.graphDuration({
        expr,
        grouping,
        start: timestampFromDate(new Date(from)),
        end: timestampFromDate(new Date(to)),
      })
    },
    options,
  })

  return {timeseries: data?.timeseries ?? [], error: error as ConnectError, status}
}

// usePreviewGraphDuration fetches the same percentiles for a draft SLO that
// isn't stored yet, so the Create editor's preview can render the graph too.
export const usePreviewGraphDuration = (
  client: Client<typeof ObjectiveService>,
  config: string,
  grouping: string,
  from: number,
  to: number,
  options?: QueryOptions,
): DurationTimeseriesResponse => {
  const {data, error, status} = useConnectQuery({
    key: ['previewGraphDuration', config, grouping, from, to],
    func: async () => {
      return await client.previewGraphDuration({
        config,
        grouping,
        start: timestampFromDate(new Date(from)),
        end: timestampFromDate(new Date(to)),
      })
    },
    options,
  })

  return {timeseries: data?.timeseries ?? [], error: error as ConnectError, status}
}
