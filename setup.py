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

REPO_ROOT = Path(__file__).resolve().parent


def read_minimum_python_version() -> tuple[int, int]:
    version_text = (REPO_ROOT / ".python-version").read_text().strip()
    version = version_text.split(".", 1)
    if len(version) != 2:
        raise OSError(
            "Expected .python-version to contain a major.minor version, "
            f"got: {version_text!r}",
        )

    major, minor = (int(part) for part in version)
    if major != 3:
        raise OSError(
            "Expected .python-version to declare a Python 3 interpreter, "
            f"got: {version_text!r}",
        )

    return major, minor


MINIMUM_PYTHON_MAJOR, MINIMUM_PYTHON_MINOR = read_minimum_python_version()
MINIMUM_PY_LIMITED_API = f"0x{MINIMUM_PYTHON_MAJOR:02X}{MINIMUM_PYTHON_MINOR:02X}0000"
MINIMUM_WHEEL_PY_LIMITED_API = f"cp3{MINIMUM_PYTHON_MINOR}"


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
        stable_abi_lib = libs_dir / "python3.lib"
        if not stable_abi_lib.exists():
            raise OSError(f"Stable ABI import library not found: {stable_abi_lib}")
        return f"-L{libs_dir} -lpython3"
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


if platform.python_implementation() == "CPython":
    try:
        from setuptools.command.bdist_wheel import bdist_wheel as _bdist_wheel
    except ImportError:
        cmdclass = {}
    else:

        class bdist_wheel(_bdist_wheel):
            def finalize_options(self) -> None:
                self.py_limited_api = MINIMUM_WHEEL_PY_LIMITED_API
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
            define_macros=[("Py_LIMITED_API", MINIMUM_PY_LIMITED_API)],
        ),
    ],
    cmdclass=cmdclass,
)
