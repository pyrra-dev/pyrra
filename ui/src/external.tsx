import {EXTERNAL_GRAFANA_DATASOURCE_ID, EXTERNAL_GRAFANA_ORG_ID, EXTERNAL_URL} from './App'
import {formatDuration} from './duration';

export const buildExternalHRef = (
  queries: string[],
  from: number,
  to: number
): string => {
  const rangeDuration = formatDuration(to - from)
  if (EXTERNAL_GRAFANA_DATASOURCE_ID !== "") {
    const panes = {
      "A": {
        queries: queries.map((query, i) => {
          return {
            refId: `Q${i}`,
            datasource: {
              type: "prometheus",
              uid: EXTERNAL_GRAFANA_DATASOURCE_ID
            },
            expr: query
          }
        }),
        range: {
          from: `now-${rangeDuration}`,
          to: "now"
        }
      }
    }
    const encodedQuery = encodeURIComponent(JSON.stringify(panes));
    return `${EXTERNAL_URL}/explore?schemaVersion=1&panes=${encodedQuery}&orgId=${EXTERNAL_GRAFANA_ORG_ID}`
  } else {
    const queryParams = queries.map((query, i) => {
      const encodedQuery = encodeURIComponent(query);
      return `g${i}.expr=${encodedQuery}&g${i}.range_input=${rangeDuration}&g${i}.tab=0`;
    }).join("&");
    return `${EXTERNAL_URL}/graph?${queryParams}`
  }
}

export const externalName = () => {
  if (EXTERNAL_GRAFANA_DATASOURCE_ID !== "") {
    return "Grafana"
  } else {
    return "Prometheus"
  }
}
