package pygfried

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/config"
	"github.com/richardlehane/siegfried/pkg/core"
	"github.com/richardlehane/siegfried/pkg/static"
	"github.com/richardlehane/siegfried/pkg/writer"
	"golang.org/x/sync/errgroup"
)

var (
	sf     *siegfried.Siegfried
	sfOnce sync.Once
)

const maxWorkers = 1024

func load() *siegfried.Siegfried {
	sfOnce.Do(func() {
		sf = static.New()
	})

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

type IdentifyOptions struct {
	Workers int
}

type IdentifyDirOptions struct {
	Recursive      bool
	Workers        int
	FollowSymlinks bool
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

func IdentifyWithJSON(path string) (string, error) {
	return IdentifyAllWithJSON([]string{path})
}

func IdentifyAll(paths []string) ([]*Result, error) {
	return IdentifyAllWithOptions(paths, IdentifyOptions{Workers: 1})
}

func IdentifyAllWithOptions(paths []string, opts IdentifyOptions) ([]*Result, error) {
	fileResults, err := identifyAll(paths, opts)
	if err != nil {
		return nil, err
	}

	rs := make([]*Result, len(paths))
	for idx, result := range fileResults {
		rs[idx] = buildResult(result.path, result.ids, result.err)
	}

	return rs, nil
}

type fileResult struct {
	path string
	size int64
	mod  time.Time
	ids  []core.Identification
	err  error
}

func normalizeWorkers(workers int) (int, error) {
	if workers == 0 {
		return 1, nil
	}
	if workers < 1 || workers > maxWorkers {
		return 0, fmt.Errorf("workers must be between 1 and %d", maxWorkers)
	}
	return workers, nil
}

func identifyAll(paths []string, opts IdentifyOptions) ([]fileResult, error) {
	workers, err := normalizeWorkers(opts.Workers)
	if err != nil {
		return nil, err
	}

	sf := load()
	results := make([]fileResult, len(paths))

	var group errgroup.Group
	group.SetLimit(workers)
	for idx, path := range paths {
		group.Go(func() error {
			results[idx] = identifyPath(sf, path)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func identifyPath(sf *siegfried.Siegfried, path string) fileResult {
	result := fileResult{path: path}
	info, err := os.Stat(path)
	if err == nil {
		result.mod = info.ModTime()
		result.size = info.Size()
	}

	result.ids, result.err = identify(sf, path)
	return result
}

type jsonEscapedError struct {
	err error
}

func (e jsonEscapedError) Error() string {
	return escapeJSONStringContent(e.err.Error())
}

func escapeJSONError(err error) error {
	if err == nil {
		return nil
	}
	// Work around siegfried's JSON writer escaping filenames but interpolating
	// error strings directly. Remove this once upstream escapes errors too.
	return jsonEscapedError{err: err}
}

func escapeJSONStringContent(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return s
	}
	return string(b[1 : len(b)-1])
}

func IdentifyAllWithJSON(paths []string) (string, error) {
	return IdentifyAllWithJSONOptions(paths, IdentifyOptions{Workers: 1})
}

func IdentifyAllWithJSONOptions(paths []string, opts IdentifyOptions) (string, error) {
	scanDate := time.Now()
	results, err := identifyAll(paths, opts)
	if err != nil {
		return "", err
	}

	return writeJSONResults(results, scanDate), nil
}

func IdentifyDirWithJSON(path string, opts IdentifyDirOptions) (string, error) {
	scanDate := time.Now()
	paths, err := collectDirPaths(path, opts)
	if err != nil {
		return "", err
	}

	results, err := identifyAll(paths, IdentifyOptions{Workers: opts.Workers})
	if err != nil {
		return "", err
	}

	return writeJSONResults(results, scanDate), nil
}

func writeJSONResults(results []fileResult, scanDate time.Time) string {
	sf := load()

	var buf bytes.Buffer
	w := writer.JSON(&buf)

	w.Head(config.SignatureBase(), scanDate, sf.C, config.Version(), sf.Identifiers(), sf.Fields(), "")

	for _, result := range results {
		identifyErr := escapeJSONError(result.err)
		w.File(result.path, result.size, result.mod.Format(time.RFC3339), nil, identifyErr, result.ids)
	}

	w.Tail()

	return strings.TrimSpace(buf.String())
}

func collectDirPaths(root string, opts IdentifyDirOptions) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	if !opts.Recursive {
		return collectDirectChildPaths(root, opts.FollowSymlinks)
	}

	return collectRecursivePaths(root, opts.FollowSymlinks)
}

func collectRecursivePaths(root string, followSymlinks bool) ([]string, error) {
	var paths []string
	err := collectRecursiveDirPaths(root, followSymlinks, true, map[string]bool{}, &paths)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func collectRecursiveDirPaths(dir string, followSymlinks bool, root bool, visited map[string]bool, paths *[]string) error {
	key, err := filepath.EvalSymlinks(dir)
	if err != nil {
		if root {
			return err
		}
		*paths = append(*paths, dir)
		return nil
	}
	key, err = filepath.Abs(key)
	if err != nil {
		if root {
			return err
		}
		*paths = append(*paths, dir)
		return nil
	}
	key = filepath.Clean(key)
	if visited[key] {
		return nil
	}

	visited[key] = true

	entries, err := os.ReadDir(dir)
	if err != nil {
		if root {
			return err
		}
		*paths = append(*paths, dir)
		return nil
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.Type()&os.ModeSymlink != 0 {
			if !followSymlinks {
				continue
			}
			info, err := os.Stat(path)
			if err != nil {
				*paths = append(*paths, path)
				continue
			}
			if info.IsDir() {
				if err := collectRecursiveDirPaths(path, followSymlinks, false, visited, paths); err != nil {
					return err
				}
				continue
			}
			if info.Mode().IsRegular() {
				*paths = append(*paths, path)
			}
			continue
		}
		if entry.IsDir() {
			if err := collectRecursiveDirPaths(path, followSymlinks, false, visited, paths); err != nil {
				return err
			}
			continue
		}
		add, err := shouldIdentifyDirEntry(path, entry, followSymlinks)
		if err != nil || add {
			*paths = append(*paths, path)
		}
	}

	return nil
}

func collectDirectChildPaths(root string, followSymlinks bool) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		add, err := shouldIdentifyDirEntry(path, entry, followSymlinks)
		if err != nil || add {
			paths = append(paths, path)
		}
	}

	return paths, nil
}

func shouldIdentifyDirEntry(path string, entry os.DirEntry, followSymlinks bool) (bool, error) {
	mode := entry.Type()
	if mode&os.ModeSymlink != 0 {
		if !followSymlinks {
			return false, nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return true, err
		}
		return info.Mode().IsRegular(), nil
	}

	info, err := entry.Info()
	if err != nil {
		return true, err
	}
	return info.Mode().IsRegular(), nil
}

func Version() string {
	v := config.Version()
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}
