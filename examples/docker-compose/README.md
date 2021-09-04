# Pyrra docker-compose Example

This example starts Pyrra and Prometheus in 3 containers. One Prometheus container and two for Pyrra.  

The first container is Pyrra with the UI and API and the second container is the Pyrra filesystem backend.
This filesystem backend loads the SLO configs from the `pyrra/` folder and then generates the Prometheus recording and alerting rules.
Additionally, does the filesystem backend return the availability SLOs to the API/UI container.

## Running the Example

To run the example docker-compose setup you need to have Docker and docker-compose installed.

```bash
docker-compose up -d
```

Now, this should pull down all 2 container images for Pyrra and Prometheus.
Next the 3 containers are started. 

Pyrra is available on [localhost:9099](http://localhost:9099) and Prometheus at [localhost:9090](http://localhost:9090).

Pyrra should show you the available SLOs on its overview page and you can click on the individual ones to see the specific SLO and all it's details.

Prometheus is configured by Pyrra with the necessary [recording rules](http://localhost:9090/rules) and [alerting rules](http://localhost:9090/alerts).
