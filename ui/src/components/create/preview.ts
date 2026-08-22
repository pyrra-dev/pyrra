// Backend seam for the live Create-SLO preview.
//
// previewObjective() materializes a draft SLO (as YAML) into a live Objective by
// running it through Pyrra's own Go parsing on the backend, so the preview renders
// exactly as a stored SLO would — same queries, same tiles, same graphs against
// real Prometheus data.
//
// The backend `Preview` endpoint is not wired up yet. Until it lands this throws
// PreviewUnavailableError and the editor falls back to the (fully live) YAML view.
// When the endpoint exists, implement the call here and the DetailPreview lights up
// with no further changes to the page.

import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

export type PreviewStatus = 'idle' | 'loading' | 'success' | 'unavailable' | 'error'

export class PreviewUnavailableError extends Error {
  constructor(message = 'preview backend not available yet') {
    super(message)
    this.name = 'PreviewUnavailableError'
  }
}

export async function previewObjective(_baseUrl: string, _config: string): Promise<Objective> {
  // TODO(backend): POST the YAML to the Preview endpoint, which parses it via
  // ServiceLevelObjective.Internal() + FromInternal() and fills in the queries
  // (the same path as ObjectiveService.List), then return the materialized Objective.
  throw new PreviewUnavailableError()
}
