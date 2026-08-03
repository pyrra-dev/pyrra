import React from 'react'
import Tiles from './Tiles'
import ObjectiveTile from './ObjectiveTile'
import AvailabilityTile from './AvailabilityTile'
import ErrorBudgetTile from './ErrorBudgetTile'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

interface ObjectiveTilesProps {
  objective: Objective
  loading: boolean
  success: boolean
  errors: number | undefined
  total: number | undefined
}

// The three headline tiles (Objective / Availability / Error Budget) shared by the
// SLO detail page and the Create-SLO preview. Each caller wraps it in its own layout.
const ObjectiveTiles = ({
  objective,
  loading,
  success,
  errors,
  total,
}: ObjectiveTilesProps): React.JSX.Element => (
  <Tiles>
    <ObjectiveTile objective={objective} />
    <AvailabilityTile
      objective={objective}
      loading={loading}
      success={success}
      errors={errors}
      total={total}
    />
    <ErrorBudgetTile
      objective={objective}
      loading={loading}
      success={success}
      errors={errors}
      total={total}
    />
  </Tiles>
)

export default ObjectiveTiles
