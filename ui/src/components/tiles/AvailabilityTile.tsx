import React from 'react'
import {Spinner} from '@/components/ui/spinner'
import {cn} from '@/lib/utils'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
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
  const headline = <h6 className="font-sans text-xl font-semibold opacity-50">Availability</h6>
  if (loading) {
    return (
      <div className="rounded-lg bg-card p-9 text-card-foreground">
        {headline}
        <Spinner className="h-12 w-12 opacity-25" />
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
        <div className={cn('rounded-lg p-9 font-sans', percentage > objective.target ? 'bg-success text-success-foreground' : 'bg-destructive text-destructive-foreground')}>
          {headline}
          <h2 className="inline-block mr-2 font-sans text-[40px] font-normal mb-0">{(100 * percentage).toFixed(3)}%</h2>
          <table className="opacity-50 font-medium">
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
        <div className="rounded-lg bg-card p-9 text-card-foreground">
          {headline}
          <h2>No data</h2>
        </div>
      )
    }
  }

  return (
    <div className="rounded-lg bg-card p-9 text-card-foreground">
      <>
        {headline}
        <h2 className="text-destructive font-sans text-[40px] font-normal">Error</h2>
      </>
    </div>
  )
}

export default AvailabilityTile
