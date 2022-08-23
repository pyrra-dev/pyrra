import React, {EventHandler, useRef} from 'react'

interface ToggleProps {
  checked?: boolean
  onChange?: EventHandler<any>
  onText?: string
  offText?: string
}

const Toggle = ({checked, onChange, onText, offText}: ToggleProps): JSX.Element => {
  const element = useRef<HTMLDivElement>(null)
  const elementOn = useRef<HTMLLabelElement>(null)
  const elementOff = useRef<HTMLLabelElement>(null)

  let width = 0
  let height = 0
  if (element.current !== null && elementOn.current !== null && elementOff.current !== null) {
    width = elementOn.current.clientWidth
    height = Math.max(elementOn.current.clientHeight, elementOff.current.clientHeight)
  }

  const defaultClasses = ['toggle', 'btn']
  const classes =
    checked !== undefined && checked
      ? defaultClasses.concat('btn-dark')
      : defaultClasses.concat('btn-light off')

  return (
    <div
      className={classes.join(' ')}
      role="button"
      onClick={onChange}
      style={{width, height}}
      ref={element}>
      <input type="checkbox" checked={checked} />
      <div className="toggle-group">
        <label htmlFor="" className="btn btn-light active toggle-on" ref={elementOn}>
          {onText !== undefined ? onText : 'On'}
        </label>
        <label
          htmlFor=""
          className="btn btn-light toggle-off"
          ref={elementOff}
          style={{width: width}}>
          {offText !== undefined ? offText : 'Off'}
        </label>
        <span className="toggle-handle btn btn-light"></span>
      </div>
    </div>
  )
}

export default Toggle
