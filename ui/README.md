# Pyrra UI

The Pyrra UI is a React app built with [Vite](https://vite.dev/) and tested with [Vitest](https://vitest.dev/).

## Available Scripts

In the `ui` directory, run `npm install` first, then:

### `npm start`

Runs the app in development mode with the Vite dev server.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

API requests are proxied to [demo.pyrra.dev](https://demo.pyrra.dev) (see `vite.config.ts`).
The page reloads when you make edits.

### `npm test`

Runs the test suite once with Vitest.

### `npm run lint`

Lints the sources in `src` with ESLint. Use `npm run fix` to apply automatic fixes.

### `npm run build`

Checks that the generated CRD types are up to date, type-checks the code and builds the app for production into the `build` folder.
The Go binary embeds this folder, so `make ui/build` from the repository root runs this step.

### `npm run generate:crd-types`

Regenerates the TypeScript types for the Pyrra CRDs (`make generate` also runs this).

## Learn More

- [Vite documentation](https://vite.dev/guide/)
- [Vitest documentation](https://vitest.dev/guide/)
- [React documentation](https://react.dev/)
