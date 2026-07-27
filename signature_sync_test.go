package pygfried_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/artefactual-labs/pygfried/internal/signatures"

	"gotest.tools/v3/assert"
)

const (
	siegfriedModule = "github.com/richardlehane/siegfried"
	updateCommand   = "go run ./internal/signatures/cmd/update"
)

func TestArchivematicaSignatureMatchesSiegfriedDependency(t *testing.T) {
	var provenance struct {
		Module  string `json:"module"`
		Version string `json:"version"`
		Source  string `json:"source"`
		SHA256  string `json:"sha256"`
	}
	err := json.Unmarshal(signatures.ArchivematicaProvenance, &provenance)
	assert.NilError(t, err)

	goMod, err := os.ReadFile("go.mod")
	assert.NilError(t, err)
	versionPattern := regexp.MustCompile(
		`(?m)^\s*` + regexp.QuoteMeta(siegfriedModule) + `\s+(v[^\s]+)`,
	)
	match := versionPattern.FindSubmatch(goMod)
	if match == nil {
		t.Fatalf("could not find %s in go.mod", siegfriedModule)
	}

	assert.Equal(t, provenance.Module, siegfriedModule)
	assert.Equal(
		t,
		provenance.Version,
		string(match[1]),
		"bundled signature version does not match go.mod; run %s",
		updateCommand,
	)
	assert.Equal(t, provenance.Source, "cmd/roy/data/archivematica.sig")

	sum := sha256.Sum256(signatures.Archivematica)
	assert.Equal(
		t,
		provenance.SHA256,
		hex.EncodeToString(sum[:]),
		"bundled signature checksum is stale; run %s",
		updateCommand,
	)
}
