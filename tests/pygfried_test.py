import pytest

import pygfried


def test_version():
    assert pygfried.version() == "1.10.0"
