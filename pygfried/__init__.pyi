from typing import Literal
from typing import Union

class GoError: ...
def identify(path: str) -> Union[Literal["UNKNOWN"], str, None]: ...
def version() -> str: ...
