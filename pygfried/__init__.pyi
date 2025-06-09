from typing import Union, Literal, Any

class GoError: ...

SimpleIdentifyResult = Union[Literal["UNKNOWN"], str, None]
DetailedIdentifyResult = Any

def identify(
    path: str, detailed: bool = False
) -> Union[SimpleIdentifyResult, DetailedIdentifyResult]: ...
def version() -> str: ...
