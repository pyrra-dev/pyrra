import React from 'react'
import {Spinner} from 'react-bootstrap'
import {Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {hasObjectiveType, ObjectiveType} from '../../App'

interface AvailabilityTileProps {
  objective: Objective
  loading: boolean
  success: boolean
  errors: number | undefined
  total: number | undefined
}

const AvailabilityTile = ({
  objective,
  loading,
  success,
  errors,
  total,
}: AvailabilityTileProps): React.JSX.Element => {
  console.log(loading, success, errors, total)

  const headline = <h6 className="headline">Availability</h6>
  if (loading) {
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

  if (success) {
    if (errors !== undefined && total !== undefined) {
      const percentage = 1 - errors / total

      const objectiveType = hasObjectiveType(objective)
      const objectiveTypeLatency =
        objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative

      return (
        <div className={percentage > objective.target ? 'good' : 'bad'}>
          {headline}
          <h2 className="metric">{(100 * percentage).toFixed(3)}%</h2>
          <table className="details">
            <tbody>
              <tr>
                <td>{objectiveTypeLatency ? 'Slow:' : 'Errors:'}</td>
                <td>{Math.floor(errors).toLocaleString()}</td>
              </tr>
              <tr>
                <td>Total:</td>
                <td>{Math.floor(total).toLocaleString()}</td>
              </tr>
            </tbody>
          </table>
        </div>
      )
    } else {
      return (
        <div>
          {headline}
          <h2>No data</h2>
        </div>
      )
    }
  }

  return (
    <div>
      <>
        {headline}
        <h2 className="error">Error</h2>
      </>
    </div>
  )
}

export default AvailabilityTile
