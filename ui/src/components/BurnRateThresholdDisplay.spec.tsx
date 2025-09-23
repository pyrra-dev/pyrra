import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { PromiseClient, ConnectError } from '@connectrpc/connect'
import { PrometheusService } from '../proto/prometheus/v1/prometheus_connect'
import { Objective } from '../proto/objectives/v1alpha1/objectives_pb'
import { QueryResponse, Vector, Sample, SamplePair } from '../proto/prometheus/v1/prometheus_pb'
import BurnRateThresholdDisplay from './BurnRateThresholdDisplay'
import { BurnRateType } from '../burnrate'

// Mock the prometheus hook
jest.mock('../prometheus', () => ({
  usePrometheusQuery: jest.fn()
}))

// Mock the burnrate module
jest.mock('../burnrate', () => ({
  getBurnRateType: jest.fn(),
  BurnRateType: {
    Static: 'static',
    Dynamic: 'dynamic'
  }
}))

import { usePrometheusQuery } from '../prometheus'
import { getBurnRateType } from '../burnrate'

const mockUsePrometheusQuery = usePrometheusQuery as jest.MockedFunction<typeof usePrometheusQuery>
const mockGetBurnRateType = getBurnRateType as jest.MockedFunction<typeof getBurnRateType>

// Mock Prometheus client
const mockPromClient = {} as PromiseClient<typeof PrometheusService>

