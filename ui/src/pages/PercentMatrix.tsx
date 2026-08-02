// TEMPORARY — a bench for PercentValue, not a product page. Delete when the
// component has settled.
//
// Route: /objectives/percent-matrix
//
// Container width runs along one axis in em rather than px, deliberately: the
// breakpoints are in em too, so every column should render the same precision at
// every font size. A column that disagrees between rows means the em thresholds
// aren't holding and something is resolving against the wrong font.

import React, {type JSX, useState} from 'react'
import {Link} from 'react-router-dom'
import Navbar from '../components/Navbar'
import PercentValue from '../components/tiles/PercentValue'
import {formatTargetPercent} from '../percent'

const VALUES: Array<{label: string; value: number}> = [
  {label: 'target 99%', value: 0.99},
  {label: 'target 99.9%', value: 0.999},
  {label: 'long decimals', value: 0.9876543},
  {label: 'full', value: 1},
  {label: 'low availability', value: 0.3308},
  {label: 'sub-1%', value: 0.005},
  {label: 'budget breached', value: -5.692},
  {label: 'budget very breached', value: -12.345678},
]

const CONTEXTS: Array<{label: string; className: string}> = [
  {label: 'Tile', className: 'text-[40px]'},
  {label: 'Heading', className: 'text-2xl'},
  {label: 'Body', className: 'text-base'},
  {label: 'List cell', className: 'text-sm'},
  {label: 'Dense', className: 'text-xs'},
]

// Widths in em, so they mean the same thing at every font size.
const WIDTHS = [2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5, 7, 7.5, 8]

const Cell = ({value, em}: {value: number; em: number}): JSX.Element => (
  <td className="border border-border p-1 align-bottom">
    <div style={{width: `${em}em`}} className="outline outline-dashed outline-border/60">
      <PercentValue value={value} />
    </div>
  </td>
)

const PercentMatrix = (): JSX.Element => {
  const [playgroundValue, setPlaygroundValue] = useState(0.9876543)

  return (
    <>
      <Navbar>
        <div>
          <Link to="/">Objectives</Link> &gt; <span>PercentValue matrix</span>
        </div>
      </Navbar>

      <div className="mt-[100px]">
        <div className="container-responsive pb-24">
          <h3>PercentValue matrix</h3>
          <p className="mb-8 max-w-prose text-sm text-muted-foreground">
            Each cell is a fixed-width container (dashed outline) holding the component. Widths are
            in <code>em</code>, matching the units the breakpoints use — so a column should show the
            same number of decimals in every block below, whatever the font size. The value is
            rendered at 1, 3 and 5 decimals; CSS picks whichever fits.
          </p>

          <section className="mb-12">
            <h4>Drag to resize</h4>
            <p className="mb-3 text-sm text-muted-foreground">
              The handle is in the bottom-right corner of the box.
            </p>
            <div className="mb-3 flex flex-wrap gap-2">
              {VALUES.map((v) => (
                <button
                  key={v.label}
                  onClick={() => { setPlaygroundValue(v.value) }}
                  className="rounded-md border border-border px-2 py-1 text-xs hover:bg-muted">
                  {v.label}
                </button>
              ))}
            </div>
            <div className="resize-x overflow-auto rounded-md border border-border p-4 text-[40px]" style={{width: '20rem'}}>
              <PercentValue value={playgroundValue} />
            </div>
          </section>

          {CONTEXTS.map((ctx) => (
            <section key={ctx.label} className="mb-12">
              <h4>
                {ctx.label} <span className="text-sm text-muted-foreground">({ctx.className})</span>
              </h4>
              <div className="overflow-x-auto">
                <table className={`${ctx.className} border-collapse`}>
                  <thead>
                    <tr>
                      <th className="border border-border p-1 text-left text-xs font-medium text-muted-foreground">
                        value
                      </th>
                      {WIDTHS.map((w) => (
                        <th
                          key={w}
                          className="border border-border p-1 text-left text-xs font-medium text-muted-foreground">
                          {w}em
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {VALUES.map((v) => (
                      <tr key={v.label}>
                        <th className="border border-border p-1 text-left text-xs font-normal whitespace-nowrap text-muted-foreground">
                          {v.label}
                          <br />
                          <code>{formatTargetPercent(v.value)}%</code>
                        </th>
                        {WIDTHS.map((w) => (
                          <Cell key={w} value={v.value} em={w} />
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          ))}
        </div>
      </div>
    </>
  )
}

export default PercentMatrix
