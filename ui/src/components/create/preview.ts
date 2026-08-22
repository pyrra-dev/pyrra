// Backend seam for the live Create-SLO preview.
//
// previewObjective() materializes a draft SLO (as YAML) into a live Objective by
// calling the ObjectiveService.Preview RPC, which runs it through Pyrra's own Go
// parsing on the backend, so the preview renders exactly as a stored SLO would —
// same queries, same tiles, same graphs against real Prometheus data.

import {Code, ConnectError, createClient} from '@connectrpc/connect'
import {createConnectTransport} from '@connectrpc/connect-web'
import {ObjectiveService, type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'

export type PreviewStatus = 'idle' | 'loading' | 'success' | 'unavailable' | 'error'

export class PreviewUnavailableError extends Error {
  constructor(message = 'preview backend not available yet') {
    super(message)
    this.name = 'PreviewUnavailableError'
  }
}

// grouping optionally scopes the preview to a single grouping label set (e.g.
// {handler="/api"}); empty previews the objective grouped by its grouping labels.
export async function previewObjective(baseUrl: string, config: string, grouping = ''): Promise<Objective> {
  const client = createClient(ObjectiveService, createConnectTransport({baseUrl}))
  try {
    const response = await client.preview({config, grouping})
    if (response.objective === undefined) {
      throw new Error('preview response did not contain an objective')
    }
    return response.objective
  } catch (err) {
    // An API server that predates the Preview RPC answers Unimplemented. Surface
    // that as "unavailable" so the editor falls back to the (fully live) YAML view
    // instead of showing it as a hard error.
    if (err instanceof ConnectError && err.code === Code.Unimplemented) {
      throw new PreviewUnavailableError()
    }
    throw err
  }
}
