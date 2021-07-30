import { ReactNode } from 'react'
import { Container, Navbar as BNavbar } from 'react-bootstrap'
import { Link } from 'react-router-dom'
import logo from '../logo.svg'

interface NavbarProps {
  children: ReactNode
}

const Navbar = ({ children }: NavbarProps): JSX.Element => {
  return (
    <BNavbar>
      <Container>
        <div className="breadcrumb">{children}</div>
        <Link to="/" className="logo">
          <img src={logo} alt="" height={40}/>
        </Link>
      </Container>
    </BNavbar>
  )
}

export default Navbar
