from typing import Any

import shutil
import sys

from hatchling.builders.hooks.plugin.interface import BuildHookInterface


class CustomBuildHook(BuildHookInterface):
    def __init__(self, *args: Any, **kwargs: Any) -> None:
        super().__init__(*args, **kwargs)

    def initialize(self, version, build_data):
        self.app.display_info(f"Initializing build for version {version!s}…")
        print(build_data)
        # Check for Go compiler
        if not shutil.which("go"):
            self.app.display_info(
                "Go compiler not found. Please install Go from https://golang.org/dl/."
            )
            sys.exit(1)
        self.build_ext()

    def build_ext(self):
        self.app.display_info("Building Go extension…")
