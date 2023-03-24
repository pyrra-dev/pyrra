# How to Contribute

If you want to add a new feature please [open an issue](https://github.com/pyrra-dev/pyrra/issues/new) first.

## Git Workflow

A good example on how to setup your repository for development can be found on the
[Kubernetes Github Workflow](https://github.com/kubernetes/community/blob/master/contributors/guide/github-workflow.md) page.

## Dependencies

Make sure you have a recent version of:

- [Go](https://go.dev/doc/install) (1.17+)
- [npm](https://docs.npmjs.com/cli/v8/configuring-npm/install)

In order to build the UI you'll also need

```
npm install react-scripts
```

Then install all dependencies

```bash
make install
```

## Building

Please check out the [Makefile](Makefile) on which targets are available to test
and build the project.

## Running locally

Build the UI and compile the Go binaries

```bash
make all
```

### Run the API and UI

Run the API binary in one terminal

```bash
./pyrra api
```

*Note: the API assumes a Prometheus is running on [localhost:9090](http://localhost:9090) and a backend on [localhost:9444](http://localhost:9444)) by default. Check  `./bin/api --help` flag for the parameters to change those.*

### Run a Kubernetes or filesystem backend

Run the filesystem binary in another terminal

```bash
./pyrra filesystem
```

Or run the Kubernetes binary in the other terminal

```bash
./pyrra kubernetes
```

*Note: This binary tries to run against your default Kubernetes context. Use the `-kubeconfig` flag to change for another kubeconfig*

### Running the UI standalone

Run the Node server to work on the UI itself

```bash
cd ui
npm run start
```

*Note: This still needs the API and one of the backends to really work.*

Most likely you need to update the `window.PUBLIC_API` constant in `ui/public/index.html`.

```diff
-    <script>window.PUBLIC_API = '/'</script>
+    <script>window.PUBLIC_API = 'http://localhost:9099/'</script>
```
