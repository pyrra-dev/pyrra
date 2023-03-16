import {convertAlignedData, mergeAlignedData} from './aligneddata'
import {QueryRangeResponse} from '../../proto/prometheus/v1/prometheus_pb'

describe('convertAlignedData', () => {
  it('should convert null into empty alignedData', () => {
    expect(convertAlignedData(null)).toEqual({labels: [], data: []})
  })
  it('should convert vector responses into empty alignedData', function () {
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"vector":{"samples":[{"time":"1676673386","value":2058,"metric":{"job":"parca-load"}}]}}',
        ),
      ),
    ).toEqual({labels: [], data: []})
  })
  it('should convert responses with no series into alignedData', () => {
    expect(convertAlignedData(QueryRangeResponse.fromJsonString('{}'))).toEqual({
      labels: [],
      data: [],
    })
  })
  it('should convert responses with single series into alignedData', () => {
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"matrix":{"samples":[{"values":[],"metric":{"job":"pyrra"}}]}}',
        ),
      ),
    ).toEqual({
      labels: [{job: 'pyrra'}],
      data: [[], []],
    })
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"matrix":{"samples":[{"values":[{"time":"1","value":100},{"time":"2","value":200}],"metric":{"job":"pyrra"}}]}}',
        ),
      ),
    ).toEqual({
      labels: [{job: 'pyrra'}],
      data: [
        [1, 2],
        [100, 200],
      ],
    })
  })
  it('should convert responses with single series and NaNs into alignedData', () => {
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"matrix":{"samples":[{"values":[{"time":"1","value":100},{"time":"2","value":"NaN"}],"metric":{"job":"pyrra"}}]}}',
        ),
      ),
    ).toEqual({
      labels: [{job: 'pyrra'}],
      data: [
        [1, 2],
        [100, null],
      ],
    })
  })
  it('should convert responses with two series into alignedData', () => {
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"matrix":{"samples":[{"values":[],"metric":{"job":"pyrra"}},{"values":[],"metric":{"job":"parca"}}]}}',
        ),
      ),
    ).toEqual({
      labels: [{job: 'pyrra'}, {job: 'parca'}],
      data: [[], [], []],
    })
  })
  expect(
    convertAlignedData(
      QueryRangeResponse.fromJsonString(
        '{"matrix": {"samples":[{"values":[{"time":"1","value":100},{"time":"2","value":200}],"metric":{"job":"pyrra"}},{"values":[{"time":"1","value":200},{"time":"2","value":400}],"metric":{"job":"parca"}}]}}',
      ),
    ),
  ).toEqual({
    labels: [{job: 'pyrra'}, {job: 'parca'}],
    data: [
      [1, 2],
      [100, 200],
      [200, 400],
    ],
  })
  it('should convert responses with multiple series and misaligned timestamps into alignedData', () => {
    expect(
      convertAlignedData(
        QueryRangeResponse.fromJsonString(
          '{"matrix":{"samples":[{"values":[{"time":"1","value":100},{"time":"2","value":200},{"time":"3","value":300}],"metric":{"job":"pyrra"}},{"values":[{"time":"2","value":200},{"time":"3","value":400}],"metric":{"job":"parca"}}]}}',
        ),
      ),
    ).toEqual({
      labels: [{job: 'pyrra'}, {job: 'parca'}],
      data: [
        [1, 2, 3],
        [100, 200, 300],
        [null, 200, 400],
      ],
    })
  })
})

describe('mergeAlignedData', () => {
  it('should convert null into empty alignedData', () => {
    expect(mergeAlignedData([])).toEqual({labels: [], data: []})
  })
  it('should merge by returning the single input', () => {
    expect(
      mergeAlignedData([
        {
          labels: [{job: 'pyrra'}],
          data: [
            [1, 2, 3],
            [100, 200, 300],
          ],
        },
      ]),
    ).toEqual({
      labels: [{job: 'pyrra'}],
      data: [
        [1, 2, 3],
        [100, 200, 300],
      ],
    })
  })
  it('should merge two aligned inputs', () => {
    expect(
      mergeAlignedData([
        {
          labels: [{job: 'pyrra'}],
          data: [
            [1, 2, 3],
            [100, 200, 300],
          ],
        },
        {
          labels: [{job: 'parca'}],
          data: [
            [1, 2, 3],
            [1000, 2000, 3000],
          ],
        },
      ]),
    ).toEqual({
      labels: [{job: 'pyrra'}, {job: 'parca'}],
      data: [
        [1, 2, 3],
        [100, 200, 300],
        [1000, 2000, 3000],
      ],
    })
  })
  it('should merge misaligned inputs', () => {
    expect(
      mergeAlignedData([
        {
          labels: [{job: 'pyrra'}],
          data: [
            [2, 3, 4],
            [200, 300, 400],
          ],
        },
        {
          labels: [{job: 'parca'}],
          data: [
            [1, 2, 3, 4, 5],
            [1000, 2000, null, 4000, null],
          ],
        },
      ]),
    ).toEqual({
      labels: [{job: 'pyrra'}, {job: 'parca'}],
      data: [
        [1, 2, 3, 4, 5],
        [null, 200, 300, 400, null],
        [1000, 2000, null, 4000, null],
      ],
    })
  })
})
