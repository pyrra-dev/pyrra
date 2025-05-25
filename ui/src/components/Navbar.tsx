import {ReactNode} from 'react'
import {Col, Container, Navbar as BootstrapNavbar, ButtonGroup, ToggleButton} from 'react-bootstrap'
import {Link} from 'react-router-dom'
import logo from '../logo.svg'
import {useTheme} from '../ThemeContext'
import {IconSun, IconMoon, IconSystem} from './Icons'

interface NavbarProps {
  children?: ReactNode
}

const Navbar = ({children}: NavbarProps): JSX.Element => {
  return (
    <BootstrapNavbar className={children !== undefined ? 'navbar-tall' : ''}>
      {children !== undefined ? (
        <Container>
          <Col className="col-xxxl-10 offset-xxxl-1">
            <div className="breadcrumb">{children}</div>
          </Col>
        </Container>
      ) : (
        <></>
      )}
      <Link to="/" className="logo">
        <img src={logo} alt="" height={40} />
      </Link>
      <ThemeToggle />
    </BootstrapNavbar>
  )
}

const ThemeToggle = (): JSX.Element => {
  const { mode, setMode } = useTheme()
  
  const themeOptions = [
    { name: 'Light', value: 'light' as const, icon: <IconSun size={16} /> },
    { name: 'Dark', value: 'dark' as const, icon: <IconMoon size={16} /> },
    { name: 'System', value: 'system' as const, icon: <IconSystem size={16} /> }
  ]
  
  return (
    <ButtonGroup className="theme-toggle">
      {themeOptions.map((option) => (
        <ToggleButton
          key={option.value}
          id={`theme-${option.value}`}
          type="radio"
          variant="outline-secondary"
          name="theme"
          value={option.value}
          checked={mode === option.value}
          onChange={(e) => setMode(e.currentTarget.value as 'light' | 'dark' | 'system')}
          size="sm"
          className={mode === option.value ? 'active' : ''}
        >
          <span style={{ display: 'flex', alignItems: 'center', gap: '0.375rem' }}>
            {option.icon}
            {option.name}
          </span>
        </ToggleButton>
      ))}
    </ButtonGroup>
  )
}

export default Navbar
