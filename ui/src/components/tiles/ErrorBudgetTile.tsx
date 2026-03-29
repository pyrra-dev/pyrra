import React from 'react'
import {Spinner} from '@/components/ui/spinner'
import {cn} from '@/lib/utils'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ErrorBudgetTileProps {
  objective: Objective
  loading: boolean
  success: boolean
  errors: number | undefined
  total: number | undefined
}

const ErrorBudgetTile = ({objective, loading, success, errors, total}: ErrorBudgetTileProps) => {
  const headline = <h6 className="font-sans text-xl font-semibold opacity-50">Error Budget</h6>

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
      const budget = 1 - objective.target
      const unavailability = errors / total
      const availableBudget = (budget - unavailability) / budget

      return (
        <div className={cn('rounded-lg p-9 font-sans', availableBudget > 0 ? 'bg-success text-success-foreground' : 'bg-destructive text-destructive-foreground')}>
          {headline}
          <h2 className="inline-block mr-2 font-sans text-[40px] font-normal mb-0">{(100 * availableBudget).toFixed(3)}%</h2>
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
      {headline}
      <h2 className="text-destructive font-sans text-[40px] font-normal">Error</h2>
    </div>
  )
}

export default ErrorBudgetTile
