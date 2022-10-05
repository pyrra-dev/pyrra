import React from 'react'
import {Spinner} from 'react-bootstrap'
import {QueryStatus} from 'react-query/types/core/types'

interface ErrorBudgetTileProps {
  target: number
  total: CounterProps
  errors: CounterProps
}

interface CounterProps {
  count: number
  status: QueryStatus
}

const ErrorBudgetTile = ({target, total, errors}: ErrorBudgetTileProps) => {
  const headline = <h6 className="headline">Error Budget</h6>

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

  const budget = 1 - target
  const unavailability = errors.count / total.count
  const remainingBudget = (budget - unavailability) / budget

  return (
    <div className={remainingBudget > 0 ? 'good' : 'bad'}>
      {headline}
      <h2>{(100 * remainingBudget).toFixed(3)}%</h2>
    </div>
  )
}

export default ErrorBudgetTile
