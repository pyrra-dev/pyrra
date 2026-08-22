import React from 'react'
import {Badge} from '@/components/ui/badge'
import {cn} from '@/lib/utils'
import {type Labels, MetricName} from '../labels'

interface ObjectiveLabelsProps {
  labels: Labels
  grouping?: Labels
  className?: string
}

// Renders an objective's labels (and optional grouping labels) as secondary badges,
// dropping the __name__ metric label. Shared by the detail page and the Create preview.
const ObjectiveLabels = ({labels, grouping, className}: ObjectiveLabelsProps): React.JSX.Element => (
  <>
    {Object.entries({...labels, ...grouping})
      .filter(([k]) => k !== MetricName)
      .map(([k, v]) => (
        <Badge key={k} variant="secondary" className={cn('mr-1 font-normal', className)}>
          {k}={v}
        </Badge>
      ))}
  </>
)

export default ObjectiveLabels
