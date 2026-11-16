The mapping regression uses the supplied collector checkout's CEL evaluator and
mapper. It copies the topology connector into a temporary directory and reads
its dependencies without changing the checkout.

```sh
python3 tests/otel-k8s-crd/run.py \
  --collector-source /path/to/sts-opentelemetry-collector \
  --go /path/to/the/collectors/go
```

It covers current and stale Ready conditions, missing feedback, pending and
blocked collection, retained configuration/status, generic CR selection and
identity-only deletion. Run the platform validator separately to check the
presentation schema and expressions.

These checks do not establish installation, backend persistence or rendered UI
behavior. They require a collector revision compatible with the StackPack and
run in the existing CI validation job for `otel-k8s-crd`, before packaging.
The workflow pins the collector engine revision and derives its Go toolchain
from that revision's module file.

The presentation regression uses the platform image's actual STQL matching,
highlight evaluation, overview composition and CRD query specialization. Use
the validator image from the repository workflow, pinned by digest, and JDK 21:

```sh
python3 tests/otel-k8s-crd/presentation.py \
  --validator-image 'quay.io/stackstate/stackstate-server@sha256:<digest>'
```

It extracts the image's libraries into a temporary directory and leaves the
platform checkout untouched. It verifies the collection states, missing observed
generation, name/cluster/age columns, and absence of namespace fields. These local
API checks do not prove persistence or browser rendering.
CI runs the same presentation check with JDK 21 from the GitHub-hosted runner and
the existing validator image pinned by digest.
