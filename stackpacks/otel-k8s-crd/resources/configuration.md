## Prerequisites

- SUSE Observability Agent installed with Helm
- Network connectivity from the agent to the platform OTLP ingest endpoint

## Enable the k8s resource collector

The k8s resource collector is part of the agent's OpenTelemetry components, which are disabled by default. Enable them with `otel.enabled`; the collector itself is enabled by default once OpenTelemetry is on:

```yaml
otel:
  enabled: true
```

Or, when installing or upgrading the agent with Helm:

```bash
--set otel.enabled=true
```

## Select Custom Resource API groups

CRDs are always collected. Custom Resource instances are only collected for selected API groups. Enabled integration presets add common SUSE-related API groups, such as Kubewarden, SUSE Runtime Enforcer and SUSE Virtualization. Add more with `otel.k8sResourceCollector.crDiscovery.apiGroups.include`:

```yaml
otel:
  enabled: true
  k8sResourceCollector:
    crDiscovery:
      discoveryMode: api_groups
      apiGroups:
        include:
          "policies.kubewarden.io": true
          "kubevirt.io": true
        exclude:
          "internal.example.com": true
```

Set an included API group to `false` to disable an integration-provided default. Set `discoveryMode: all` to collect CRs for every CRD API group; `apiGroups` filters are then ignored.

When `otel.k8sResourceCollector.rbac.useWildcard: false`, the truthy `crDiscovery.apiGroups.include` entries are also used for restricted RBAC. Kubernetes RBAC only supports exact API groups or `"*"`, so wildcard patterns like `"*.example.com"` require `rbac.useWildcard: true`.

For the full configuration guide, see the [k8s resource collector documentation](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/otel/k8s-resource-collector.html).
