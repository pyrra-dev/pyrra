import {cn} from '@/lib/utils'

interface SpinnerProps {
  className?: string
}

const Spinner = ({className}: SpinnerProps) => (
  <div
    className={cn('h-8 w-8 animate-spin rounded-full border-2 border-muted border-t-foreground', className)}
    role="status"
  >
    <span className="sr-only">Loading...</span>
  </div>
)

export {Spinner}
