import sys
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
from pathlib import Path
from shutil import copyfile

import pytest

import pygfried

conftest_path = Path(__file__).parent / "conftest.py"
archivematica_signature = (
    Path(__file__).parent.parent / "internal" / "signatures" / "archivematica.sig"
)


def test_version(siegfried_version):
    assert pygfried.version() == siegfried_version


def test_identify():
    result = pygfried.identify(str(conftest_path))
    assert result == "fmt/938"


def test_identify_detailed(siegfried_version):
    result = pygfried.identify(str(conftest_path), detailed=True)

    assert result["siegfried"] == siegfried_version
    parse_iso_date(result["scandate"])
    assert result["signature"] == "default.sig"
    parse_iso_date(result["created"])
    assert len(result["identifiers"]) > 0
    assert len(result["files"]) == 1

    finfo = dict(result["files"][0])
    parse_iso_date(str(finfo.pop("modified")))
    assert finfo == {
        "filename": str(conftest_path),
        "filesize": conftest_path.stat().st_size,
        "errors": "",
        "matches": [
            {
                "ns": "pronom",
                "id": "fmt/938",
                "format": "Python Source Code File",
                "version": "",
                "mime": "",
                "class": "Text (Structured)",
                "basis": "extension match py",
                "warning": "match on extension only",
            }
        ],
    }


def test_identify_detailed_escapes_error_strings():
    path = "C:\\Users\\j472\\Documents\\missing"

    result = pygfried.identify(path, detailed=True)

    assert len(result["files"]) == 1
    assert result["files"][0]["filename"] == path
    assert path in result["files"][0]["errors"]


def test_identify_many(siegfried_version):
    result = pygfried.identify_many(
        [str(conftest_path), str(conftest_path.parent / "__init__.py")],
        workers=2,
    )

    assert result["siegfried"] == siegfried_version
    assert [file["filename"] for file in result["files"]] == [
        str(conftest_path),
        str(conftest_path.parent / "__init__.py"),
    ]


def test_identify_many_empty():
    result = pygfried.identify_many([])

    assert result["files"] == []


def test_identify_many_rejects_bad_workers():
    with pytest.raises(ValueError, match="workers must be between 1 and 1024"):
        pygfried.identify_many([], workers=0)


def test_identify_many_rejects_non_string_paths():
    with pytest.raises(pygfried.GoError, match="paths must contain only strings"):
        pygfried.identify_many([conftest_path])


def test_identify_many_rejects_null_bytes():
    with pytest.raises(pygfried.GoError, match="paths must not contain null bytes"):
        pygfried.identify_many([f"{conftest_path}\0missing"])


def test_identify_many_rejects_string_path_iterable():
    with pytest.raises(TypeError, match="not a string"):
        pygfried.identify_many(str(conftest_path))


def test_identify_dir_recursive(tmp_path):
    root_file = tmp_path / "a.py"
    root_file.write_text("print('hello')\n")
    nested = tmp_path / "nested"
    nested.mkdir()
    nested_file = nested / "b.py"
    nested_file.write_text("print('hello')\n")

    result = pygfried.identify_dir(str(tmp_path), workers=2)

    assert [file["filename"] for file in result["files"]] == [
        str(root_file),
        str(nested_file),
    ]


def test_identify_dir_non_recursive(tmp_path):
    root_file = tmp_path / "a.py"
    root_file.write_text("print('hello')\n")
    nested = tmp_path / "nested"
    nested.mkdir()
    (nested / "b.py").write_text("print('hello')\n")

    result = pygfried.identify_dir(str(tmp_path), recursive=False)

    assert [file["filename"] for file in result["files"]] == [str(root_file)]


def test_identify_dir_skips_symlinks(tmp_path):
    target = tmp_path / "target.py"
    target.write_text("print('hello')\n")
    link = tmp_path / "link.py"
    try:
        link.symlink_to(target)
    except OSError as err:
        pytest.skip(f"symlink unavailable: {err}")

    result = pygfried.identify_dir(str(tmp_path))

    assert [file["filename"] for file in result["files"]] == [str(target)]


def test_identify_dir_follows_file_symlinks(tmp_path):
    target = tmp_path / "target.py"
    target.write_text("print('hello')\n")
    link = tmp_path / "link.py"
    try:
        link.symlink_to(target)
    except OSError as err:
        pytest.skip(f"symlink unavailable: {err}")

    result = pygfried.identify_dir(str(tmp_path), follow_symlinks=True)

    assert [file["filename"] for file in result["files"]] == [
        str(link),
        str(target),
    ]


