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

## Using Grafana Instead of Prometheus for External Links

If you prefer to use Grafana Explore for viewing metrics instead of Prometheus, you can modify the `docker-compose.yaml` file to use Grafana external URL. Replace the `pyrra-api` service command with:

```yaml
command:
  - api
  - --prometheus-url=http://prometheus:9090
  - --grafana-external-url=http://your-grafana-instance:3000
  - --grafana-external-datasource-id=your-prometheus-datasource-uid
  - --api-url=http://pyrra-filesystem:9444
```

Make sure to replace:
- `http://your-grafana-instance:3000` with your actual Grafana URL
- `your-prometheus-datasource-uid` with the UID of your Prometheus datasource in Grafana

### Finding Your Grafana Datasource UID

1. Log into your Grafana instance
2. Go to Configuration → Data Sources
3. Click on your Prometheus datasource
4. The UID is shown in the settings (or can be found in the URL)

### Complete Example with Grafana

We've included a `docker-compose-with-grafana.yaml` file that demonstrates a complete setup with Prometheus, Grafana, and Pyrra configured to use Grafana for external links.

To use it:
1. Start all services: `docker-compose -f docker-compose-with-grafana.yaml up -d`
2. Access Grafana at http://localhost:3000 (default login: admin/admin)
3. Add Prometheus as a datasource:
   - URL: `http://prometheus:9090`
   - Save and note the datasource UID
4. Update the `docker-compose-with-grafana.yaml` file with the actual datasource UID
5. Restart Pyrra API: `docker-compose -f docker-compose-with-grafana.yaml restart pyrra-api`
