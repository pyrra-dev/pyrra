import {type JSX, type ReactNode} from 'react'
import {Link} from 'react-router-dom'
import logo from '../logo.svg'

interface NavbarProps {
  children?: ReactNode
}

const Navbar = ({children}: NavbarProps): JSX.Element => {
  return (
    <nav
      className={`relative flex bg-muted ${
        children !== undefined
          ? 'h-28 items-end lg:h-14 lg:items-center'
          : 'h-14 items-center'
      }`}
    >
      {children !== undefined ? (
        <div className="container-responsive">
          <div className="3xl:mx-auto 3xl:w-10/12">
            <div className="text-lg [&_a]:text-muted-foreground [&_a]:no-underline hover:[&_a]:text-foreground [&_span]:font-bold [&_span]:text-foreground">
              {children}
            </div>
          </div>
        </div>
      ) : null}
      <Link
        to="/"
        className="absolute left-1/2 top-2 -translate-x-1/2"
      >
        <img src={logo} alt="" height={40} />
      </Link>
    </nav>
  )
}

export default Navbar
