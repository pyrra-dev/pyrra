import {type JSX} from 'react'

interface FooterProps {
  version?: string
}

const Footer = ({version}: FooterProps): JSX.Element => {
  if (version === undefined || version === '') {
    return <></>
  }

  return (
    <footer className="sticky top-[100vh] py-6 text-center">
      <div className="container-responsive">
        <small className="text-muted-foreground">Pyrra {version}</small>
      </div>
    </footer>
  )
}

export default Footer
