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
})
