import {parseLabels} from './labels';

describe('parseLabels', () => {
  it('parses metric labels into an object', () => {
    expect(parseLabels('{__name__="test-metric", namespace="test", team="testers"}')).toEqual({
      __name__: "test-metric",
      namespace: "test",
      team: "testers"
    })
  })
  it ('can handle empty strings', () => {
    expect(parseLabels('')).toEqual({})
    expect(parseLabels(null)).toEqual({})
    expect(parseLabels('{}')).toEqual({})
  })

  it('can handle special characters in label values', () => {
    expect(parseLabels('{comma_in_label="test,tester"}')).toEqual({
      comma_in_label: "test,tester"
    })
    expect(parseLabels('{equal_sign_in_label="test=tester"}')).toEqual({
      equal_sign_in_label: "test=tester"
    })
  })
})
