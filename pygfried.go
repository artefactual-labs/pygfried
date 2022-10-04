package pygfried

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/static"
)

var sf *siegfried.Siegfried

func load() *siegfried.Siegfried {
	if sf != nil {
		return sf
	}

	sf = static.New()

	return sf
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
	sf := load()

	ids, err := identify(sf, path)
	if err != nil {
		return nil, err
	}

	r := buildResult(path, ids, err)
	return r, err
}

func IdentifyAll(paths []string) ([]*Result, error) {
	sf := load()

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
