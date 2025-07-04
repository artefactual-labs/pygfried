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
'1.11.2'
>>> pygfried.identify("example.png")
'fmt/11'
>>> pygfried.identify("example.png", detailed=True)
{'siegfried': '1.11.2', 'scandate': '2025-06-10T07:16:31+02:00', 'signature': 'default.sig', 'created': '2025-03-01T15:28:08+11:00', 'identifiers': [{'name': 'pronom', 'details': 'DROID_SignatureFile_V120.xml; container-signature-20240715.xml'}], 'files': [{'filename': 'example.png', 'filesize': 237675, 'modified': '2025-06-10T07:11:26+02:00', 'errors': '', 'matches': [{'ns': 'pronom', 'id': 'fmt/11', 'format': 'Portable Network Graphics', 'version': '1.0', 'mime': 'image/png', 'class': 'Image (Raster)', 'basis': 'extension match png; byte match at [[0 16] [237663 12]]', 'warning': ''}]}]}
```

## Limitations

### Go libraries can clash

This project uses Go's `-buildmode=c-shared` to provide its Python extension.
Loading multiple Go-based shared libraries in the same process is [unsupported]
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
[unsupported]: https://github.com/golang/go/issues/65050

