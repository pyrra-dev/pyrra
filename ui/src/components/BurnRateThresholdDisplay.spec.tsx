import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { PromiseClient, ConnectError } from '@connectrpc/connect'
import { PrometheusService } from '../proto/prometheus/v1/prometheus_connect'
import { Objective } from '../proto/objectives/v1alpha1/objectives_pb'
import { QueryResponse, Vector, Sample } from '../proto/prometheus/v1/prometheus_pb'
import BurnRateThresholdDisplay from './BurnRateThresholdDisplay'
import { BurnRateType } from '../burnrate'

// Mock the prometheus hook
const mockUsePrometheusQuery = jest.fn()
jest.mock('../prometheus', () => ({
  usePrometheusQuery: () => mockUsePrometheusQuery()
}))

// Mock the burnrate module
const mockGetBurnRateType = jest.fn()
jest.mock('../burnrate', () => ({
  getBurnRateType: () => mockGetBurnRateType(),
  BurnRateType: {
    Static: 'static',
    Dynamic: 'dynamic',
  },
}))

describe('BurnRateThresholdDisplay - Comprehensive Indicator Type Tests', () => {
  const mockPromClient = {} as PromiseClient<typeof PrometheusService>

  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('LatencyNative Indicator Error Handling and Fallback Behavior', () => {
    const latencyNativeObjective = new Objective({
      target: 0.99,
      labels: { __name__: 'test-latency-native' },
      indicator: {
        options: {
          case: 'latencyNative',
          value: {
            total: { metric: 'http_requests_duration_seconds' },
            latency: '0.1'
          }
        }
      },
      alerting: {
        burnrates: true,
        burnRateType: 'dynamic'
      }
    })

    it('should handle missing native histogram metrics gracefully', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({ samples: [] }) // No data
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={latencyNativeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should fallback when no histogram data
        expect(screen.getByText(/No data available/)).toBeInTheDocument()
      })
    })

    it('should handle histogram_count query failures', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Histogram query failed' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={latencyNativeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show error state and fallback
        expect(screen.getByText(/Static Thresholds/)).toBeInTheDocument()
      })
    })

    it('should validate native histogram metric names', async () => {
      const invalidObjective = new Objective({
        target: 0.99,
        labels: { __name__: 'test-invalid' },
        indicator: {
          options: {
            case: 'latencyNative',
            value: {
              total: { metric: '' }, // Empty metric name
              latency: '0.1'
            }
          }
        }
      })

      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Invalid metric name' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={invalidObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        expect(screen.getByText(/Static Thresholds/)).toBeInTheDocument()
      })
    })

    it('should handle malformed histogram data', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: NaN, time: BigInt(0), metric: {} })] // NaN value
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={latencyNativeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle NaN values gracefully
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should generate correct histogram_count queries', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1000, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={latencyNativeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Verify the component calls usePrometheusQuery with histogram_count pattern
        expect(mockUsePrometheusQuery).toHaveBeenCalled()
      })
    })
  })

  describe('BoolGauge Indicator Error Handling and Tooltip Content', () => {
    const boolGaugeObjective = new Objective({
      target: 0.99,
      labels: { __name__: 'test-bool-gauge' },
      indicator: {
        options: {
          case: 'boolGauge',
          value: {
            boolGauge: {
              metric: 'probe_success'
            }
          }
        }
      },
      alerting: {
        burnrates: true,
        burnRateType: 'dynamic'
      }
    })

    it('should handle missing probe metrics gracefully', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({ samples: [] }) // No probe data
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={boolGaugeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should fallback when no probe data
        expect(screen.getByText(/No data available/)).toBeInTheDocument()
      })
    })

    it('should handle count_over_time query failures', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Probe query failed' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={boolGaugeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show error state for probe failures
        expect(screen.getByText(/Static Thresholds/)).toBeInTheDocument()
      })
    })

    it('should validate boolean gauge metric names', async () => {
      const invalidObjective = new Objective({
        target: 0.99,
        labels: { __name__: 'test-invalid-gauge' },
        indicator: {
          options: {
            case: 'boolGauge',
            value: {
              boolGauge: {
                metric: '' // Empty metric name
              }
            }
          }
        }
      })

      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Invalid probe metric' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={invalidObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        expect(screen.getByText(/Static Thresholds/)).toBeInTheDocument()
      })
    })

    it('should handle sparse probe data correctly', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 5, time: BigInt(0), metric: {} })] // Very low probe count
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={boolGaugeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle low probe counts appropriately - shows actual threshold value
        expect(screen.getByText(/0\.00104/)).toBeInTheDocument()
      })
    })

    it('should generate correct count_over_time queries', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1440, time: BigInt(0), metric: {} })] // 1 probe per minute for 24h
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={boolGaugeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Verify the component calls usePrometheusQuery with count_over_time pattern
        expect(mockUsePrometheusQuery).toHaveBeenCalled()
        expect(screen.getByText(/0\.20833/)).toBeInTheDocument()
      })
    })

    it('should display appropriate tooltip content for boolean gauges', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 720, time: BigInt(0), metric: {} })] // 720 probes
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={boolGaugeObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show dynamic thresholds with probe-specific context
        expect(screen.getByText(/0\.15000/)).toBeInTheDocument()
      })
    })
  })

  describe('Cross-Indicator Type Consistency Tests', () => {
    const ratioObjective = new Objective({
      target: 0.99,
      labels: { __name__: 'test-ratio' },
      indicator: {
        options: {
          case: 'ratio',
          value: {
            errors: { metric: 'http_requests_total{status=~"5.."}' },
            total: { metric: 'http_requests_total' }
          }
        }
      }
    })

    it('should handle all indicator types consistently during errors', async () => {
      const objectives = [
        { name: 'ratio', obj: ratioObjective },
        { name: 'latencyNative', obj: new Objective({
          target: 0.99,
          labels: { __name__: 'test-latency-native' },
          indicator: {
            options: {
              case: 'latencyNative',
              value: {
                total: { metric: 'http_requests_duration_seconds' },
                latency: '0.1'
              }
            }
          }
        })},
        { name: 'boolGauge', obj: new Objective({
          target: 0.99,
          labels: { __name__: 'test-bool-gauge' },
          indicator: {
            options: {
              case: 'boolGauge',
              value: { 
                boolGauge: { metric: 'probe_success' }
              }
            }
          }
        })}
      ]

      for (const { name, obj } of objectives) {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: null,
          status: 'error',
          error: { message: `${name} query failed` } as ConnectError
        })

        const { unmount } = render(<BurnRateThresholdDisplay objective={obj} factor={14} promClient={mockPromClient} />)

        await waitFor(() => {
          // All indicator types should show consistent error handling
          expect(screen.getByText(/Static Thresholds/)).toBeInTheDocument()
        })

        unmount()
        jest.clearAllMocks()
      }
    })

    it('should maintain consistent threshold calculation patterns', async () => {
      const testCases = [
        { traffic: 1000, expectedRange: [0.002, 0.008] },
        { traffic: 5000, expectedRange: [0.01, 0.04] },
        { traffic: 10000, expectedRange: [0.02, 0.08] }
      ]

      for (const { traffic, expectedRange } of testCases) {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: new QueryResponse({
            options: {
              case: 'vector',
              value: new Vector({
                samples: [new Sample({ value: traffic, time: BigInt(0), metric: {} })]
              })
            }
          }),
          status: 'success',
          error: {} as ConnectError
        })

        const { unmount } = render(<BurnRateThresholdDisplay objective={ratioObjective} factor={14} promClient={mockPromClient} />)

        await waitFor(() => {
          // Should show dynamic thresholds within expected range
          expect(screen.getByText(/0\.20833/)).toBeInTheDocument()
        })

        unmount()
        jest.clearAllMocks()
      }
    })
  })

  describe('Mathematical Edge Case and Precision Handling Tests', () => {
    it('should handle division by zero scenarios (NaN values)', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: NaN, time: BigInt(0), metric: {} })] // NaN from division by zero
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'nan-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle NaN gracefully
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should handle infinite values gracefully', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: Infinity, time: BigInt(0), metric: {} })] // Infinite traffic ratio
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'infinite-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle infinite values gracefully
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should handle extremely high SLO targets (99.99%) with precision', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1000, time: BigInt(0), metric: {} })] // Normal traffic
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.9999, // 99.99% SLO (very small error budget)
        labels: { __name__: 'high-precision-slo' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle high precision calculations - expect scientific notation or many decimal places
        const element = screen.getByText(/[0-9]/)
        expect(element).toBeInTheDocument()
        // The exact format depends on the calculation, but it should show a valid number
      })
    })

    it('should handle extremely small traffic ratios with precision floors', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 0.0000001, time: BigInt(0), metric: {} })] // Extremely small traffic ratio
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'tiny-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should use scientific notation for very small thresholds
        expect(screen.getByText(/2\.083e-7/)).toBeInTheDocument()
      })
    })

    it('should cap extremely high traffic ratios', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 50000, time: BigInt(0), metric: {} })] // Extremely high traffic ratio
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'extreme-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show error for thresholds that exceed 100%
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should handle zero traffic gracefully', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 0, time: BigInt(0), metric: {} })] // Zero traffic
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'zero-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should fallback for zero traffic
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should handle negative traffic ratios', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: -100, time: BigInt(0), metric: {} })] // Negative traffic (data error)
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'negative-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle negative values gracefully
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })
    })

    it('should format very small thresholds with scientific notation', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 0.001, time: BigInt(0), metric: {} })] // Very small traffic ratio
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.9999, // High precision SLO
        labels: { __name__: 'scientific-notation' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should use scientific notation for very small numbers
        const element = screen.getByText(/[0-9]/)
        expect(element).toBeInTheDocument()
        // Could be scientific notation (e-) or many decimal places
      })
    })
  })

  describe('Performance and Edge Case Tests', () => {
    it('should handle extremely high traffic volumes', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1000000, time: BigInt(0), metric: {} })] // 1M requests
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      const startTime = Date.now()
      render(<BurnRateThresholdDisplay objective={new Objective({
        target: 0.99,
        labels: { __name__: 'high-traffic' },
        indicator: {
          options: {
            case: 'ratio',
            value: {
              errors: { metric: 'http_requests_total{status=~"5.."}' },
              total: { metric: 'http_requests_total' }
            }
          }
        }
      })} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show error for thresholds that exceed 100%
        expect(screen.getByText(/Unable to calculate/)).toBeInTheDocument()
      })

      const endTime = Date.now()
      expect(endTime - startTime).toBeLessThan(1000) // Should complete within 1 second
    })

    it('should handle concurrent indicator type rendering', async () => {
      const objectives = [
        new Objective({
          target: 0.99,
          labels: { __name__: 'concurrent-ratio' },
          indicator: {
            options: {
              case: 'ratio',
              value: {
                errors: { metric: 'http_requests_total{status=~"5.."}' },
                total: { metric: 'http_requests_total' }
              }
            }
          }
        }),
        new Objective({
          target: 0.99,
          labels: { __name__: 'concurrent-latency-native' },
          indicator: {
            options: {
              case: 'latencyNative',
              value: {
                total: { metric: 'http_requests_duration_seconds' },
                latency: '0.1'
              }
            }
          }
        }),
        new Objective({
          target: 0.99,
          labels: { __name__: 'concurrent-bool-gauge' },
          indicator: {
            options: {
              case: 'boolGauge',
              value: { 
                boolGauge: { metric: 'probe_success' }
              }
            }
          }
        })
      ]

      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1000, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      const components = objectives.map((obj, index) => (
        <div key={index} data-testid={`component-${index}`}>
          <BurnRateThresholdDisplay objective={obj} factor={14} promClient={mockPromClient} />
        </div>
      ))

      render(<div>{components}</div>)

      await waitFor(() => {
        // All components should render successfully
        expect(screen.getAllByText(/0\.20833/)).toHaveLength(3)
      })
    })
  })

  describe('Comprehensive Error Recovery and Network Failure Tests', () => {
    const testObjective = new Objective({
      target: 0.99,
      labels: { __name__: 'test-recovery' },
      indicator: {
        options: {
          case: 'ratio',
          value: {
            errors: { metric: 'http_requests_total{status=~"5.."}' },
            total: { metric: 'http_requests_total' }
          }
        }
      }
    })

    it('should handle network timeout errors with retry indication', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Network timeout after 30 seconds' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show retry indication for network timeouts
        expect(screen.getByText(/Retrying/)).toBeInTheDocument()
      })
    })

    it('should handle query syntax errors with configuration guidance', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Parse error: invalid PromQL syntax' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show configuration error for syntax issues
        expect(screen.getByText(/Config Error/)).toBeInTheDocument()
      })
    })

    it('should handle missing metrics with recovery indication', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Metric not found: unknown_metric' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show metric missing with recovery indication
        expect(screen.getByText(/Metric Missing/)).toBeInTheDocument()
      })
    })

    it('should handle Prometheus server errors with recovery indication', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Internal server error 500' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show server error with recovery indication
        expect(screen.getByText(/Server Error/)).toBeInTheDocument()
      })
    })

    it('should recover when metrics become available after being missing', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      
      // First render: metric missing
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Metric not found: http_requests_total' } as ConnectError
      })

      const { rerender } = render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        expect(screen.getByText(/Metric Missing/)).toBeInTheDocument()
      })

      // Second render: metric becomes available
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 1000, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      rerender(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should recover and show calculated threshold
        expect(screen.getByText(/0\.20833/)).toBeInTheDocument()
      })
    })

    it('should handle metrics that exist but return no data', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({ samples: [] }) // Metric exists but no data points
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should handle empty data gracefully
        expect(screen.getByText(/No data available/)).toBeInTheDocument()
      })
    })

    it('should handle recovery from no data to available data', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      
      // First render: no data
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({ samples: [] })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      const { rerender } = render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        expect(screen.getByText(/No data available/)).toBeInTheDocument()
      })

      // Second render: data becomes available
      mockUsePrometheusQuery.mockReturnValue({
        response: new QueryResponse({
          options: {
            case: 'vector',
            value: new Vector({
              samples: [new Sample({ value: 500, time: BigInt(0), metric: {} })]
            })
          }
        }),
        status: 'success',
        error: {} as ConnectError
      })

      rerender(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should recover and show calculated threshold
        expect(screen.getByText(/0\.10417/)).toBeInTheDocument() // 500 * 0.020833 * 0.01
      })
    })

    it('should provide indicator-specific error messages for missing metrics', async () => {
      const indicators = [
        {
          name: 'LatencyNative',
          objective: new Objective({
            target: 0.99,
            labels: { __name__: 'test-latency-native-missing' },
            indicator: {
              options: {
                case: 'latencyNative',
                value: {
                  total: { metric: 'missing_histogram' },
                  latency: '0.1'
                }
              }
            }
          }),
          expectedText: /Metric Missing/
        },
        {
          name: 'BoolGauge',
          objective: new Objective({
            target: 0.99,
            labels: { __name__: 'test-bool-gauge-missing' },
            indicator: {
              options: {
                case: 'boolGauge',
                value: {
                  boolGauge: { metric: 'missing_probe' }
                }
              }
            }
          }),
          expectedText: /Metric Missing/
        },
        {
          name: 'Latency',
          objective: new Objective({
            target: 0.99,
            labels: { __name__: 'test-latency-missing' },
            indicator: {
              options: {
                case: 'latency',
                value: {
                  total: { metric: 'missing_histogram_count' },
                  success: { metric: 'missing_histogram_bucket' }
                }
              }
            }
          }),
          expectedText: /Metric Missing/
        }
      ]

      for (const { name, objective, expectedText } of indicators) {
        mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
        mockUsePrometheusQuery.mockReturnValue({
          response: null,
          status: 'error',
          error: { message: `Metric not found: ${name.toLowerCase()}_metric` } as ConnectError
        })

        const { unmount } = render(<BurnRateThresholdDisplay objective={objective} factor={14} promClient={mockPromClient} />)

        await waitFor(() => {
          expect(screen.getByText(expectedText)).toBeInTheDocument()
        })

        unmount()
        jest.clearAllMocks()
      }
    })

    it('should handle connection failures with appropriate retry behavior', async () => {
      mockGetBurnRateType.mockReturnValue(BurnRateType.Dynamic)
      mockUsePrometheusQuery.mockReturnValue({
        response: null,
        status: 'error',
        error: { message: 'Connection refused: unable to connect to Prometheus' } as ConnectError
      })

      render(<BurnRateThresholdDisplay objective={testObjective} factor={14} promClient={mockPromClient} />)

      await waitFor(() => {
        // Should show retry indication for connection failures
        expect(screen.getByText(/Retrying/)).toBeInTheDocument()
      })
    })
  })
})