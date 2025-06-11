from datetime import datetime
from pathlib import Path

import pygfried

conftest_path = Path(__file__).parent / "conftest.py"


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

    finfo = result["files"][0]
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
