# Configure service availability checks

The data flow is:

```text
Probe → Target Allocator → OTel Prometheus scraper → Blackbox Exporter
                               ↓                       ↓
                        SUSE Observability        HTTP(S) endpoint
```

The Target Allocator discovers the Probe and gives the scraper its scrape job.
The scraper calls Blackbox Exporter's `/probe` endpoint with the target URL.
Blackbox Exporter performs the check and returns metrics, which the scraper
forwards to SUSE Observability. This StackPack turns those metrics into endpoint
components and health checks. It also uses the scraper's `up` metric to detect
failed scrapes, including exporter connection errors, scrape timeouts and unknown
modules.

## Install Blackbox Exporter

Use the **Prometheus Blackbox Exporter** chart from **SUSE Application Collection
(AppCo)**, available at
[SUSE Application Collection](https://apps.rancher.io/). Follow Application
Collection's access instructions to authenticate Helm and create its image pull
secret in the `monitoring` namespace.

Save the following as `blackbox-values.yaml`. Set `appco-credentials` to the name
of your Application Collection image pull secret. The `http_2xx` module follows
redirects and considers the endpoint healthy when the final response is 2xx.
It supports HTTP and HTTPS and verifies TLS certificates.

```yaml
global:
  imagePullSecrets:
    - appco-credentials
config:
  modules:
    http_2xx:
      prober: http
      timeout: 5s
      http:
        valid_http_versions: ["HTTP/1.0", "HTTP/1.1", "HTTP/2.0"]
        follow_redirects: true
        fail_if_not_ssl: true
```

Install through your deployment tooling, or use Helm:

```bash
helm upgrade --install prometheus-blackbox-exporter \
  oci://dp.apps.rancher.io/charts/prometheus-blackbox-exporter \
  --namespace monitoring --create-namespace \
  --values blackbox-values.yaml
```

If you already have an exporter, reuse its `http_2xx` module after checking its
settings. For example, `fail_if_not_ssl: true` requires HTTPS. Preserve modules
used by existing checks when updating its configuration.

## Enable Probe discovery

Merge these overrides into your existing SUSE Observability agent installation:

```yaml
otel:
  enabled: true
  prometheusScraping:
    monitorCrds:
      enabled: true
```

Prometheus scraping is enabled by default. If your existing values explicitly
disable it, also set `otel.prometheusScraping.enabled: true`.
The chart's default Probe selector matches `observability.suse.com/agent: scrape`,
as used in the example below, and its default namespace selector includes all
namespaces. If your installation overrides these selectors, use matching Probe
labels and include the `monitoring` namespace.

`monitorCrds.enabled: true` installs the ServiceMonitor, PodMonitor and Probe
CRDs. Omit this override when those CRDs are already managed by another chart
or deployment tool.

The scraper must reach the exporter, and the exporter must reach each endpoint, including DNS
and HTTPS connections. An internal endpoint is supported when it is reachable
from the exporter's network.

## Create a Probe

Create this resource through your Kubernetes deployment tooling.
**Replace all uppercase `REPLACE_WITH_...` values and the example URLs before creating the Probe.** Keep the label
name `availability_service`; replace its value with your own service identity.
Multiple targets can be listed under `static`; their full URLs identify separate
endpoint components. The example groups both URLs under one service and enables
optional certificate expiration warnings.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Probe
metadata:
  name: REPLACE_WITH_YOUR_PROBE_NAME
  namespace: monitoring
  labels:
    observability.suse.com/agent: scrape
spec:
  jobName: REPLACE_WITH_YOUR_JOB_NAME
  interval: 60s
  scrapeTimeout: 10s
  module: http_2xx
  prober:
    url: prometheus-blackbox-exporter.monitoring.svc.cluster.local:9115
    scheme: http
    path: /probe
  targets:
    staticConfig:
      static:
        - https://example.com/health
        - https://example.com/api/ready
      labels:
        service_availability: "true"
        availability_service: REPLACE_WITH_YOUR_SERVICE_NAME
        monitor_certificate_expiry: "true"
      relabelingConfigs:
        - sourceLabels: [__param_target]
          targetLabel: endpoint
```

| Field | What to change |
| --- | --- |
| `metadata.name` | Replace `REPLACE_WITH_YOUR_PROBE_NAME` with a valid lowercase Kubernetes resource name, such as `checkout-health`. |
| `spec.jobName` | Replace `REPLACE_WITH_YOUR_JOB_NAME` with any descriptive scrape job name. |
| `targets.staticConfig.static` | Replace `https://example.com/health` and `https://example.com/api/ready` with your complete HTTP or HTTPS URLs, including the correct paths. Add or remove URLs as needed. |
| `targets.staticConfig.labels.availability_service` | Replace `REPLACE_WITH_YOUR_SERVICE_NAME` with a service identity such as `production/my-service`. Remove this label if you do not want service grouping. |
| `metadata.namespace` and `spec.prober.url` | Adjust if your Probe or exporter uses a different namespace, or the exporter has a different service name. |
| `targets.staticConfig.labels.monitor_certificate_expiry` | Keep `"true"` to enable certificate checks; set `"false"` or remove the label to disable them. |

The relabeling copies the complete target URL into the `endpoint` metric label.
Keep this relabeling and the `service_availability: "true"` label so the
StackPack can discover your endpoints. For an existing Probe, add both to its
`targets.staticConfig` and select the exporter's `http_2xx` module.
The interval can be changed; keep it well below the 15-minute topology expiry.

**Do not put credentials in a target URL or its query string**: that URL is stored
as a metric label and shown in the UI. Configure authentication separately in
the exporter's HTTP module when an endpoint requires it.

## Monitor several URLs for one service

Set `availability_service` to the same nonempty value on the targets of each
Probe belonging to a service. One Probe can list several URLs, as above. Separate
Probes can use the same service label when URLs need different modules, scrape
intervals, or certificate warning settings.

The StackPack creates a service component and links it to each endpoint. Open
**Service Availability → Services**, then select a service to inspect its related
endpoints. The service availability monitor becomes critical when any of its
currently observed endpoint checks or scrapes fails. Each URL also retains its own
availability monitor and highlight page.

Use a unique service identity across your monitored environments, for example
`production/checkout` and `staging/checkout`, if they should be separate services.
Omit `availability_service` to monitor URLs individually without creating a
service component. Every URL assigned to a service contributes to availability;
monitor optional dependencies separately.

The service monitor uses available `probe_success` and `up` samples, rather than
an inventory of expected URLs. A failed scrape reports `up=0`, so its URL remains
in the monitored endpoint count and makes availability Critical. If the scraper
stops emitting both metrics entirely for a URL, its absence is not automatically
counted as a failure. Services and relations expire 15 minutes after their last
matching metric.

## Optional certificate expiration warnings

To enable both certificate monitors, add `monitor_certificate_expiry: "true"`
to the Probe's `targets.staticConfig.labels`. Omit the label or set it to
`"false"` to disable certificate checks for those targets.

Both monitors evaluate every five minutes. **Endpoint certificate expiring** reports
**Deviating** below 14 days. **Endpoint certificate expiration imminent** reports
**Critical** below two days.

The **Certificate expires in** overview column shows the remaining validity of
the certificate chain. The **TLS certificate** highlight chart shows whole days,
rounded down, with lines marking the warning and critical thresholds. Certificate
data is also available for HTTP endpoints that redirect to HTTPS.

To change the thresholds, edit the monitors in the Monitors UI. Values are in
seconds: 14 days is `1209600`, and two days is `172800`. Update the chart's
threshold lines to match.

An expired or invalid certificate fails the availability check even when
certificate warnings are disabled. Checks without certificate data show no
expiration value and do not trigger the certificate monitors.

## Check your endpoints

Open **Service Availability → Endpoints** to check that your URLs appear with
their health, HTTP response codes and probe durations. Select an endpoint to
inspect its health checks and metric history. A failed check or scrape makes
its availability monitor Critical on the next evaluation.

Open **Service Availability → Services** to see grouped services and their
endpoint counts. Select a service to inspect its related endpoints.

## Troubleshooting

1. **Endpoints or services are missing.** In the Metrics Explorer, run:

   ```promql
   {__name__=~"probe_success|up",service_availability="true"}
   ```

   Each result needs an `endpoint` label containing the full HTTP or HTTPS URL.
   For service grouping, it also needs a nonempty `availability_service` label.
   If labels are missing, check the Probe's target labels and relabeling.
   If no metrics arrive, check Probe discovery, the Target Allocator and scraper
   logs. Components expire after 15 minutes without data.

2. **An endpoint is Critical.** Check its availability monitor:
   `probe_success=0` means the endpoint check failed; `up=0` means the scraper
   could not collect the probe result. Check the endpoint's application, network
   path and TLS certificate for a failed probe. For a failed scrape, check the
   exporter connection and module name. If availability is healthy, check
   whether a certificate expiration monitor is reporting the problem.

3. **Test the exporter directly.** Port-forward its service:

   ```bash
   kubectl -n monitoring port-forward svc/prometheus-blackbox-exporter 9115:9115
   ```

   In another terminal, replace the target URL with your endpoint and use the
   same module as your Probe:

   ```bash
   curl --get 'http://localhost:9115/probe' \
     --data-urlencode 'module=http_2xx' \
     --data-urlencode 'target=https://example.com/health' \
     --data-urlencode 'debug=true'
   ```

   `Unknown module` means the exporter has no module with that name; configure
   it or select an existing module. A completed probe returns metrics even when
   the endpoint is unhealthy. Use the debug output to identify HTTP, connection
   or TLS failures.

   If this check succeeds but metrics are missing in SUSE Observability, check
   the Probe selector, Target Allocator, scraper logs and connectivity from the
   scraper to the exporter.