describe('BurnRateThresholdDisplay', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Clear localStorage before each test
    localStorage.clear()
  })

  describe('Static Burn Rate', () => {
    it('displays calculated threshold for static burn rate', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Static)
      
      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-slo' }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      // Static threshold = factor * (1 - target) = 14 * (1 - 0.95) = 14 * 0.05 = 0.7
      expect(screen.getByText('0.70000')).toBeInTheDocument()
    })

    it('handles different factors correctly', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Static)
      
      const objective = new Objective({
        target: 0.99,
        labels: { __name__: 'test-slo' }
      })
      
      const { rerender } = render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={7} 
          promClient={mockPromClient} 
        />
      )
      
      // Static threshold = 7 * (1 - 0.99) = 7 * 0.01 = 0.07
      expect(screen.getByText('0.07000')).toBeInTheDocument()
      
      // Test different factor
      rerender(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={2} 
          promClient={mockPromClient} 
        />
      )
      
      // Static threshold = 2 * (1 - 0.99) = 2 * 0.01 = 0.02
      expect(screen.getByText('0.02000')).toBeInTheDocument()
    })
  })

  describe('Dynamic Burn Rate - Ratio Indicators', () => {
    const createRatioObjective = () => new Objective({
      target: 0.95,
      labels: { __name__: 'test-ratio-slo' },
      indicator: {
        options: {
          case: 'ratio',
          value: {
            total: { metric: 'http_requests_total' },
            errors: { metric: 'http_requests_total{status=~"5.."}' }
          }
        }
      }
    })

    it('displays loading state while calculating dynamic threshold', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'loading',
        error: {} as ConnectError
      })
      
      const objective = createRatioObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      expect(screen.getByText('Calculating...')).toBeInTheDocument()
      expect(screen.getByTitle('Calculating dynamic threshold...')).toBeInTheDocument()
    })

    it('displays calculated dynamic threshold for ratio indicators', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 2.5, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })
      
      const objective = createRatioObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        // Dynamic threshold = traffic_ratio * threshold_constant * (1 - target)
        // threshold_constant for factor 14 = 1/48 = 0.020833
        // Dynamic threshold = 2.5 * 0.020833 * (1 - 0.95) = 2.5 * 0.020833 * 0.05 = 0.00260
        expect(screen.getByText('0.00260')).toBeInTheDocument()
      })
    })

    it('handles scalar response format', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'scalar',
            value: new SamplePair({ value: 1.5, time: BigInt(0) })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })
      
      const objective = createRatioObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={7} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        // Dynamic threshold = 1.5 * (1/16) * (1 - 0.95) = 1.5 * 0.0625 * 0.05 = 0.00469
        expect(screen.getByText('0.00469')).toBeInTheDocument()
      })
    })
  })

  describe('Dynamic Burn Rate - LatencyNative Indicators', () => {
    const createLatencyNativeObjective = () => new Objective({
      target: 0.995,
      labels: { __name__: 'test-latency-native-slo' },
      indicator: {
        options: {
          case: 'latencyNative',
          value: {
            total: { metric: 'http_request_duration_native' },
            latency: '0.1' // 100ms threshold
          }
        }
      }
    })

    it('validates latencyNative indicator has total metric', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'loading',
        error: {} as ConnectError
      })
      
      const objective = new Objective({
        target: 0.995,
        labels: { __name__: 'test-latency-native-slo' },
        indicator: {
          options: {
            case: 'latencyNative',
            value: {
              total: { metric: '' }, // Missing total metric
              latency: '0.1'
            }
          }
        }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      expect(screen.getByText('Unable to calculate (see console)')).toBeInTheDocument()
      expect(screen.getByTitle('Missing native histogram metric for latency calculation')).toBeInTheDocument()
    })

    it('displays calculated dynamic threshold for latencyNative indicators', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1.8, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })
      
      const objective = createLatencyNativeObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={1} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        // Dynamic threshold = 1.8 * (1/7) * (1 - 0.995) = 1.8 * 0.142857 * 0.005 = 0.00129
        expect(screen.getByText('0.00129')).toBeInTheDocument()
      })
    })

    it('generates correct query for latencyNative indicators', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'loading',
        error: {} as ConnectError
      })
      
      const objective = createLatencyNativeObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      // Verify the query uses histogram_count() for native histograms
      expect(mockUsePrometheusQuery).toHaveBeenCalledWith(
        mockPromClient,
        expect.stringContaining('sum(histogram_count(increase(http_request_duration_native[30d]))) / sum(histogram_count(increase(http_request_duration_native[1h4m]))'),
        expect.any(Number),
        expect.any(Object)
      )
    })
  })

  describe('Dynamic Burn Rate - BoolGauge Indicators', () => {
    const createBoolGaugeObjective = () => new Objective({
      target: 0.98,
      labels: { __name__: 'test-bool-gauge-slo' },
      indicator: {
        options: {
          case: 'boolGauge',
          value: {
            boolGauge: { metric: 'service_up' }
          }
        }
      }
    })

    it('validates boolGauge indicator has boolGauge metric', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'loading',
        error: {} as ConnectError
      })
      
      const objective = new Objective({
        target: 0.98,
        labels: { __name__: 'test-bool-gauge-slo' },
        indicator: {
          options: {
            case: 'boolGauge',
            value: {
              boolGauge: { metric: '' } // Missing boolGauge metric
            }
          }
        }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      expect(screen.getByText('Unable to calculate (see console)')).toBeInTheDocument()
      expect(screen.getByTitle('Missing boolean gauge metric for calculation')).toBeInTheDocument()
    })

    it('displays calculated dynamic threshold for boolGauge indicators', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 0.8, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })
      
      const objective = createBoolGaugeObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={7} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        // Dynamic threshold = 0.8 * (1/16) * (1 - 0.98) = 0.8 * 0.0625 * 0.02 = 0.00100
        expect(screen.getByText('0.00100')).toBeInTheDocument()
      })
    })

    it('generates correct query for boolGauge indicators', () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'loading',
        error: {} as ConnectError
      })
      
      const objective = createBoolGaugeObjective()
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      // Verify the query uses count_over_time() for boolean gauges
      expect(mockUsePrometheusQuery).toHaveBeenCalledWith(
        mockPromClient,
        expect.stringContaining('sum(count_over_time(service_up[30d])) / sum(count_over_time(service_up[1h4m])'),
        expect.any(Number),
        expect.any(Object)
      )
    })
  })

  describe('Error Handling', () => {
    it('handles query errors', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Prometheus query failed' } as ConnectError
      })
      
      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-slo' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              total: { metric: 'test_metric' },
              errors: { metric: 'test_metric{status="error"}' }
            }
          }
        }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        expect(screen.getByText('Unable to calculate (see console)')).toBeInTheDocument()
        expect(screen.getByTitle('Query failed: Prometheus query failed')).toBeInTheDocument()
      })
    })

    it('handles empty query results', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [] // No data returned
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })
      
      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-slo' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              total: { metric: 'test_metric' },
              errors: { metric: 'test_metric{status="error"}' }
            }
          }
        }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      await waitFor(() => {
        expect(screen.getByText('No data available')).toBeInTheDocument()
        expect(screen.getByTitle('No traffic data available for this time range')).toBeInTheDocument()
      })
    })
  })

  describe('Fallback Cases', () => {
    it('displays N/A for unknown burn rate types', () => {
      // Mock an unknown burn rate type
      mockGetBurnRateType.mockReturnValue('unknown' as any)
      
      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-slo' }
      })
      
      render(
        <BurnRateThresholdDisplay 
          objective={objective} 
          factor={14} 
          promClient={mockPromClient} 
        />
      )
      
      expect(screen.getByText('N/A')).toBeInTheDocument()
    })
  })
})

