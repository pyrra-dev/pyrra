import React from 'react'
import {hasObjectiveType, ObjectiveType, renderLatencyTarget} from '../../App'
import {formatDuration} from '../../duration'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ObjectiveTileProps {
  objective: Objective
}

const ObjectiveTile = ({objective}: ObjectiveTileProps): React.JSX.Element => {
  const objectiveType = hasObjectiveType(objective)
  switch (objectiveType) {
    case ObjectiveType.Ratio:
      return (
        <div className="rounded-lg bg-card p-9 text-card-foreground">
          <h6 className="font-sans text-xl font-semibold opacity-50">Objective</h6>
          <h2 className="inline-block mr-2 font-sans text-[40px] font-normal mb-0">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
        </div>
      )
    case ObjectiveType.BoolGauge:
      return (
        <div className="rounded-lg bg-card p-9 text-card-foreground">
          <h6 className="font-sans text-xl font-semibold opacity-50">Objective</h6>
          <h2 className="inline-block mr-2 font-sans text-[40px] font-normal mb-0">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
        </div>
      )
    case ObjectiveType.Latency:
    case ObjectiveType.LatencyNative:
      return (
        <div className="rounded-lg bg-card p-9 text-card-foreground">
          <h6 className="font-sans text-xl font-semibold opacity-50">Objective</h6>
          <h2 className="inline-block mr-2 font-sans text-[40px] font-normal mb-0">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
          <br />
          <p className="opacity-50 font-medium">faster than {renderLatencyTarget(objective)}</p>
        </div>
      )
    default:
      return <div></div>
  }
}

export default ObjectiveTile
