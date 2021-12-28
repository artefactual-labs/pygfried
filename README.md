[![PyPI version](https://badge.fury.io/py/pygfried.svg)](https://badge.fury.io/py/pygfried)

# Pygfried

CPython extension of [sigfried], a signature-based file format identification
tool written in Go.

No animals were harmed in the making of this extension.

## Usage

```
$ pip install pygfried
$ python -q
>>> import pygfried
>>> pygfried.version()
'1.9.1'
>>> pygfried.identify("/bin/ls")
'fmt/690'
```

[sigfried]: https://www.itforarchivists.com/siegfried