def test_identify_dir_follows_directory_symlinks(tmp_path):
    root = tmp_path / "root"
    root.mkdir()
    target_dir = tmp_path / "target"
    target_dir.mkdir()
    target = target_dir / "target.py"
    target.write_text("print('hello')\n")
    link = root / "linked"
    try:
        link.symlink_to(target_dir, target_is_directory=True)
    except OSError as err:
        pytest.skip(f"symlink unavailable: {err}")

    result = pygfried.identify_dir(str(root), follow_symlinks=True)

    assert [file["filename"] for file in result["files"]] == [
        str(link / "target.py"),
    ]


def test_identify_dir_skips_symlink_cycles(tmp_path):
    target = tmp_path / "target.py"
    target.write_text("print('hello')\n")
    link = tmp_path / "loop"
    try:
        link.symlink_to(tmp_path, target_is_directory=True)
    except OSError as err:
        pytest.skip(f"symlink unavailable: {err}")

    result = pygfried.identify_dir(str(tmp_path), follow_symlinks=True)

    assert [file["filename"] for file in result["files"]] == [str(target)]


def test_identify_dir_skips_repeated_symlink_targets(tmp_path):
    root = tmp_path / "root"
    root.mkdir()
    target_dir = tmp_path / "target"
    target_dir.mkdir()
    target = target_dir / "target.py"
    target.write_text("print('hello')\n")
    first = root / "a"
    second = root / "b"
    try:
        first.symlink_to(target_dir, target_is_directory=True)
        second.symlink_to(target_dir, target_is_directory=True)
    except OSError as err:
        pytest.skip(f"symlink unavailable: {err}")

    result = pygfried.identify_dir(str(root), follow_symlinks=True)

    assert [file["filename"] for file in result["files"]] == [
        str(first / "target.py"),
    ]


def test_identify_dir_requires_directory():
    with pytest.raises(pygfried.GoError, match="is not a directory"):
        pygfried.identify_dir(str(conftest_path))


def test_scanner_defaults_to_default_profile():
    scanner = pygfried.Scanner()

    assert scanner.profile == "default"
    assert scanner.signature == "default.sig"
    assert scanner.identify(str(conftest_path)) == pygfried.identify(str(conftest_path))


def test_scanner_accepts_explicit_default_profile():
    scanner = pygfried.Scanner(profile="default")

    assert scanner.profile == "default"
    assert scanner.signature == "default.sig"


def test_scanner_profile_properties_are_read_only():
    scanner = pygfried.Scanner()

    with pytest.raises(AttributeError):
        scanner.profile = "archivematica"
    with pytest.raises(AttributeError):
        scanner.signature = "custom.sig"


def test_scanner_rejects_positional_profile():
    with pytest.raises(TypeError, match="takes no positional arguments"):
        pygfried.Scanner("archivematica")


def test_scanner_rejects_conflicting_sources():
    with pytest.raises(
        ValueError,
        match="profile and signature are mutually exclusive",
    ):
        pygfried.Scanner(
            profile="default",
            signature=str(archivematica_signature),
        )


def test_scanner_rejects_unknown_profile():
    with pytest.raises(ValueError, match='unknown profile "unknown"'):
        pygfried.Scanner(profile="unknown")


def test_scanner_rejects_empty_signature():
    with pytest.raises(ValueError, match="signature must not be empty"):
        pygfried.Scanner(signature="")


def test_archivematica_profile_identifies_extended_signatures(tmp_path):
    ad1 = tmp_path / "sample.ad1"
    ad1.write_bytes(b"ADSEGMENTEDFILE")
    raw = tmp_path / "sample.dd"
    raw.write_bytes(b"small raw image")
    encrypted = tmp_path / "encrypted.ad1"
    encrypted.write_bytes(b"ADCRYPT")
    aff = tmp_path / "sample.aff"
    aff.write_bytes(b"AFF10\r\n\0")

    default = pygfried.Scanner()
    archivematica = pygfried.Scanner(profile="archivematica")

    assert default.identify(str(ad1)) == "fmt/842"
    assert archivematica.identify(str(ad1)) == "archivematica-fmt/2"
    assert default.identify(str(raw)) == "x-fmt/111"
    assert archivematica.identify(str(raw)) == "archivematica-fmt/4"
    assert default.identify(str(encrypted)) == "fmt/843"
    assert archivematica.identify(str(encrypted)) == "archivematica-fmt/3"
    assert default.identify(str(aff)) == "fmt/844"
    assert archivematica.identify(str(aff)) == "archivematica-fmt/5"
    assert archivematica.profile == "archivematica"
    assert archivematica.signature == "archivematica.sig"

    detailed = archivematica.identify(str(ad1), detailed=True)
    assert detailed["signature"] == "archivematica.sig"
    assert detailed["files"][0]["matches"][0]["id"] == "archivematica-fmt/2"


