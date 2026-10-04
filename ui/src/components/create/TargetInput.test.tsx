import {useState} from 'react'
import {fireEvent, render, screen, cleanup} from '@testing-library/react'
import {afterEach, describe, expect, it} from 'vitest'
import TargetInput from './TargetInput'

afterEach(cleanup)

const Editor = ({initial = '99'}): React.JSX.Element => {
  const [value, setValue] = useState(initial)
  return <TargetInput value={value} onChange={setValue} />
}

describe('target input', () => {
  it('walks common targets in both directions, including five nines', () => {
    render(<Editor />)
    const input = screen.getByRole('spinbutton')
    for (const target of [99.5, 99.9, 99.95, 99.99, 99.995, 99.999, 99.9995, 99.9999, 99.99995, 99.99999, 100, 100]) {
      fireEvent.keyDown(input, {key: 'ArrowUp'})
      expect(input).toHaveValue(target)
    }
    fireEvent.keyDown(input, {key: 'ArrowDown'})
    expect(input).toHaveValue(99.99999)
  })

  it('steps off arbitrary targets and clamps at zero', () => {
    render(<Editor initial="98.12345" />)
    const input = screen.getByRole('spinbutton')
    fireEvent.keyDown(input, {key: 'ArrowUp'})
    expect(input).toHaveValue(99)
    fireEvent.keyDown(input, {key: 'ArrowDown'})
    expect(input).toHaveValue(95)
    fireEvent.change(input, {target: {value: '0'}})
    fireEvent.keyDown(input, {key: 'ArrowDown'})
    expect(input).toHaveValue(0)
  })

  it('accepts five decimals, limits extra decimals, and allows clearing', () => {
    render(<Editor />)
    const input = screen.getByRole('spinbutton')
    fireEvent.change(input, {target: {value: '99.99999'}})
    expect(input).toHaveValue(99.99999)
    expect(input).toBeValid()
    fireEvent.change(input, {target: {value: '98.123456'}})
    expect(input).toHaveValue(98.12346)
    fireEvent.change(input, {target: {value: '1e-6'}})
    expect(input).toHaveValue(0)
    fireEvent.change(input, {target: {value: ''}})
    expect(input).toHaveValue(null)
    fireEvent.keyDown(input, {key: 'ArrowUp'})
    expect(input).toHaveValue(99.5)
  })
})
