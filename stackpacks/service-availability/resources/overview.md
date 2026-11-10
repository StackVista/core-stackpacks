# Service Availability

Monitor HTTP and HTTPS endpoints with Prometheus Blackbox Exporter and the SUSE
Observability agent's OpenTelemetry Prometheus scraper.

Each endpoint URL becomes a component with a health state. Open **Service
Availability → Endpoints** to see endpoint health, HTTP response codes and probe
duration. Select an endpoint to inspect its health checks and metric history.
HTTPS endpoints also show remaining TLS certificate validity as a duration in
the overview and as whole days in a time-series chart when available.

Use the optional `availability_service` Probe target label to group several URLs
under one service. Open **Service Availability → Services** to see their combined
availability and navigate to the individual endpoints.

The Probe example uses `http_2xx`: it follows redirects and requires a final 2xx
response. Connection failures and invalid or expired TLS certificates also fail
the availability check.

The exporter and Probe resources are installed separately. Topology is created
directly from metrics; no Kubernetes or OpenTelemetry StackPack installation is
required. Certificate expiration checks are optional: opt in with the
`monitor_certificate_expiry` Probe target label to warn below 14 days and report
Critical below two days.
