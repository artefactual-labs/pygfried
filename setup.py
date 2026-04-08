import os
import platform
import shlex
import shutil
import subprocess
import sys
from pathlib import Path

from setuptools import Extension
from setuptools import setup
from setuptools.command.build_ext import build_ext as _build_ext


def format_cgo_cflags(
    include_dirs: list[str],
    macros: list[tuple[str, str | None]],
) -> str:
    parts = [f"-I{include_dir}" for include_dir in include_dirs]
    for macro_name, macro_value in macros:
        if macro_value is None:
            parts.append(f"-D{macro_name}")
        else:
            parts.append(f"-D{macro_name}={macro_value}")
    return " ".join(parts)


def platform_ldflags() -> str:
    if sys.platform == "darwin":
        return "-Wl,-undefined,dynamic_lookup"
    if sys.platform == "win32":
        libs_dir = Path(sys.base_prefix) / "libs"
        version = f"{sys.version_info[0]}{sys.version_info[1]}"
        return f"-L{libs_dir} -lpython{version}"
    return "-Wl,--unresolved-symbols=ignore-all"


class build_ext(_build_ext):
    def build_extension(self, ext: Extension) -> None:
        go_sources = [source for source in ext.sources if source.endswith(".go")]
        if not go_sources:
            super().build_extension(ext)
            return

        if len(ext.sources) != 1 or len(go_sources) != 1:
            raise OSError(
                f"Error building extension `{ext.name}`: sources must be a "
                f"single file in the `main` package.\nReceived: {ext.sources!r}",
            )

        if not shutil.which("go"):
            raise OSError("Go compiler not found on PATH")

        main_file = Path(go_sources[0])
        if not main_file.exists():
            raise OSError(
                f"Error building extension `{ext.name}`: {main_file} does not exist",
            )

        ext_path = Path(self.get_ext_fullpath(ext.name))
        ext_path.parent.mkdir(parents=True, exist_ok=True)

        include_dirs = list(self.compiler.include_dirs or [])
        include_dirs.extend(ext.include_dirs or [])
        macros = list(ext.define_macros or [])

        env = os.environ.copy()
        env["CGO_CFLAGS"] = format_cgo_cflags(include_dirs, macros)
        env["CGO_LDFLAGS"] = platform_ldflags()

        cmd = [
            "go",
            "build",
            "-buildmode=c-shared",
            "-o",
            str(ext_path.resolve()),
            "-ldflags=-s -w",
        ]
        self.announce(
            "$ "
            + " ".join(
                [
                    f"CGO_CFLAGS={shlex.quote(env['CGO_CFLAGS'])}",
                    f"CGO_LDFLAGS={shlex.quote(env['CGO_LDFLAGS'])}",
                    shlex.join(cmd),
                ]
            ),
            level=2,
        )
        subprocess.run(cmd, cwd=main_file.parent, env=env, check=True)


if sys.platform != "win32" and platform.python_implementation() == "CPython":
    try:
        from setuptools.command.bdist_wheel import bdist_wheel as _bdist_wheel
    except ImportError:
        cmdclass = {}
    else:

        class bdist_wheel(_bdist_wheel):
            def finalize_options(self) -> None:
                self.py_limited_api = f"cp3{sys.version_info[1]}"
                super().finalize_options()

        cmdclass = {"bdist_wheel": bdist_wheel}
else:
    cmdclass = {}

cmdclass["build_ext"] = build_ext

setup(
    ext_modules=[
        Extension(
            "pygfried",
            ["pylib/main.go"],
            py_limited_api=True,
            define_macros=[("Py_LIMITED_API", None)],
        ),
    ],
    cmdclass=cmdclass,
)
