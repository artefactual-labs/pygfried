package pygfried_test

import (
	"encoding/json"
	"fmt"
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

func TestVersion(t *testing.T) {
	v := config.Version()
	have, want := pygfried.Version(), fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])

	assert.Equal(t, have, want)
}
