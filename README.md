<p align="center">
  <img src="ui/src/logo.svg" alt="Pyrra: SLOs with Prometheus" height="75">
</p>

<p align="center">Making SLOs with Prometheus manageable, accessible, and easy to use for everyone!
</p>
<p align="center"><img src="docs/screenshot-readme.png" width=700 alt="Screenshot of Pyrra"></p>


## Features

- Support for Kubernetes, Docker, and filesystem
- Alerting: Generates 4 Multi Burn Rate Alerts with different severity
- Page listing all Service Level Objectives
  - All columns sortable
  - Sorted by remaining error budget to see worst ones quickly
  - Tool-tips when hovering for extra context
- Page with details for a Service Level Objective

  - Objective, Availability, Error Budget highlighted as 3 most important numbers
  - Graph to see how the error budget develops over time
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

[@aditya-konarde](https://github.com/aditya-konarde), [@brancz](https://github.com/brancz), [@cbrgm](https://github.com/cbrgm), [@codesome](https://github.com/codesome), [@ekeih](https://github.com/ekeih), [@guusvw](https://github.com/guusvw), [@jzelinskie](https://github.com/jzelinskie), [@kakkoyun](https://github.com/kakkoyun), [@lilic](https://github.com/lilic), [@markusressel](https://github.com/markusressel), [@morremeyer](https://github.com/morremeyer), [@mxinden](https://github.com/mxinden), [@numbleroot](https://github.com/numbleroot), [@paulfantom](https://github.com/paulfantom), [@RiRa12621](https://github.com/RiRa12621), [@tboerger](https://github.com/tboerger), and Maria Franke.

While we were working on Pyrra in private these amazing people helped us with a look of feedback and some even took an extra hour for a in-depth testing! Thank you all so much!

Additionally, [@metalmatze](https://github.com/metalmatze) would like to thank [Polar Signals](https://www.polarsignals.com/) for allowing us to work on this project in his 20% time.

## Demo

Check out our live demo on [demo.pyrra.dev](https://demo.pyrra.dev)!

Feel free to give it a try there!

## Installation

There are pre-build container images available:

```bash
docker pull ghcr.io/pyrra-dev/pyrra:v0.3.3
```

While running Pyrra on its own works there won't be any SLO configured nor will there be any data from a Prometheus to work with.

Therefore, you can find a docker-compose example in [examples/docker-compose](examples/docker-compose).  
This stack comes with Pyrra and Prometheus pre-configured, as well as [some SLOs](examples/docker-compose/pyrra).

## Tech Stack

**Client:** TypeScript with React, Bootstrap, and uPlot.

**Server:** Go with libraries such as: chi, ristretto, xxhash, client-go.

OpenAPI generated API with server (Go) and clients (Go & TypeScript).

## Roadmap

Best to check the [Projects board](https://github.com/pyrra-dev/pyrra/projects) and if you cannot find what you're looking for feel free to open an issue!

## Contributing

Contributions are always welcome!

See [CONTRIBUTING.md](CONTRIBUTING.md) for ways to get started.

Please adhere to this project's `code of conduct`.

## Maintainers

| Name           | Area        | GitHub                                             | Twitter                                             | Company       |
|:---------------|:------------|:---------------------------------------------------|:----------------------------------------------------|:--------------|
| Nadine Vehling | UX/UI       | [@nadinevehling](https://github.com/nadinevehling) | [@nadinevehling](https://twitter.com/nadinevehling) | Grafana Labs  |
| Matthias Loibl | Engineering | [@metalmatze](https://github.com/metalmatze)       | [@metalmatze](https://twitter.com/MetalMatze)       | Polar Signals |

We are mostly maintaining Pyrra in our free time.

## FAQ

#### Why not use Grafana in this particular use case?

Right now we could have used Grafana indeed. In upcoming releases, we plan to add more interactive features to give you better context when coming up with new SLOs. This is something we couldn't do with Grafana.

#### Do I still need Grafana?

Yes, Grafana is an amazing data visualization tool for Prometheus metrics. You can create your own custom dashboards and dive a lot deeper into each component while debugging.

#### Does it work with Thanos too?

Yes, in fact I've been developing this against my little Thanos cluster most of the time.  
The queries even dynamically add headers for downsampling and disable partial responses.

#### How many instances should I deploy?

It depends on the topology of your infrastructure, however, we think that alerting should still happen within each individual Prometheus and therefore running one instance with one Prometheus (pair) makes the most sense. Pyrra itself only needs one instance per Prometheus (pair).

#### Why don't you support more complex SLOs?

For now, we try to accomplish an easy-to-setup workflow for the most common SLOs.
It is still possible to write these more complex SLOs manually and deploy them to Prometheus along those generated.
You can base more complex SLOs on the output of one SLO from this tool.

#### Why is the objective target a string not a float?

[Kubebuilder doesn't support floats in CRDs](https://github.com/kubernetes-sigs/controller-tools/issues/245)...  
Therefore, we need to pass it as string and internally convert it from string to float64.

## Related

Here are some related projects:

* [slok/sloth](https://github.com/slok/sloth)
* [metalmatze/slo-libsonnet](https://github.com/metalmatze/slo-libsonnet)
