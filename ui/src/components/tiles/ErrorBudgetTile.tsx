import React from 'react'
import {Spinner} from 'react-bootstrap'
import {Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ErrorBudgetTileProps {
  objective: Objective
  loading: boolean
  success: boolean
  errors: number | undefined
  total: number | undefined
}

const ErrorBudgetTile = ({objective, loading, success, errors, total}: ErrorBudgetTileProps) => {
  const headline = <h6 className="headline">Error Budget</h6>

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
      const budget = 1 - objective.target
      const unavailability = errors / total
      const availableBudget = (budget - unavailability) / budget

      return (
        <div className={availableBudget > 0 ? 'good' : 'bad'}>
          {headline}
          <h2 className="metric">{(100 * availableBudget).toFixed(3)}%</h2>
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
      {headline}
      <h2 className="error">Error</h2>
    </div>
  )
}

export default ErrorBudgetTile