describe('Integration Tests - Query Generation', () => {
    it('generates syntactically valid PromQL for all indicator types', () => {
      const testCases = [
        {
          name: 'Ratio indicator',
          objective: new Objective({
            target: 0.95,
            labels: { __name__: 'test-ratio' },
            indicator: {
              options: {
                case: 'ratio',
                value: {
                  total: { metric: 'http_requests_total' },
                  errors: { metric: 'http_requests_total{status=~"5.."}' }
                }
              }
            }
          }),
          expectedPattern: /sum\(increase\(http_requests_total\[30d\]\)\) \/ sum\(increase\(http_requests_total\[1h4m\]\)\)/
        },
        {
          name: 'Latency indicator',
          objective: new Objective({
            target: 0.99,
            labels: { __name__: 'test-latency' },
            indicator: {
              options: {
                case: 'latency',
                value: {
                  total: { metric: 'http_request_duration_seconds_count' },
                  success: { metric: 'http_request_duration_seconds_bucket{le="0.1"}' }
                }
              }
            }
          }),
          expectedPattern: /sum\(increase\(http_request_duration_seconds_count\[30d\]\)\) \/ sum\(increase\(http_request_duration_seconds_count\[1h4m\]\)\)/
        },
        {
          name: 'LatencyNative indicator',
          objective: new Objective({
            target: 0.995,
            labels: { __name__: 'test-latency-native' },
            indicator: {
              options: {
                case: 'latencyNative',
                value: {
                  total: { metric: 'connect_server_requests_duration_seconds' },
                  latency: '200ms'
                }
              }
            }
          }),
          expectedPattern: /sum\(histogram_count\(increase\(connect_server_requests_duration_seconds\[30d\]\)\)\) \/ sum\(histogram_count\(increase\(connect_server_requests_duration_seconds\[1h4m\]\)\)\)/
        },
        {
          name: 'BoolGauge indicator',
          objective: new Objective({
            target: 0.98,
            labels: { __name__: 'test-bool-gauge' },
            indicator: {
              options: {
                case: 'boolGauge',
                value: {
                  boolGauge: { metric: 'up{job="prometheus"}' }
                }
              }
            }
          }),
          expectedPattern: /sum\(count_over_time\(up\{job="prometheus"\}\[30d\]\)\) \/ sum\(count_over_time\(up\{job="prometheus"\}\[1h4m\]\)\)/
        }
      ]

      testCases.forEach(({ name, objective, expectedPattern }) => {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: null,
          status: 'loading',
          error: {} as ConnectError
        })

        render(
          <BurnRateThresholdDisplay 
            objective={objective} 
            factor={14} 
            promClient={mockPromClient} 
          />
        )

        // Verify the query matches the expected pattern for this indicator type
        expect(mockUsePrometheusQuery).toHaveBeenCalledWith(
          mockPromClient,
          expect.stringMatching(expectedPattern),
          expect.any(Number),
          expect.any(Object)
        )

        // Clear mocks for next test case
        jest.clearAllMocks()
      })
    })

    it('handles all factor-to-window mappings correctly', () => {
      const factorWindowMappings = [
        { factor: 14, sloWindow: '30d', longWindow: '1h4m' },
        { factor: 7, sloWindow: '30d', longWindow: '6h26m' },
        { factor: 2, sloWindow: '30d', longWindow: '1d1h43m' },
        { factor: 1, sloWindow: '30d', longWindow: '4d6h51m' }
      ]

      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-factor-mapping' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              total: { metric: 'test_metric' },
              errors: { metric: 'test_metric{status="error"}' }
            }
          }
        }
      })

      factorWindowMappings.forEach(({ factor, sloWindow, longWindow }) => {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: null,
          status: 'loading',
          error: {} as ConnectError
        })

        render(
          <BurnRateThresholdDisplay 
            objective={objective} 
            factor={factor} 
            promClient={mockPromClient} 
          />
        )

        // Verify the query uses the correct windows for this factor
        expect(mockUsePrometheusQuery).toHaveBeenCalledWith(
          mockPromClient,
          expect.stringContaining(`[${sloWindow}]) / sum(increase(test_metric[${longWindow}])`),
          expect.any(Number),
          expect.any(Object)
        )

        // Clear mocks for next test case
        jest.clearAllMocks()
      })
    })

    it('validates threshold constant calculations for all factors', () => {
      const factorConstantMappings = [
        { factor: 14, expectedConstant: 1/48, description: 'Critical alert 1' },
        { factor: 7, expectedConstant: 1/16, description: 'Critical alert 2' },
        { factor: 2, expectedConstant: 1/14, description: 'Warning alert 1' },
        { factor: 1, expectedConstant: 1/7, description: 'Warning alert 2' }
      ]

      const objective = new Objective({
        target: 0.99,
        labels: { __name__: 'test-threshold-constants' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              total: { metric: 'test_metric' },
              errors: { metric: 'test_metric{status="error"}' }
            }
          }
        }
      })

      factorConstantMappings.forEach(({ factor, expectedConstant, description }) => {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: new QueryResponse({
            options: {
              case: 'vector',
              value: new Vector({
                samples: [new Sample({ value: 2.0, time: BigInt(0), metric: {} })] // 2x traffic ratio
              })
            }
          }),
          status: 'success',
          error: {} as ConnectError
        })

        const { rerender } = render(
          <BurnRateThresholdDisplay 
            objective={objective} 
            factor={factor} 
            promClient={mockPromClient} 
          />
        )

        // Calculate expected threshold: traffic_ratio * threshold_constant * (1 - target)
        const expectedThreshold = 2.0 * expectedConstant * (1 - 0.99)
        
        expect(screen.getByText(expectedThreshold.toFixed(5))).toBeInTheDocument()

        // Clear for next test
        jest.clearAllMocks()
      })
    })
  })

  describe('Edge Case Handling', () => {
    it('handles missing metric names gracefully', () => {
      const testCases = [
        {
          name: 'Ratio with missing total metric',
          objective: new Objective({
            target: 0.95,
            labels: { __name__: 'test-missing-total' },
            indicator: {
              options: {
                case: 'ratio',
                value: {
                  total: { metric: '' }, // Missing
                  errors: { metric: 'test_errors' }
                }
              }
            }
          })
        },
        {
          name: 'LatencyNative with missing total metric',
          objective: new Objective({
            target: 0.99,
            labels: { __name__: 'test-missing-native-total' },
            indicator: {
              options: {
                case: 'latencyNative',
                value: {
                  total: { metric: '' }, // Missing
                  latency: '100ms'
                }
              }
            }
          })
        },
        {
          name: 'BoolGauge with missing metric',
          objective: new Objective({
            target: 0.98,
            labels: { __name__: 'test-missing-bool' },
            indicator: {
              options: {
                case: 'boolGauge',
                value: {
                  boolGauge: { metric: '' } // Missing
                }
              }
            }
          })
        }
      ]

      testCases.forEach(({ name, objective }) => {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: null,
          status: 'loading',
          error: {} as ConnectError
        })

        render(
          <BurnRateThresholdDisplay 
            objective={objective} 
            factor={14} 
            promClient={mockPromClient} 
          />
        )

        expect(screen.getByText('Unable to calculate (see console)')).toBeInTheDocument()

        // Clear for next test
        jest.clearAllMocks()
      })
    })

    it('handles non-finite traffic ratios correctly', () => {
      const edgeCases = [
        { value: Infinity, expectedError: 'Invalid traffic ratio (not finite)' },
        { value: -Infinity, expectedError: 'Invalid traffic ratio (not finite)' },
        { value: NaN, expectedError: 'Invalid traffic ratio (not finite)' },
        { value: -1, expectedError: 'No traffic data available' },
        { value: 0, expectedError: 'No traffic data available' }
      ]

      const objective = new Objective({
        target: 0.95,
        labels: { __name__: 'test-edge-cases' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              total: { metric: 'test_metric' },
              errors: { metric: 'test_metric{status="error"}' }
            }
          }
        }
      })

      edgeCases.forEach(({ value, expectedError }) => {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: new QueryResponse({
            options: {
              case: 'vector',
              value: new Vector({
                samples: [new Sample({ value, time: BigInt(0), metric: {} })]
              })
            }
          }),
          status: 'success',
          error: {} as ConnectError
        })

        const { rerender } = render(
          <BurnRateThresholdDisplay 
            objective={objective} 
            factor={14} 
            promClient={mockPromClient} 
          />
        )

        expect(screen.getByText('Unable to calculate (see console)')).toBeInTheDocument()
        expect(screen.getByTitle(`Error: ${expectedError}`)).toBeInTheDocument()

        // Clear for next test
        jest.clearAllMocks()
      })
    })
  })
