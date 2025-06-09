[![PyPI version](https://badge.fury.io/py/pygfried.svg)](https://badge.fury.io/py/pygfried)

# Pygfried

pygfried is a CPython extension that brings [siegfried] - a powerful,
signature-based file format identification tool written in Go - into the Python
ecosystem. Instead of reimplementing siegfried's logic, pygfried embeds the
original Go code directly, making core siegfried functionality available to
Python users without any changes to the underlying detection engine.

No animals were harmed in the making of this extension.

## Usage

```
$ pip install pygfried
$ python -q
>>> import pygfried
>>> pygfried.version()
'1.11.2'
>>> pygfried.identify("/bin/ls")
'fmt/690'
```

## Limitations

### Metadata exposure

Currently, `pygfried.identify` only returns the first match's PUID, and does not
expose additional details or warnings reported by siegfried, such as
human-readable format names or analysis issues.

See [issue #7](https://github.com/artefactual-labs/pygfried/issues/7) for more
details.

### Platform support

Right now, only Python wheels for linux/amd64 are available. If you're using
another operating system or architecture, you'll need to build pygfried from
source or wait until pre-built wheels are provided for your platform.

See [issue #1](https://github.com/artefactual-labs/pygfried/issues/1) for more
details.

### Go libraries can clash

This project uses Go's `-buildmode=c-shared` to provide its Python extension.
Loading multiple Go-based shared libraries in the same process is [unsupported]
and may result in panics or crashes due to conflicts between separate Go runtimes.

This limitation should only affect you if you're using pygfried together with
another Python library that also uses a Go extension (built with the same
c-shared mechanism) in the same process. If you're just using pygfried on its
own, you don't need to worry - everything should work as expected.

[siegfried]: https://www.itforarchivists.com/siegfried
[unsupported]: https://github.com/golang/go/issues/65050
