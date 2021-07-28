# Pyrra

Pyrra is the ultimate Service Level Objectives (SLOs) tool for Prometheus presenting the most important facts at a glance!

## Features

- Support for Kubernetes, Docker, and filesystem
- Alerting: Generates 4 Multi Burn Rate Alerts with different severity
- Page listing all Service Level Objectives
  - All columns sortable
  - Sorted by remaining error budget to see worst ones quickly
  - Tool-tips when hovering for extra context
- Page with details for a Service Level Objective

  - Objective, Availability, Error Budget highlighted as 3 most important numbers
  - Error budget graph to see how it develops over time
  - Time range picker to change graphs
  - Request, Errors, Duration (RED) graphs for the underlying service
  - Multi Burn Rate Alerts overview table
- Caching of Prometheus query results
- Thanos: Disabling of partial responses and downsampling to 5m and 1h
- OpenAPI generated API

## Feedback & Support

If you have any feedback, please open a discussion in the GitHub Discussions of this project.  
We would love to learn what you think!

## Acknowledgements

[Aditya Konarde](https://github.com/aditya-konarde), [Christian Bargmann](https://github.com/cbrgm), [Frederic Branczyk](https://github.com/brancz), [Ganesh Vernekar](https://github.com/codesome), [Guus van Weelden](https://github.com/guusvw), [Jimmy Zelinskie](https://github.com/jzelinskie), [Kemal Akkoyun](https://github.com/kakkoyun), [Lennart Oldenburg](https://github.com/numbleroot), [Lili Cosic](https://github.com/lilic), [Maria Franke](), [Markus Ressel](https://github.com/markusressel), [Max Inden](https://github.com/mxinden), [Max Rosin](https://github.com/ekeih), [Morre Meyer](https://github.com/morremeyer), [Paweł Krupa](https://github.com/paulfantom), [Rick Rackow](https://github.com/RiRa12621), [Thomas Boerger](https://github.com/tboerger)

While [Nadine](https://github.com/nadinevehling) and [Matthias](https://github.com/metalmatze) were working on Pyrra in private these amazing people helped us with a look of feedback and some even took an extra hour for a in-depth testing! Thank you all so much! 

## Installation

Install Pyrra

```bash
# TODO
```

## Documentation

[Documentation](https://linktodocumentation)


## Tech Stack

**Client:** TypeScript with React, Bootstrap, Recharts

**Server:** Go with libraries such as: chi, ristretto, xxhash, client-go.

OpenAPI generated API with server (Go) and clients (Go & TypeScript).


## Run Locally

You need to have [Go](https://golang.org/) and [Node](https://nodejs.org/en/download/) installed.

Clone the project

```bash
git clone https://github.com/pyrra-dev/pyrra.git
```

Go to the project directory

```bash
cd pyrra
```

Install dependencies

```bash
make install
```

Build the UI and compile the Go binaries

```bash
make
```

### Run the API and UI

Run the API binary in one terminal

```bash
./bin/api
```

*Note: the API assumes a Prometheus is running on [localhost:9090](http://localhost:9090) and a backend on [localhost:9444](http://localhost:9444)) by default. Check  `./bin/api --help` flag for the parameters to change those.*

### Run a Kubernetes or filesystem backend
Run the filesystem binary in another terminal

```bash
./bin/filesystem
```

Or run the Kubernetes binary in the other terminal

```bash
./bin/kubernetes
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



## Roadmap

Best to check the [Projects board](https://github.com/pyrra-dev/pyrra/projects/1) and if you cannot find what you're looking for feel free to open an issue!

## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

Please adhere to this project's `code of conduct`.

## FAQ

#### Why not simply use Grafana?

Right now we could have simply used Grafana indeed. In the next releases we plan to add more interactive features to give you a lot better context when coming up with new SLOs. This is something we couldn't do with Grafana.

#### Do I still need Grafana?

Yes, Grafana is still an amazing visualization tool for Prometheus metrics. You can create your own custom dashboards and dive a lot deeper into each component while debugging.

#### Does it work with Thanos too?

Yes, in fact I've been developing this against my little Thanos cluster most of the time.  
The queries even dynamically add headers for downsampling and disable partial responses.

#### How many instances should I deploy?

It obviously depends on the topology of your infrastructure, however, we think that alerting should still happen within each individual Prometheus and therefore running one instance with one Prometheus (pair) makes most sense.

#### Why don't you support more complex SLOs?

For now, we try to accomplish and easy to set up workflow for the most common SLOs.
It is still possible to write these more complex SLOs manually and deploy them to Prometheus along those generated.
It probably makes sense to base your more complex SLOs on the output of one SLO from this tool.

#### Why is the objective target a string not a float?

[Kubebuilder doesn't support floats in CRDs](https://github.com/kubernetes-sigs/controller-tools/issues/245)...  
Therefore, we need to pass it as string and internally convert it from string to float64.

## Related

Here are some related projects:

* [slok/sloth](https://github.com/slok/sloth)
* [metalmatze/slo-libsonnet](https://github.com/metalmatze/slo-libsonnet)
