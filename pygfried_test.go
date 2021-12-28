package pygfried_test

import (
	"fmt"
	"testing"

	"github.com/artefactual-labs/pygfried"

	"github.com/richardlehane/siegfried/pkg/config"
	"gotest.tools/v3/assert"
)

func TestIdentify(t *testing.T) {
	result, err := pygfried.Identify("default.sig")

	assert.NilError(t, err)
	assert.DeepEqual(t, result, &pygfried.Result{
		Path:        "default.sig",
		Identifiers: []string{"fmt/883"},
		Known:       true,
	})
}

func TestIdentifyAll(t *testing.T) {
	paths := []string{"README.md", "default.sig"}
	results, err := pygfried.IdentifyAll(paths)

	assert.NilError(t, err)
	assert.DeepEqual(t, results, []*pygfried.Result{
		{Path: "README.md", Identifiers: []string{"fmt/1149"}, Known: true},
		{Path: "default.sig", Identifiers: []string{"fmt/883"}, Known: true},
	})
}

func TestVersion(t *testing.T) {
	v := config.Version()
	have, want := pygfried.Version(), fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])

	assert.Equal(t, have, want)
}
