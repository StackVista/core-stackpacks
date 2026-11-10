# Core StackPacks

The SUSE Observability team maintains these StackPacks:

- `open-telemetry`
- `otel-k8s-crd`
- [`service-availability`](stackpacks/service-availability/README.md) — HTTP and HTTPS endpoint availability, service grouping, and optional TLS certificate expiration checks.
- `suse-observability`

The product includes the core bundle by default.

## Repository ownership

| Repository | Content | Maintainer | Delivery |
| --- | --- | --- | --- |
| [core-stackpacks](https://github.com/StackVista/core-stackpacks) | Product integrations | SUSE Observability team | Included with the product |
| [suse-stackpacks](https://github.com/StackVista/suse-stackpacks) | SUSE product integrations | Respective product teams | Product distribution |
| [contrib-stackpacks](https://github.com/StackVista/contrib-stackpacks) | Community integrations for external technologies | Named community maintainers | Customers add individual packages |

The existing `stackpacks` repository and SDK keep their current roles.

## Build and validate

The CI workflow validates every StackPack with the platform validator.
It packages the content with the `sts` CLI.
Changes to an existing StackPack require a higher package version.
The migration keeps the existing package names, versions, and content.

For a local package build, run:

```sh
cd stackpacks/open-telemetry
STS_EXPERIMENTAL_STACKPACKS=1 sts stackpack package
```

Use the validator image declared in `.github/workflows/stackpacks-ci.yml`:

```sh
VALIDATOR_IMAGE='<validator image from the workflow>' \
  bash scripts/ci/validate_stackpack.sh stackpacks/open-telemetry
```

Use the [Create a StackPack skill](.agents/skills/create-stackpack/SKILL.md)
to create or extend an integration.

## Image delivery

CI builds a bundle for `linux/amd64` and `linux/arm64`.
The image contains the packaged files in `/stackpacks`.
It uses SUSE BCI and a non-root user.
The Helm chart supplies the script that copies these files into the platform.

The image builds compare the packaged files against the image contents.
Trivy secret scanning and vulnerability scanning gate publication.
The shared image pipeline publishes those tested images and signs their combined manifest.
Fork pull requests run validation and image checks without publication credentials.
Same-repository pull requests and `main` pushes can publish a bundle.

The image reference uses this form:

```text
quay.io/stackstate/core-stackpacks:<timestamp>-<branch>-<commit>
```

The timestamp and commit identify the workflow checkout.
The tag on a pull request can identify GitHub's test merge commit.
Retried publication jobs append `-retry-<attempt>` to create a new immutable tag.
The content versions remain in each `stackpack.yaml`.

After a successful `main` publication, CI requests the Helm chart version update.
Pull requests and failed publications do not trigger this update.

## Publication credentials

CI requires `QUAY_USER` and `QUAY_PASSWORD` for the public Quay repository.
The chart update also requires GitHub App credentials.
Infrastructure configuration manages these credentials.
The publication workflow makes sure that the image permits anonymous pulls.

Before the chart adopts a new bundle, make sure that this command succeeds:

```sh
image='quay.io/stackstate/core-stackpacks:<published-tag>'
temporary_docker_configuration=$(mktemp -d)
DOCKER_CONFIG="${temporary_docker_configuration}" docker buildx imagetools inspect "${image}"
rm -rf "${temporary_docker_configuration}"
```

The chart's existing bundle stays in place until the replacement image is available.
