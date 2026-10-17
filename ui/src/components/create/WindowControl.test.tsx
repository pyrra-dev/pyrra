import {useState} from 'react'
import {fireEvent, render, screen, cleanup} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {afterEach, describe, expect, it} from 'vitest'
import {WindowControl} from './editorFields'

afterEach(cleanup)

const Editor = ({initial = '7d'}): React.JSX.Element => {
  const [value, setValue] = useState(initial)
  return (
    <>
      <WindowControl value={value} onChange={setValue} />
      <output data-testid="window">{value}</output>
    </>
  )
}

describe('window control', () => {
  it('keeps typing a custom window whose prefix is a preset', async () => {
    render(<Editor />)
    const input = screen.getByLabelText('Custom window')
    await userEvent.clear(input)
    await userEvent.type(input, '1w3d')
    expect(input).toHaveValue('1w3d')
    expect(screen.getByTestId('window')).toHaveTextContent('1w3d')
  })

  it('clears the custom field when a preset is picked', async () => {
    render(<Editor />)
    const input = screen.getByLabelText('Custom window')
    await userEvent.clear(input)
    await userEvent.type(input, '2w')
    fireEvent.click(screen.getByLabelText('4w'))
    fireEvent.click(screen.getByLabelText('2w'))
    expect(input).toHaveValue('')
    expect(screen.getByTestId('window')).toHaveTextContent('2w')
  })
})