def test_external_signature_is_loaded_eagerly(tmp_path):
    signature = tmp_path / "custom.sig"
    copyfile(archivematica_signature, signature)
    scanner = pygfried.Scanner(signature=str(signature))
    signature.unlink()

    ad1 = tmp_path / "sample.ad1"
    ad1.write_bytes(b"ADSEGMENTEDFILE")

    assert scanner.profile is None
    assert scanner.signature == "custom.sig"
    assert scanner.identify(str(ad1)) == "archivematica-fmt/2"
    assert scanner.identify(str(ad1), detailed=True)["signature"] == "custom.sig"


def test_external_signature_escapes_name_in_detailed_results(tmp_path):
    signature = tmp_path / 'custom"\n.sig'
    try:
        copyfile(archivematica_signature, signature)
    except OSError as err:
        pytest.skip(f"signature filename unavailable: {err}")

    scanner = pygfried.Scanner(signature=str(signature))
    result = scanner.identify(str(conftest_path), detailed=True)

    assert scanner.signature == signature.name
    assert result["signature"] == signature.name


def test_external_signature_rejects_missing_file(tmp_path):
    with pytest.raises(pygfried.GoError, match="error opening signature file"):
        pygfried.Scanner(signature=str(tmp_path / "missing.sig"))


def test_external_signature_rejects_invalid_file(tmp_path):
    signature = tmp_path / "invalid.sig"
    signature.write_bytes(b"not a signature")

    with pytest.raises(pygfried.GoError, match="not a siegfried signature file"):
        pygfried.Scanner(signature=str(signature))


def test_scanner_identify_many():
    scanner = pygfried.Scanner(profile="archivematica")

    result = scanner.identify_many(
        [str(conftest_path), str(conftest_path.parent / "__init__.py")],
        workers=2,
    )

    assert result["signature"] == "archivematica.sig"
    assert [file["filename"] for file in result["files"]] == [
        str(conftest_path),
        str(conftest_path.parent / "__init__.py"),
    ]


def test_scanner_identify_dir(tmp_path):
    scanner = pygfried.Scanner(profile="archivematica")
    root_file = tmp_path / "sample.py"
    root_file.write_text("print('hello')\n")

    result = scanner.identify_dir(str(tmp_path), workers=2)

    assert result["signature"] == "archivematica.sig"
    assert [file["filename"] for file in result["files"]] == [str(root_file)]


def test_scanner_can_be_shared_between_threads(tmp_path):
    scanner = pygfried.Scanner(profile="archivematica")
    ad1 = tmp_path / "sample.ad1"
    ad1.write_bytes(b"ADSEGMENTEDFILE")

    with ThreadPoolExecutor(max_workers=8) as executor:
        results = list(
            executor.map(
                scanner.identify,
                [str(ad1)] * 32,
            )
        )

    assert results == ["archivematica-fmt/2"] * 32


def test_scanner_releases_type_reference():
    scanner = pygfried.Scanner()
    del scanner
    initial_refcount = sys.getrefcount(pygfried.Scanner)

    for _ in range(10):
        scanner = pygfried.Scanner()
        del scanner

    final_refcount = sys.getrefcount(pygfried.Scanner)
    assert final_refcount == initial_refcount


def parse_iso_date(date_string):
    """Parse ISO date string with compatibility for Python 3.9-3.11"""
    try:
        return datetime.fromisoformat(date_string)
    except ValueError:
        if date_string.endswith("Z"):
            date_string = date_string[:-1] + "+00:00"
        try:
            return datetime.fromisoformat(date_string)
        except ValueError:
            if "." in date_string:
                parts = date_string.split(".")
                if len(parts) == 2:
                    microsecs = parts[1]
                    if "+" in microsecs:
                        microsecs, tz = microsecs.split("+", 1)
                        microsecs = microsecs[:6].ljust(6, "0")
                        date_string = f"{parts[0]}.{microsecs}+{tz}"
                    elif microsecs.endswith("+00:00"):
                        microsecs = microsecs[:-6]
                        microsecs = microsecs[:6].ljust(6, "0")
                        date_string = f"{parts[0]}.{microsecs}+00:00"
                    else:
                        microsecs = microsecs[:6].ljust(6, "0")
                        date_string = f"{parts[0]}.{microsecs}"
            return datetime.fromisoformat(date_string)
