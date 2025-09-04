# Getting Started with Create React App

This project was bootstrapped with [Create React App](https://github.com/facebook/create-react-app).

## Pyrra UI Development Workflow

### Two UI Serving Methods

Pyrra uses **two different UI serving methods** that require different workflows:

#### Development UI (Recommended for Development)
- **Command**: `npm start`
- **URL**: http://localhost:3000
- **Source**: Live source files from `src/`
- **Updates**: Real-time hot reload
- **API**: Configure via `public/index.html` (`window.API_BASEPATH`)

#### Embedded UI (Production)
- **Command**: `../pyrra api` (Go binary)
- **URL**: http://localhost:9099
- **Source**: Compiled files from `build/` (via Go embed)
- **Updates**: Requires rebuild workflow
- **API**: Built into Go binary

### 🚨 Critical: Complete UI Change Workflow

**❌ Common Mistake**: Testing only development UI and assuming embedded UI will work

**✅ Required Steps for UI Changes**:
```bash
# 1. Make changes to src/ files
# 2. Test in development
npm start  # → http://localhost:3000

# 3. Build for production (REQUIRED)
npm run build

# 4. Rebuild Go binary (REQUIRED)
cd ..
make build

# 5. Restart Pyrra service
# 6. Test production at http://localhost:9099
```

**Why This Matters**: Production users only see embedded UI. Development success ≠ Production success.

## Available Scripts

In the project directory, you can run:

### `npm start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### `npm test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `npm run build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `npm run eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.

## Learn More

You can learn more in the [Create React App documentation](https://facebook.github.io/create-react-app/docs/getting-started).

To learn React, check out the [React documentation](https://reactjs.org/).
