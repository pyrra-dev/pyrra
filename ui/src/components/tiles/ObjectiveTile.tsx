import React from 'react'
import {hasObjectiveType, ObjectiveType, renderLatencyTarget} from '../../App'
import {formatDuration} from '../../duration'
import {Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ObjectiveTileProps {
  objective: Objective
}

const ObjectiveTile = ({objective}: ObjectiveTileProps): React.JSX.Element => {
  const objectiveType = hasObjectiveType(objective)
  switch (objectiveType) {
    case ObjectiveType.Ratio:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
        </div>
      )
    case ObjectiveType.BoolGauge:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
        </div>
      )
    case ObjectiveType.Latency:
    case ObjectiveType.LatencyNative:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
          <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
          <br />
          <p className="details">faster than {renderLatencyTarget(objective)}</p>
        </div>
      )
    default:
      return <div></div>
  }
}

export default ObjectiveTile
