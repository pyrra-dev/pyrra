import React from 'react'
import {ObjectiveType} from '../../App'
import {formatDuration} from '../../duration'

interface ObjectiveTileProps {
  window: number
  target: number
  objectiveType: ObjectiveType
  latency?: number
}

const ObjectiveTile = ({target, latency, window, objectiveType}: ObjectiveTileProps) => {
  if (window === 0 || target === 0) {
    return (
      <div>
        <h6>Objective</h6>
        <h2></h2>
      </div>
    )
  }

  switch (objectiveType) {
    case ObjectiveType.Ratio:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * target).toFixed(3)}%</h2>
          <>in {formatDuration(window * 1000)}</>
        </div>
      )
    case ObjectiveType.Latency:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * target).toFixed(3)}%</h2>
          <>in {formatDuration(window * 1000)}</>
          <br />
          <p className="details">faster than {formatDuration(latency ?? 0)}</p>
        </div>
      )
    case ObjectiveType.BoolGauge:
      return (
        <div>
          <h6 className="headline">Objective</h6>
          <h2 className="metric">{(100 * target).toFixed(3)}%</h2>
          <>in {formatDuration(window * 1000)}</>
        </div>
      )
    default:
      return <div></div>
  }
}

export default ObjectiveTile
