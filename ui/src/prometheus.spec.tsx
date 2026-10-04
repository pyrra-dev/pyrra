import {replaceInterval} from './prometheus'

describe('replaceInterval', () => {
  it('should parse a request graph query', () => {
    expect(
      replaceInterval(
        'sum(rate(grpc_server_handling_seconds_count{grpc_method="ProfileTypes",grpc_service="parca.query.v1alpha1.QueryService"}[1s]))',
        0,
        60 * 60 * 1000,
      ),
    ).toEqual(
      'sum(rate(grpc_server_handling_seconds_count{grpc_method="ProfileTypes",grpc_service="parca.query.v1alpha1.QueryService"}[5m]))',
    )
  })
  it('should parse a errors graph query', () => {
    expect(
      replaceInterval(
        '(sum(rate(caddy_http_response_duration_seconds_count{code!~"5..",handler="subroute",job="caddy"}[1s])) - sum(rate(caddy_http_response_duration_seconds_bucket{code!~"5..",handler="subroute",job="caddy",le="0.05"}[1s]))) / sum(rate(caddy_http_response_duration_seconds_count{code!~"5..",handler="subroute",job="caddy"}[1s]))',
        0,
        14 * 24 * 60 * 60 * 1000,
      ),
    ).toEqual(
      '(sum(rate(caddy_http_response_duration_seconds_count{code!~"5..",handler="subroute",job="caddy"}[1h])) - sum(rate(caddy_http_response_duration_seconds_bucket{code!~"5..",handler="subroute",job="caddy",le="0.05"}[1h]))) / sum(rate(caddy_http_response_duration_seconds_count{code!~"5..",handler="subroute",job="caddy"}[1h]))',
    )
  })
  it('should scale the bool gauge per-second divisor with the interval', () => {
    expect(
      replaceInterval(
        'sum by (instance) (count_over_time(probe_success{job="blackbox"}[1s])) / 1',
        0,
        60 * 60 * 1000,
      ),
    ).toEqual('sum by (instance) (count_over_time(probe_success{job="blackbox"}[5m])) / 300')
  })
  it('should not touch a division by a number starting with 1', () => {
    expect(replaceInterval('sum(count_over_time(up[1s])) / 10', 0, 60 * 60 * 1000)).toEqual(
      'sum(count_over_time(up[5m])) / 10',
    )
  })
})
