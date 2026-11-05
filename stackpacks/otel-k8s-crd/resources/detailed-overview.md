## How it works

The StackPack defines SUSE Observability settings (component and relation mappings) that the platform's OTel collector applies to telemetry produced by the `k8scrdreceiver` OpenTelemetry receiver, which runs in the SUSE Observability Agent's k8s CRD collector. The receiver discovers CRDs and their CR instances using watch mode (real-time informer events) plus periodic snapshots (full-state emission from the informer cache, controlled by `snapshotInterval`).

### Components created

- **Custom Resource Definition** - Each CRD installed in the cluster (e.g., `helmcharts.helm.cattle.io`)
- **Custom Resource** - Each CR instance (e.g., a specific HelmChart resource)

### Relations created

- **instance of** - Links each Custom Resource to its defining Custom Resource Definition

### Data sources

- CRD and CR discovery via the Kubernetes API server
- Watch mode for real-time change detection
- Periodic snapshots from the informer cache for state reconciliation
