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
	result, err := pygfried.Identify("setup.cfg")

	assert.NilError(t, err)
	assert.DeepEqual(t, result, &pygfried.Result{
		Path:        "setup.cfg",
		Identifiers: []string{"UNKNOWN"},
		Known:       false,
	})
}

func TestIdentifyWithJSON(t *testing.T) {
	blob, err := pygfried.IdentifyWithJSON("setup.cfg")

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
				Filename: "setup.cfg",
				Matches:  []Match{{ID: "UNKNOWN"}},
			},
		},
	})
}

func TestIdentifyAll(t *testing.T) {
	paths := []string{"README.md", "setup.cfg"}
	results, err := pygfried.IdentifyAll(paths)

	assert.NilError(t, err)
	assert.DeepEqual(t, results, []*pygfried.Result{
		{Path: "README.md", Identifiers: []string{"fmt/1149"}, Known: true},
		{Path: "setup.cfg", Identifiers: []string{"UNKNOWN"}, Known: false},
	})
}

func TestIdentifyAllWithJSON(t *testing.T) {
	paths := []string{"README.md", "setup.cfg"}
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
				Filename: "setup.cfg",
				Matches:  []Match{{ID: "UNKNOWN"}},
			},
		},
	})
}

func TestVersion(t *testing.T) {
	v := config.Version()
	have, want := pygfried.Version(), fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])

	assert.Equal(t, have, want)
}
