## The Open Telemetry Kubernetes CRD StackPack is installed

### What's next

Explore the discovered Custom Resources and their relationships in SUSE Observability.

### No data

Make sure the k8s resource collector is enabled in the SUSE Observability Agent. It is part of the agent's OpenTelemetry components, which are disabled by default:

```yaml
otel:
  enabled: true
```

Or, when installing or upgrading the agent with Helm:

```bash
--set otel.enabled=true
```

CRDs are always collected. Custom Resource instances are only collected for selected API groups. Add more with `otel.k8sResourceCollector.crDiscovery.apiGroups.include`, or set `discoveryMode: all` to collect CRs for every CRD API group:

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
```

When `otel.k8sResourceCollector.rbac.useWildcard: false`, the truthy `crDiscovery.apiGroups.include` entries are also used for restricted RBAC. Kubernetes RBAC only supports exact API groups or `"*"`, so wildcard patterns like `"*.example.com"` require `rbac.useWildcard: true`.

For the full configuration guide, see the [k8s resource collector documentation](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/otel/k8s-resource-collector.html).
