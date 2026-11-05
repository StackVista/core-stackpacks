## Installation

You can click on `Install` to install the Open Telemetry StackPack. Then follow the instructions here and in [the documentation](https://l.stackstate.com/open-telemetry-setup) to finish the Open Telemetry setup.

The SUSE Observability Agent's Open Telemetry components are disabled by default. To receive traces and metrics pushed over OTLP from application SDKs, enable the agent's telemetry gateway when installing or upgrading the agent with Helm:

```yaml
otel:
  enabled: true
  telemetryGateway:
    enabled: true
```

or

```bash
--set otel.enabled=true --set otel.telemetryGateway.enabled=true
```

Then point your SDKs at the telemetry gateway service, as described in the [telemetry gateway documentation](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/otel/telemetry-gateway.html).

On Rancher-managed clusters, also set `--set otel.integrations.rancherAgent=true` to enrich emitted logs with Rancher Manager URL and Harvester cluster ID metadata. Keep the default, `false`, on non-Rancher clusters.
