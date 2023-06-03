interface TilesProps {
  children: React.ReactNode
}

const Tiles = (props: TilesProps) => {
  return <div className="tiles">{props.children}</div>
}

export default Tiles
