package pygfried_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artefactual-labs/pygfried"

	"github.com/richardlehane/siegfried/pkg/config"
	"gotest.tools/v3/assert"
)

func TestIdentify(t *testing.T) {
	result, err := pygfried.Identify("pyproject.toml")

	assert.NilError(t, err)
	assert.DeepEqual(t, result, &pygfried.Result{
		Path:        "pyproject.toml",
		Identifiers: []string{"fmt/2065"},
		Known:       true,
	})
}

func TestIdentifyWithJSON(t *testing.T) {
	blob, err := pygfried.IdentifyWithJSON("pyproject.toml")

	assert.NilError(t, err)

	type Match struct {
		ID string `json:"id"`
	}
	type File struct {
		Filename string  `json:"filename"`
		Matches  []Match `json:"matches"`
	}
	type Response struct {
		Files []File `json:"files"`
	}

	var response Response
	err = json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)

	assert.Equal(t, len(response.Files), 1)
	assert.DeepEqual(t, response, Response{
		Files: []File{
			{
				Filename: "pyproject.toml",
				Matches:  []Match{{ID: "fmt/2065"}},
			},
		},
	})
}

func TestIdentifyAll(t *testing.T) {
	paths := []string{"README.md", "pyproject.toml"}
	results, err := pygfried.IdentifyAll(paths)

	assert.NilError(t, err)
	assert.DeepEqual(t, results, []*pygfried.Result{
		{Path: "README.md", Identifiers: []string{"fmt/1149"}, Known: true},
		{Path: "pyproject.toml", Identifiers: []string{"fmt/2065"}, Known: true},
	})
}

func TestIdentifyAllWithJSON(t *testing.T) {
	paths := []string{"README.md", "pyproject.toml"}
	blob, err := pygfried.IdentifyAllWithJSON(paths)

	assert.NilError(t, err)

	type Match struct {
		ID string `json:"id"`
	}
	type File struct {
		Filename string  `json:"filename"`
		Matches  []Match `json:"matches"`
	}
	type Response struct {
		Files []File `json:"files"`
	}

	var response Response
	err = json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)

	assert.Equal(t, len(response.Files), 2)
	assert.DeepEqual(t, response, Response{
		Files: []File{
			{
				Filename: "README.md",
				Matches:  []Match{{ID: "fmt/1149"}},
			},
			{
				Filename: "pyproject.toml",
				Matches:  []Match{{ID: "fmt/2065"}},
			},
		},
	})
}

func TestIdentifyAllWithJSONOptionsEmpty(t *testing.T) {
	blob, err := pygfried.IdentifyAllWithJSONOptions(nil, pygfried.IdentifyOptions{Workers: 2})

	assert.NilError(t, err)

	type Response struct {
		Files []struct{} `json:"files"`
	}
	var response Response
	err = json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)
	assert.Equal(t, len(response.Files), 0)
}

func TestIdentifyAllWithJSONOptionsWorkers(t *testing.T) {
	paths := []string{"README.md", "pyproject.toml"}
	blob, err := pygfried.IdentifyAllWithJSONOptions(paths, pygfried.IdentifyOptions{Workers: 2})

	assert.NilError(t, err)

	type File struct {
		Filename string `json:"filename"`
	}
	type Response struct {
		Files []File `json:"files"`
	}
	var response Response
	err = json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)

	assert.DeepEqual(t, response.Files, []File{
		{Filename: "README.md"},
		{Filename: "pyproject.toml"},
	})
}

func TestIdentifyAllWithJSONOptionsInvalidWorkers(t *testing.T) {
	_, err := pygfried.IdentifyAllWithJSONOptions(nil, pygfried.IdentifyOptions{Workers: -1})

	assert.Error(t, err, "workers must be between 1 and 1024")
}

func TestIdentifyAllWithJSONEscapesErrorStrings(t *testing.T) {
	path := `C:\Users\j472\Documents\missing`
	blob, err := pygfried.IdentifyAllWithJSON([]string{path})

	assert.NilError(t, err)

	type File struct {
		Filename string `json:"filename"`
		Errors   string `json:"errors"`
	}
	type Response struct {
		Files []File `json:"files"`
	}

	var response Response
	err = json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)

	assert.Equal(t, len(response.Files), 1)
	assert.Equal(t, response.Files[0].Filename, path)
	assert.Assert(t, strings.Contains(response.Files[0].Errors, path))
}

func TestIdentifyDirWithJSONRecursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.py"))
	nested := filepath.Join(dir, "nested")
	assert.NilError(t, os.Mkdir(nested, 0o755))
	writeFile(t, filepath.Join(nested, "b.py"))

	blob, err := pygfried.IdentifyDirWithJSON(dir, pygfried.IdentifyDirOptions{
		Recursive: true,
		Workers:   2,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{
		filepath.Join(dir, "a.py"),
		filepath.Join(nested, "b.py"),
	})
}

func TestIdentifyDirWithJSONNonRecursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.py"))
	nested := filepath.Join(dir, "nested")
	assert.NilError(t, os.Mkdir(nested, 0o755))
	writeFile(t, filepath.Join(nested, "b.py"))

	blob, err := pygfried.IdentifyDirWithJSON(dir, pygfried.IdentifyDirOptions{
		Recursive: false,
		Workers:   1,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{filepath.Join(dir, "a.py")})
}

func TestIdentifyDirWithJSONSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.py")
	link := filepath.Join(dir, "link.py")
	writeFile(t, target)
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	blob, err := pygfried.IdentifyDirWithJSON(dir, pygfried.IdentifyDirOptions{
		Recursive:      true,
		Workers:        1,
		FollowSymlinks: false,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{target})
}

func TestIdentifyDirWithJSONFollowsDirectorySymlinks(t *testing.T) {
	root := t.TempDir()
	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "target.py")
	link := filepath.Join(root, "linked")
	writeFile(t, target)
	if err := os.Symlink(targetDir, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	blob, err := pygfried.IdentifyDirWithJSON(root, pygfried.IdentifyDirOptions{
		Recursive:      true,
		Workers:        1,
		FollowSymlinks: true,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{filepath.Join(link, "target.py")})
}

func TestIdentifyDirWithJSONSkipsSymlinkCycles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.py")
	link := filepath.Join(dir, "loop")
	writeFile(t, target)
	if err := os.Symlink(dir, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	blob, err := pygfried.IdentifyDirWithJSON(dir, pygfried.IdentifyDirOptions{
		Recursive:      true,
		Workers:        1,
		FollowSymlinks: true,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{target})
}

func TestIdentifyDirWithJSONSkipsRepeatedSymlinkTargets(t *testing.T) {
	root := t.TempDir()
	targetDir := t.TempDir()
	target := filepath.Join(targetDir, "target.py")
	first := filepath.Join(root, "a")
	second := filepath.Join(root, "b")
	writeFile(t, target)
	if err := os.Symlink(targetDir, first); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	assert.NilError(t, os.Symlink(targetDir, second))

	blob, err := pygfried.IdentifyDirWithJSON(root, pygfried.IdentifyDirOptions{
		Recursive:      true,
		Workers:        1,
		FollowSymlinks: true,
	})

	assert.NilError(t, err)
	assert.DeepEqual(t, filenames(t, blob), []string{filepath.Join(first, "target.py")})
}

func TestIdentifyDirWithJSONRequiresDirectory(t *testing.T) {
	_, err := pygfried.IdentifyDirWithJSON("pyproject.toml", pygfried.IdentifyDirOptions{
		Recursive: true,
		Workers:   1,
	})

	assert.Error(t, err, "pyproject.toml is not a directory")
}

func TestVersion(t *testing.T) {
	v := config.Version()
	have, want := pygfried.Version(), fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])

	assert.Equal(t, have, want)
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	assert.NilError(t, os.WriteFile(path, []byte("print('hello')\n"), 0o644))
}

func filenames(t *testing.T, blob string) []string {
	t.Helper()

	type File struct {
		Filename string `json:"filename"`
	}
	type Response struct {
		Files []File `json:"files"`
	}
	var response Response
	err := json.Unmarshal([]byte(blob), &response)
	assert.NilError(t, err)

	names := make([]string, len(response.Files))
	for idx, file := range response.Files {
		names[idx] = file.Filename
	}
	return names
}
