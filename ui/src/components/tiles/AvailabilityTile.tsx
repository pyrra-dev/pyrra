import React from 'react'
import {Spinner} from 'react-bootstrap'
import {QueryStatus} from 'react-query/types/core/types'
import {ObjectiveType} from '../../App'
import {QueryResponse} from '../../proto/prometheus/v1/prometheus_pb'

interface AvailabilityProps {
  target: number
  objectiveType: ObjectiveType
  total: CounterProps
  errors: CounterProps
}

interface CounterProps {
  count: number
  status: QueryStatus
}

const AvailabilityTile = ({
  target,
  objectiveType,
  total,
  errors,
}: AvailabilityProps): JSX.Element => {
  const headline = <h6 className="headline">Availability</h6>

  if (total.status === 'idle' || errors.status === 'idle') {
    return <div>{headline}</div>
  }

  if (total.status === 'loading' || errors.status === 'loading') {
    return (
      <div>
        {headline}
        <Spinner
          animation={'border'}
          style={{
            width: 50,
            height: 50,
            padding: 0,
            borderRadius: 50,
            borderWidth: 2,
            opacity: 0.25,
          }}
        />
      </div>
    )
  }

  if (total.status === 'error' || errors.status === 'error') {
    return (
      <div>
        {headline}
        <h2 className="error">Error</h2>
      </div>
    )
  }

  const percentage = 1 - errors.count / total.count

  return (
    <div className={percentage > target ? 'good' : 'bad'}>
      {headline}
      <h2 className="metric">{(100 * percentage).toFixed(3)}%</h2>
      <table className="details">
        <tbody>
          <tr>
            <td>{objectiveType === ObjectiveType.Latency ? 'Slow:' : 'Errors:'}</td>
            <td>{Math.floor(errors.count).toLocaleString()}</td>
          </tr>
          <tr>
            <td>Total:</td>
            <td>{Math.floor(total.count).toLocaleString()}</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

export default AvailabilityTile

export const responseToCounterProps = (
  response: QueryResponse | null,
  status: QueryStatus,
): CounterProps => {
  if (status === 'success' && response?.options.case === 'vector') {
    return {
      count: response.options.value.samples[0].value,
      status,
    }
  }

  return {count: 0, status}
}
