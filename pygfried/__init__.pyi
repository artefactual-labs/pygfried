from collections.abc import Iterable
from typing import Literal
from typing import TypedDict
from typing import overload

class GoError(Exception):
    """Exception raised when Go code encounters an error."""

SimpleIdentifyResult = Literal["UNKNOWN"] | str | None

class Match(TypedDict):
    ns: str
    id: str
    format: str
    version: str
    mime: str
    basis: str
    warning: str

Match.__annotations__["class"] = str

class File(TypedDict):
    filename: str
    filesize: int
    modified: str
    errors: str
    matches: list[Match]

class Identifier(TypedDict):
    name: str
    details: str

class DetailedIdentifyResult(TypedDict):
    siegfried: str
    scandate: str
    signature: str
    created: str
    identifiers: list[Identifier]
    files: list[File]

@overload
def identify(path: str, detailed: Literal[True]) -> DetailedIdentifyResult: ...
@overload
def identify(path: str, detailed: Literal[False] = False) -> SimpleIdentifyResult: ...
def identify_many(
    paths: Iterable[str],
    *,
    workers: int = 1,
) -> DetailedIdentifyResult: ...
def identify_dir(
    path: str,
    *,
    recursive: bool = True,
    workers: int = 1,
    follow_symlinks: bool = False,
) -> DetailedIdentifyResult: ...
def version() -> str: ...
