import {ConnectError, PromiseClient} from '@bufbuild/connect-web'
import {QueryStatus} from 'react-query/types/core/types'
import {ObjectiveService} from './proto/objectives/v1alpha1/objectives_connectweb'
import {Timestamp} from '@bufbuild/protobuf'
import {QueryOptions, useConnectQuery} from './query'
import {GetStatusResponse, ListResponse} from './proto/objectives/v1alpha1/objectives_pb'

export interface ObjectivesListResponse {
  response: ListResponse | null
  error: ConnectError
  status: QueryStatus
}

export const useObjectivesList = (
  client: PromiseClient<typeof ObjectiveService>,
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
  client: PromiseClient<typeof ObjectiveService>,
  expr: string,
  grouping: string,
  to: number,
  options?: QueryOptions,
): ObjectivesQueryResponse => {
  const {data, error, status} = useConnectQuery({
    key: ['status', expr, grouping],
    func: async () => {
      return await client.getStatus({expr, grouping, time: Timestamp.fromDate(new Date(to))})
    },
    options,
  })

  return {response: data ?? null, error: error as ConnectError, status}
}
