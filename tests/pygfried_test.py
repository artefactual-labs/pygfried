import pytest

import pygfried


def test_version():
    assert pygfried.version() == "1.9.6"
