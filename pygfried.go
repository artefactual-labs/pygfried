package pygfried

import (
	"bytes"
	"compress/flate"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
)

//go:embed default.sig
var signatureFile []byte

var sf *siegfried.Siegfried

func load() (*siegfried.Siegfried, error) {
	if sf != nil {
		return sf, nil
	}
	offset := len(config.Magic()) + 2
	r := bytes.NewBuffer(signatureFile[offset:])
	rc := flate.NewReader(r)
	defer rc.Close()
	s, err := siegfried.LoadReader(rc)
	if err != nil {
		return nil, err
	}
	sf = s
	return sf, err
}

func identify(sf *siegfried.Siegfried, path string) ([]core.Identification, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return sf.Identify(f, filepath.Base(path), "")
}

type Result struct {
	Path        string
	Identifiers []string
	Known       bool
	Error       string
}

func buildResult(path string, ids []core.Identification, err error) *Result {
	res := &Result{
		Path: path,
	}
	if err != nil {
		res.Error = err.Error()
	} else {
		for _, id := range ids {
			res.Identifiers = append(res.Identifiers, id.String())
			res.Known = id.Known()
		}
	}
	return res
}

func Identify(path string) (*Result, error) {
	sf, err := load()
	if err != nil {
		return nil, err
	}

	ids, err := identify(sf, path)
	if err != nil {
		return nil, err
	}

	r := buildResult(path, ids, err)
	return r, err
}

func IdentifyAll(paths []string) ([]*Result, error) {
	sf, err := load()
	if err != nil {
		return nil, err
	}

	rs := make([]*Result, len(paths))
	for idx, path := range paths {
		ids, err := identify(sf, path)
		rs[idx] = buildResult(path, ids, err)
	}

	return rs, nil
}

func Version() string {
	v := config.Version()
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}
