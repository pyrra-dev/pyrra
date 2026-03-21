import {type JSX} from 'react'
import {Container} from 'react-bootstrap'

interface FooterProps {
  version?: string
}

const Footer = ({version}: FooterProps): JSX.Element => {
  if (version === undefined || version === '') {
    return <></>
  }

  return (
    <footer className="footer">
      <Container>
        <small>Pyrra {version}</small>
      </Container>
    </footer>
  )
}

export default Footer
