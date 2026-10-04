import {fireEvent, render, screen, waitFor, cleanup} from '@testing-library/react'
import {QueryClient, QueryClientProvider} from '@tanstack/react-query'
import {MemoryRouter} from 'react-router-dom'
import {afterEach, describe, expect, it, vi} from 'vitest'
import {previewObjective} from '../components/create/preview'
import Create from './Create'

vi.mock('../App', () => ({API_BASEPATH: 'http://localhost:9099'}))
vi.mock('../components/create/preview', async (original) => ({
  ...await original<typeof import('../components/create/preview')>(),
  previewObjective: vi.fn().mockResolvedValue({labels: {}, target: 0.99}),
}))
vi.mock('../components/create/DetailPreview', () => ({
  default: ({target, config, stale}: {target: number; config: string; stale: boolean}) => (
    <div data-testid="preview" data-stale={String(stale)}>
      <span data-testid="target">{target}</span><pre>{config}</pre>
    </div>
  ),
}))

afterEach(() => { cleanup(); vi.clearAllMocks(); })

describe('live target preview', () => {
  it('keeps a finite query baseline and original duration config while target edits stay local', async () => {
    render(
      <QueryClientProvider client={new QueryClient({defaultOptions: {queries: {retry: false}}})}>
        <MemoryRouter><Create /></MemoryRouter>
      </QueryClientProvider>,
    )
    const input = screen.getByLabelText('Target')
    fireEvent.change(input, {target: {value: '100'}})
    fireEvent.click(screen.getByRole('button', {name: /^preview$/i}))
    await waitFor(() => { expect(screen.getByTestId('target')).toHaveTextContent('1'); })
    expect(vi.mocked(previewObjective).mock.calls[0][1]).toContain("target: '99'")
    expect(screen.getByTestId('preview')).toHaveTextContent("target: '100'")
    fireEvent.keyDown(input, {key: 'ArrowDown'})
    expect(Number(screen.getByTestId('target').textContent)).toBeCloseTo(0.9999999, 10)
    fireEvent.change(input, {target: {value: '99'}})
    expect(screen.getByTestId('target')).toHaveTextContent('0.99')
    fireEvent.change(input, {target: {value: ''}})
    expect(screen.getByTestId('target')).toHaveTextContent('0.99')
    expect(screen.getByTestId('preview')).toHaveAttribute('data-stale', 'false')
    expect(previewObjective).toHaveBeenCalledTimes(1)
  })
})
