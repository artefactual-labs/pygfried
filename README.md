[![PyPI version](https://badge.fury.io/py/pygfried.svg)](https://badge.fury.io/py/pygfried)

# Pygfried

pygfried is a CPython extension that brings [siegfried] - a powerful,
signature-based file format identification tool written in Go - into the Python
ecosystem.

![Identifying a file with pygfried](./example.png)

Instead of reimplementing siegfried's logic, pygfried embeds the original Go
code directly, making core siegfried functionality available to Python users
without any changes to the underlying detection engine.

No animals were harmed in the making of this extension.

## Usage

```
$ pip install pygfried
$ python -q
>>> import pygfried
>>> pygfried.version()
'1.11.8'
>>> pygfried.identify("example.png")
'fmt/13'
>>> pygfried.identify("example.png", detailed=True)
{'siegfried': '1.11.8', 'scandate': '2026-09-23T18:18:29+02:00', 'signature': 'default.sig', 'created': '2026-09-15T19:45:35+10:00', 'identifiers': [{'name': 'pronom', 'details': 'DROID_SignatureFile_V125.xml; container-signature-20260119.xml'}], 'files': [{'filename': 'example.png', 'filesize': 676214, 'modified': '2025-06-13T11:53:28+02:00', 'errors': '', 'matches': [{'ns': 'pronom', 'id': 'fmt/13', 'format': 'Portable Network Graphics', 'version': '1.2', 'mime': 'image/png', 'class': 'Image (Raster)', 'basis': 'extension match png; byte match at [[0 16] [2962 4] [676202 12]]', 'warning': ''}]}]}
>>> pygfried.identify_many(["example.png", "README.md"], workers=2)
{'siegfried': '1.11.8', ...}
>>> pygfried.identify_dir("samples", recursive=True, workers=2)
{'siegfried': '1.11.8', ...}
```

### Custom scanners

The module-level functions use Siegfried's embedded `default.sig`. To select
another complete Siegfried signature database, create a `Scanner` with its
path:

```python
import pygfried

scanner = pygfried.Scanner(signature="/usr/share/siegfried/custom.sig")
scanner.identify("example.png")
scanner.identify_many(["example.png", "README.md"], workers=2)
scanner.identify_dir("samples", recursive=True, workers=2)
```

External databases are validated and loaded into memory when the scanner is
constructed, so changing or deleting the file afterward does not affect that
scanner.

As a convenience, a scanner can instead select a bundled signature profile by
name:

```python
scanner = pygfried.Scanner(profile="archivematica")
scanner.identify("disk-image.ad1")
# 'archivematica-fmt/2'
```

Available bundled profiles are:

- `default`, which is also selected by `Scanner()` and matches the behavior of
  the module-level functions.
- `archivematica`, which contains the complete PRONOM database plus
  Archivematica's extended identifiers for AD1, encrypted AD1, raw disk images,
  and AFF.

The Archivematica database is copied from the exact Siegfried module version
used to compile pygfried. It is embedded in the extension, so it does not need
to be installed separately.

`profile` and `signature` are keyword-only and mutually exclusive. The
read-only `scanner.profile` property reports the bundled profile, or `None`
for an external database. The read-only `scanner.signature` property reports
the database's filename without its directory path.

### Batch and directory scans

Use `identify_many` when you already have a list of paths, or `identify_dir`
when you want pygfried to scan a directory for you. Both functions return the
same detailed result shape as `identify(..., detailed=True)`.

```
>>> from pathlib import Path
>>> paths = [str(path) for path in Path("samples").rglob("*.png")]
>>> pygfried.identify_many(paths, workers=4)
{'siegfried': '1.11.8', ...}
>>> pygfried.identify_dir("samples", recursive=True, workers=4)
{'siegfried': '1.11.8', ...}
```

Batch and directory scans avoid repeated Python-to-Go calls. The `workers`
argument controls Go-side concurrency and defaults to `1`. Higher values allow
multiple files to be identified in parallel. To tune performance, start with
the number of CPU cores available to your process, then measure with your own
files.

By default `identify_dir` skips symlinks. Use `follow_symlinks=True` to
identify file symlinks and descend symlinked directories; directory cycles are
skipped, and repeated links to the same directory are scanned once.

### Concurrency

Both the module-level functions and `Scanner` methods are safe to call from
multiple threads. The module-level API shares one immutable default scanner;
each custom scanner owns an independent, immutable Siegfried engine.

The extension holds Python's GIL during identification, so Python threads do
not run scans in parallel. Use `workers` for parallel file identification
within a batch.

Each `identify_many` or `identify_dir` call controls its own Go-side concurrency
with `workers`, whether it is called on the module or on a scanner. There is no
process-wide worker limit beyond the per-call range of 1 to 1024.

For process-based parallelism, use the `spawn` start method explicitly
(`multiprocessing.get_context("spawn")`). Import pygfried and create any custom
scanners inside each worker process. Do not pass scanner instances between
processes. Other start methods have not been verified.

## Limitations

### Go libraries can clash

This project uses Go's `-buildmode=c-shared` to provide its Python extension.
Loading multiple Go-based shared libraries in the same process has [known issues]
and may result in panics or crashes due to conflicts between separate Go runtimes.

This limitation should only affect you if you're using pygfried together with
another Python library that also uses a Go extension (built with the same
c-shared mechanism) in the same process. If you're just using pygfried on its
own, you don't need to worry - everything should work as expected.

## Credits

pygfried is powered by the original [siegfried] project, which is distributed
under the Apache License, Version 2.0. All core file format identification logic
and signatures are provided by siegfried. We gratefully acknowledge the work of
the siegfried project and its contributors.

[siegfried]: https://www.itforarchivists.com/siegfried
[known issues]: https://github.com/golang/go/issues/65050
