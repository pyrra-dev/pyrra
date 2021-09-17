import { ReactNode } from 'react'
import { Container, Navbar as BootstrapNavbar } from 'react-bootstrap'
import { Link } from 'react-router-dom'
import logo from '../logo.svg'

interface NavbarProps {
  children?: ReactNode
}

const Navbar = ({ children }: NavbarProps): JSX.Element => {
  return (
    <BootstrapNavbar className={children !== undefined ? 'navbar-tall' : ''}>
      {children !== undefined ?
        <Container>
          <div className="breadcrumb">{children}</div>
        </Container> :
        <></>
      }
      <Link to="/" className="logo">
        <img src={logo} alt="" height={40}/>
      </Link>
    </BootstrapNavbar>
  )
}

export default Navbar
