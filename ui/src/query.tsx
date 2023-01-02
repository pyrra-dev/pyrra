import {useQuery, UseQueryResult} from 'react-query'

interface Props<ConnectResponse> {
  key: string | any[]
  func: () => Promise<ConnectResponse>
  options?: QueryOptions
}

export interface QueryOptions {
  enabled?: boolean
}

export const useConnectQuery = <ConnectResponse,>({
  key,
  func,
  options: {enabled = true} = {},
}: Props<ConnectResponse>): UseQueryResult<ConnectResponse> => {
  return useQuery<ConnectResponse>(
    key,
    async () => {
      return await func()
    },
    {enabled},
  )
}
