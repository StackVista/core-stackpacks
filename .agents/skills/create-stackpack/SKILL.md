---
name: create-stackpack
description: Create or extend SUSE Observability StackPacks for a product or technology using OpenTelemetry data, semantic conventions, mappings, presentations, and monitors.
---

# Create a StackPack

Start with the product or technology the user wants to observe.
Use the current declarative scaffold and adapt it to that integration.

## Sources

Read the official [creation guide](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/custom-integrations/create.html) before starting.
Follow its reference links for the settings being changed.
Use the [StackPack reference](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/custom-integrations/stackpacks.html) for project structure, versioning, and CLI commands.
Use [contrib-stackpacks](https://github.com/StackVista/contrib-stackpacks/tree/main/stackpacks) for working examples:

- `open-telemetry`: component and relation mappings, presentations, and monitors.
- `otel-k8s-crd`: topology from Kubernetes custom-resource events.
- `suse-observability`: presentations and monitors for existing platform components.

Read only the relevant example files. Prefer current schemas over copying an entire integration.

## Telemetry design and discovery

Use OpenTelemetry as the base data model. Prefer collection through the OpenTelemetry Collector.
Find existing receivers or instrumentation for the target technology before proposing custom collection.
Identify the required signals, resource identity, and Collector configuration.
If collection is missing, provide the required configuration and ingestion instructions with the StackPack.

Use the relevant [OpenTelemetry semantic conventions](https://github.com/open-telemetry/semantic-conventions) for resource attributes, metrics, spans, and logs.
Read the technology-specific conventions and check their stability and support in the selected receiver or instrumentation.
Prefer standard names and units. Document extensions only where no suitable convention exists.
Compare conventions with observed telemetry, including metric-name and label transformations during ingestion.
If they differ, make collection normalization or mapping compatibility explicit.

Use the official [MCP connection guide](https://documentation.suse.com/cloudnative/suse-observability/latest/en/use/ai-assistant/mcp-server.html) to connect your assistant to the SUSE Observability MCP server.
When it is available, use its query tools to inspect data for the target technology.
Read the relevant MCP domain guides first and establish that MCP and CLI target the intended instance.
Scope queries to relevant resources and time ranges.
Inspect available metrics, labels, topology, traces, and logs to establish identifiers and usable signals.
Use those results to author mappings and queries, then check their behavior after deployment.
If MCP is unavailable, use CLI queries or supplied telemetry samples.
If data is absent, document the collection gap and leave runtime behavior unverified.

## CLI and authentication

Check whether `sts` is available.
If it is missing, ask whether the user will install it.
Direct them to the **CLI** main menu item in the SUSE Observability UI for installation.
Wait for installation before running CLI commands. Continue independent design work while waiting.
Use the latest stable Go `sts` CLI from [stackstate-cli releases](https://github.com/StackVista/stackstate-cli/releases/latest).
Use the official [CLI installation guide](https://documentation.suse.com/cloudnative/suse-observability/latest/en/setup/cli/cli-sts.html) for installation details.
Resolve the release on each task. Do not embed a fixed CLI version in this skill.
Compare the installed version with that release and ask the user to upgrade when necessary.
Use the matching operating-system and architecture artifact, and check its published checksum before execution.

The user authenticates the CLI to the correct SUSE Observability instance.
Do not log in, change contexts, request tokens, or read credential files.
If authentication fails, ask the user to authenticate and continue local work.
Before instance operations, establish the intended target from the user's request without displaying credentials.
Inspect command help from the installed CLI when flags or server compatibility are unclear.

## Create and customize

1. Establish the target product or technology, integration name, and desired topology or monitoring behavior.
   Use telemetry discovery to select signals and resource attributes. Ask for missing evidence rather than inventing it.
2. Scaffold into a new directory:

   ```bash
   sts stackpack scaffold --name my-stackpack --display-name "My StackPack"
   ```

3. Read the generated `README.md` and `stackpack.yaml`. Preserve the required settings and resource paths.
4. Replace demo content with the requested integration. Keep identifiers unique within the pack namespace.
   Align component identifiers, relation endpoints, monitor bindings, and presentation selectors.
   Use actual telemetry labels in queries and meaningful production thresholds in monitors.
   Add only the mappings, presentations, monitors, and resources the integration needs.
5. Package from the StackPack directory:

   ```bash
   sts stackpack package -d .
   ```

Packaging checks local structure. It does not prove server schema validity or successful installation.

## Validate and deploy

For authorized instance validation, use:

```bash
sts stackpack validate -d .
```

Validation sends the package to the configured server.
For CI without an instance, use the offline validator described in the StackPack reference, pinned to the target server release.
For an authorized development deployment, use:

```bash
sts stackpack test-deploy -d . --yes
```

This command uploads and installs or upgrades the pack with a test version suffix.
It can upgrade an existing installation. It is not a dry run.
Repeat after edits within the authorized scope.
Check the resulting topology, relations, metric queries, monitor bindings, and presentations against matching telemetry.
Use MCP `topology_query` and `component_get` to check component identities, properties, and relation endpoints.
Use `metrics_query_instant` or `metrics_query_range` to run the actual PromQL queries, and `metrics_bound` to check component bindings.
Use `monitors_overview`, `monitor_config`, and `monitor_violations` to check installed monitors and their results.
Verify rendered presentations in the SUSE Observability UI; MCP queries do not establish their visual correctness.

For delivery, package again and report the generated `.sts` path.
If upload is requested, run `sts stackpack upload --file <generated-file.sts>`.
Upload makes a version available. Installation is a separate operation.
Report which local and instance checks passed, and which remain untested.
