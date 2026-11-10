# Service Availability

Monitor HTTP and HTTPS endpoints with Prometheus Blackbox Exporter and the SUSE
Observability agent's OpenTelemetry Prometheus scraper. See the
[setup instructions](resources/configuration.md).

The pack creates one `http endpoint` component per URL and an availability
monitor. The optional `availability_service` label groups multiple URLs under an
`availability service` component with its own availability monitor.
**Service Availability → Endpoints** and **Services** provide overviews and
highlight pages. TLS certificate validity is displayed when available. Optional
certificate monitors warn below 14 days and report Critical below two days.

## Telemetry contract

| Input | Purpose |
| --- | --- |
| `probe_success` | Creates endpoint topology and reports whether the endpoint check succeeded. |
| `up` | Creates endpoint topology even when the first scrape fails, and makes availability Critical when a scrape fails. |
| `probe_http_status_code` | Displays the received HTTP status. |
| `probe_duration_seconds` | Displays total probe duration. |
| `probe_ssl_earliest_cert_expiry` | Optional Unix timestamp of the earliest certificate expiration. |
| Datapoint label `service_availability="true"` | Opts a target into this integration. |
| Datapoint label `endpoint` | The complete target URL, including its path and query. |
| Datapoint label `availability_service` | Optional nonempty identity grouping endpoints into a service. |
| Datapoint label `monitor_certificate_expiry="true"` | Opts targets into expiration checks when TLS certificate data is reported, including HTTP redirects to HTTPS. |

These metric and attribute names are the Blackbox/Prometheus contract, not
OpenTelemetry semantic conventions. The Probe example preserves target identity
in `endpoint`, independently of the Prometheus receiver's conversion of `job`
and `instance` to OTel resource attributes.

The Probe example uses `http_2xx`, which follows redirects and requires a final
2xx response. Availability uses the selected module's `probe_success` result.
For an existing exporter, check additional module constraints, such as requiring
HTTPS.

Topology identifier: `urn:service-availability:endpoint/<endpoint URL>`.
The monitor constructs the same identifier from its `endpoint` result label.
The same URL checked by multiple Probes has one component; the monitor uses the
minimum success value, so any failing observation makes it critical. Different
paths or queries are separate endpoints.

Service identifier: `urn:service-availability:service/<availability_service>`.
The service mapping and monitor share this identity, and a hierarchical relation
links the service to every endpoint. Use environment-qualified service identities
when the same service name should represent separate installations.

Availability monitors run every minute against the latest samples, with no
additional five-minute failure delay. Any failing observation makes its endpoint
critical; any failing check labeled for a service makes that service critical.
The two optional certificate monitors run every five minutes and report
Deviating below 14 days remaining and Critical below two days. They use reported
TLS certificate data, including HTTP targets that redirect to HTTPS. Plain HTTP
targets without TLS, missing timestamps and zero timestamps produce no
certificate result.

Components and relations expire after 15 minutes without matching metrics.
The service monitor aggregates observed samples; it cannot detect a missing URL
without an inventory of expected targets. A disappearing URL can reduce the
displayed endpoint count. Missing telemetry is not evidence of successful checks.