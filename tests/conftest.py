import json
import os
import subprocess
from pathlib import Path

import pytest


@pytest.fixture(scope="session")
def siegfried_version():
    if version := os.environ.get("PYGFRIED_TEST_SIEGFRIED_VERSION"):
        return version.removeprefix("v")

    modname = "github.com/richardlehane/siegfried"
    root_dir = Path(__file__).parent
    result = subprocess.run(
        ["go", "list", "-m", "-json", modname],
        cwd=root_dir,
        capture_output=True,
        text=True,
        check=True,
    )
    module_info = json.loads(result.stdout)
    if module_info.get("Path") == modname:
        return module_info.get("Version").lstrip("v")  # Remove the "v" prefix.
    raise ValueError("Could not find siegfried in the go.mod file.")
