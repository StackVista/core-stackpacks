#!/usr/bin/env python3
"""Check presentation composition using the platform validator image's API code."""

import argparse
import os
from pathlib import Path
import re
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--validator-image", required=True)
    parser.add_argument("--pack-directory", type=Path)
    parser.add_argument("--java", default="java")
    parser.add_argument("--javac", default="javac")
    args = parser.parse_args()
    if not re.search(r"@sha256:[a-f0-9]{64}$", args.validator_image):
        parser.error("--validator-image must be pinned by digest")
    tests = Path(__file__).resolve().parent
    pack = args.pack_directory or tests.parents[1] / "stackpacks" / "otel-k8s-crd"
    with tempfile.TemporaryDirectory(prefix="resource-collection-presentation-") as temporary:
        root = Path(temporary)
        container = subprocess.check_output(
            ["docker", "create", "--network", "none", args.validator_image],
            text=True,
        ).strip()
        try:
            subprocess.run(
                ["docker", "cp", f"{container}:/opt/docker/lib", str(root)],
                check=True,
            )
        finally:
            subprocess.run(
                ["docker", "rm", container], check=True, stdout=subprocess.DEVNULL,
            )
        classes = root / "classes"
        classes.mkdir()
        libraries = str(root / "lib" / "*")
        subprocess.run(
            [
                args.javac, "-proc:none", "-cp", libraries, "-d", str(classes),
                str(tests / "ResourceCollectionPresentationTest.java"),
            ],
            cwd=root, check=True,
        )
        logging = root / "logback.xml"
        logging.write_text('<configuration><root level="OFF"/></configuration>\n')
        subprocess.run(
            [
                args.java, f"-Dlogback.configurationFile={logging}",
                "-cp", str(classes) + os.pathsep + libraries,
                "ResourceCollectionPresentationTest", str(pack.resolve()),
            ],
            cwd=root, check=True,
        )


if __name__ == "__main__":
    main()
