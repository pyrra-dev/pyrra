import React from 'react'

interface SizeProps {
  height: number
  width: number
}

interface IconExternalProps extends SizeProps {}

// Font Awesome Pro 6.0.0 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license (Commercial License) Copyright 2022 Fonticons, Inc.

export const IconExternal = ({height, width}: IconExternalProps): JSX.Element => (
  <svg
    width={width}
    height={height}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg">
    <path
      d="M18 13V19C18 19.5304 17.7893 20.0391 17.4142 20.4142C17.0391 20.7893 16.5304 21 16 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V8C3 7.46957 3.21071 6.96086 3.58579 6.58579C3.96086 6.21071 4.46957 6 5 6H11"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M15 3H21V9"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M10 14L21 3"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const IconArrowUp = (): JSX.Element => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M12 19V5"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M5 10L12 3L19 10"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M5 10L12 3L19 10"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const IconArrowDown = (): JSX.Element => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M12 5V19"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M19 14L12 21L5 14"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

export const IconArrowUpDown = (): JSX.Element => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path
      d="M19 14L12 21L5 14"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M19 10L12 3L5 10"
      stroke="black"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

interface IconWarningProps extends SizeProps {
  fill?: string
}

export const IconWarning = ({height, width, fill}: IconWarningProps): JSX.Element => (
  <svg
    width={width}
    height={height}
    viewBox="0 0 24 20"
    fill="none"
    xmlns="http://www.w3.org/2000/svg">
    <path
      d="M10.7809 15.2812C10.7809 15.6045 10.9093 15.9145 11.1379 16.143C11.3664 16.3716 11.6764 16.5 11.9997 16.5C12.3229 16.5 12.6329 16.3716 12.8615 16.143C13.09 15.9145 13.2184 15.6045 13.2184 15.2812C13.2184 14.958 13.09 14.648 12.8615 14.4195C12.6329 14.1909 12.3229 14.0625 11.9997 14.0625C11.6764 14.0625 11.3664 14.1909 11.1379 14.4195C10.9093 14.648 10.7809 14.958 10.7809 15.2812ZM11.1872 7.5625V12.2344C11.1872 12.3461 11.2786 12.4375 11.3903 12.4375H12.609C12.7208 12.4375 12.8122 12.3461 12.8122 12.2344V7.5625C12.8122 7.45078 12.7208 7.35938 12.609 7.35938H11.3903C11.2786 7.35938 11.1872 7.45078 11.1872 7.5625ZM23.2655 18.7344L12.703 0.453125C12.5456 0.181445 12.2739 0.046875 11.9997 0.046875C11.7254 0.046875 11.4512 0.181445 11.2963 0.453125L0.733847 18.7344C0.421542 19.2777 0.812557 19.9531 1.43717 19.9531H22.5622C23.1868 19.9531 23.5778 19.2777 23.2655 18.7344ZM3.37193 18.026L11.9997 3.09121L20.6274 18.026H3.37193Z"
      fill={fill ?? 'black'}
    />
  </svg>
)

interface IconChevronProps extends SizeProps {}

export const IconChevron = ({height, width}: IconChevronProps): JSX.Element => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 320 512" width={width} height={height}>
    <path d="M310.6 233.4c12.5 12.5 12.5 32.8 0 45.3l-192 192c-12.5 12.5-32.8 12.5-45.3 0s-12.5-32.8 0-45.3L242.7 256 73.4 86.6c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0l192 192z" />
  </svg>
)

interface MagnifyingGlassProps extends SizeProps {}

export const IconMagnifyingGlass = ({height, width}: MagnifyingGlassProps): JSX.Element => (
  <svg xmlns="http://www.w3.org/2000/svg" height={height} width={width} viewBox="0 0 512 512">
    <path d="M416 208c0 45.9-14.9 88.3-40 122.7L502.6 457.4c12.5 12.5 12.5 32.8 0 45.3s-32.8 12.5-45.3 0L330.7 376c-34.4 25.2-76.8 40-122.7 40C93.1 416 0 322.9 0 208S93.1 0 208 0S416 93.1 416 208zM208 352a144 144 0 1 0 0-288 144 144 0 1 0 0 288z" />
  </svg>
)

interface IconTableColumnsProps extends SizeProps {}

export const IconTableColumns = ({height, width}: IconTableColumnsProps): JSX.Element => (
  <svg xmlns="http://www.w3.org/2000/svg" height={height} width={width} viewBox="0 0 512 512">
    <path d="M0 96C0 60.7 28.7 32 64 32H448c35.3 0 64 28.7 64 64V416c0 35.3-28.7 64-64 64H64c-35.3 0-64-28.7-64-64V96zm64 64V416H224V160H64zm384 0H288V416H448V160z" />
  </svg>
)

interface IconChartLineProps extends SizeProps {
  fill?: string
}

export const IconChartLine = ({height, width, fill = 'black'}: IconChartLineProps): JSX.Element => (
  <svg xmlns="http://www.w3.org/2000/svg" height={height} width={width} viewBox="0 0 512 512">
    <path
      fill={fill}
      d="M64 64c0-17.7-14.3-32-32-32S0 46.3 0 64V400c0 44.2 35.8 80 80 80H480c17.7 0 32-14.3 32-32s-14.3-32-32-32H80c-8.8 0-16-7.2-16-16V64zm406.6 86.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L320 210.7l-57.4-57.4c-12.5-12.5-32.8-12.5-45.3 0l-112 112c-12.5 12.5-12.5 32.8 0 45.3s32.8 12.5 45.3 0L240 221.3l57.4 57.4c12.5 12.5 32.8 12.5 45.3 0l128-128z"
    />
  </svg>
)

interface IconChartAreaProps extends SizeProps {
  fill?: string
}

export const IconChartArea = ({height, width, fill = 'black'}: IconChartAreaProps): JSX.Element => (
  <svg xmlns="http://www.w3.org/2000/svg" height={height} width={width} viewBox="0 0 512 512">
    <path
      fill={fill}
      d="M64 64c0-17.7-14.3-32-32-32S0 46.3 0 64V400c0 44.2 35.8 80 80 80H480c17.7 0 32-14.3 32-32s-14.3-32-32-32H80c-8.8 0-16-7.2-16-16V64zm96 288H448c17.7 0 32-14.3 32-32V251.8c0-7.6-2.7-15-7.7-20.8l-65.8-76.8c-12.1-14.2-33.7-15-46.9-1.8l-21 21c-10 10-26.4 9.2-35.4-1.6l-39.2-47c-12.6-15.1-35.7-15.4-48.7-.6L135.9 215c-5.1 5.8-7.9 13.3-7.9 21.1v84c0 17.7 14.3 32 32 32z"
    />
  </svg>
)
