import React from 'react'
import {hasObjectiveType, ObjectiveType, renderLatencyTarget} from '../../App'
import {formatDuration} from '../../duration'
import {Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ObjectiveTileProps {
  objective: Objective
}

const ObjectiveTile = ({objective}: ObjectiveTileProps): React.JSX.Element => {
  // console.log("objective title")
  const objectiveType = hasObjectiveType(objective)
  if (objectiveType === ObjectiveType.Ratio) {
    return (
      <div>
        <h6 className="headline">Objective</h6>
        <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
        <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
      </div>
    )
  } else if (objectiveType === ObjectiveType.BoolGauge) {
    return (
      <div>
        <h6 className="headline">Objective</h6>
        <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
        <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
      </div>
    )
  } else if (objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative) {
    // latencyTarget always returns value in milliseconds, formatDuration handles the display
    const latencyText = renderLatencyTarget(objective)

    return (
      <div>
        <h6 className="headline">Objective</h6>
        <h2 className="metric">{(100 * objective.target).toFixed(3)}%</h2>
        <>in {formatDuration(Number(objective.window?.seconds) * 1000)}</>
        <br />
        <p className="details">faster than {latencyText}</p>
      </div>
    )
  } else {
    return <div></div>
  }
}

export default ObjectiveTile
