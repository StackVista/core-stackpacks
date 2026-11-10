Waiting for `probe_success` or `up` metrics with `service_availability="true"` and an
`endpoint` label containing the HTTP or HTTPS target URL.

Check that the Probe has `observability.suse.com/agent: scrape`, the OTel
Prometheus scraper and Target Allocator are enabled, and the Blackbox Exporter
service is reachable.
