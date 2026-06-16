interface TilesProps {
  children: React.ReactNode
}

const Tiles = (props: TilesProps) => {
  return (
    <div className="grid w-full grid-cols-1 gap-6 sm:grid-cols-2 sm:gap-12 lg:grid-cols-3 lg:gap-[75px]">
      {props.children}
    </div>
  )
}

export default Tiles
