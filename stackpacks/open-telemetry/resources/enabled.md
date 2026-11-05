## The Open Telemetry StackPack is installed

### What's next

Instrument one or more applications with Open Telemetry SDKs to generate traces and metrics and install and configure the Open Telemetry collector to send data to SUSE Observability. See the [SUSE Observability Open Telemetry documentation](https://l.stackstate.com/open-telemetry-setup).

To send SDK telemetry through the SUSE Observability Agent, enable its telemetry gateway when installing or upgrading the agent with Helm. The agent's Open Telemetry components are disabled by default:

```bash
--set otel.enabled=true --set otel.telemetryGateway.enabled=true
```

Then point your SDKs at the telemetry gateway service, as described in the [telemetry gateway documentation](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/otel/telemetry-gateway.html).

On Rancher-managed clusters, also set `--set otel.integrations.rancherAgent=true`. Keep the default, `false`, on non-Rancher clusters.

To send data to SUSE Observability a service token is needed.

<button value="/#serviceToken" style="width: 100%">CREATE NEW SERVICE TOKEN</button>
