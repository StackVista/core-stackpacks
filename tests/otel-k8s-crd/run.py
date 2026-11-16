#!/usr/bin/env python3
"""Evaluate the StackPack mappings with the supplied collector's CEL engine."""

import argparse
import os
from pathlib import Path
import shutil
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--collector-source", required=True, type=Path)
    parser.add_argument("--go", default="go")
    args = parser.parse_args()
    collector = args.collector_source.resolve()
    tests = Path(__file__).resolve().parent
    pack = tests.parents[1] / "stackpacks" / "otel-k8s-crd"
    source = collector / "connector" / "topologyconnector"
    if not (source / "go.mod").is_file():
        parser.error("--collector-source must contain connector/topologyconnector")
    with tempfile.TemporaryDirectory(prefix="resource-collection-mapping-") as temporary:
        root = Path(temporary)
        module = root / "connector" / "topologyconnector"
        shutil.copytree(source, module)
        for dependency in (
            "exporter/stskafkaexporter",
            "extension/settingsproviderextension",
        ):
            destination = root / dependency
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.symlink_to(collector / dependency, target_is_directory=True)
        shutil.copyfile(
            tests / "resource_collections_test.go",
            module / "internal" / "stackpack_resource_collections_test.go",
        )
        env = os.environ.copy()
        env.update(GOWORK="off", RESOURCE_COLLECTION_STACKPACK=str(pack))
        subprocess.run(
            [
                args.go, "test", "-mod=readonly", "-count=1", "-v",
                "./internal", "-run", "^TestStackPackResourceCollections$",
            ],
            cwd=module,
            env=env,
            check=True,
        )


if __name__ == "__main__":
    main()
