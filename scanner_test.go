package pygfried_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/artefactual-labs/pygfried"
	"github.com/artefactual-labs/pygfried/internal/signatures"

	"gotest.tools/v3/assert"
)

func TestNewScannerDefaultsToDefaultProfile(t *testing.T) {
	scanner, err := pygfried.NewScanner(pygfried.ScannerOptions{})

	assert.NilError(t, err)
	assert.Equal(t, scanner.Profile(), pygfried.ProfileDefault)
	assert.Equal(t, scanner.Signature(), "default.sig")

	result, err := scanner.Identify("pyproject.toml")
	assert.NilError(t, err)
	assert.DeepEqual(t, result.Identifiers, []string{"fmt/2065"})
}

func TestNewScannerRejectsConflictingOptions(t *testing.T) {
	_, err := pygfried.NewScanner(pygfried.ScannerOptions{
		Profile:   pygfried.ProfileDefault,
		Signature: "custom.sig",
	})

	assert.Error(t, err, "profile and signature are mutually exclusive")
}

func TestNewScannerRejectsUnknownProfile(t *testing.T) {
	_, err := pygfried.NewScanner(pygfried.ScannerOptions{Profile: "unknown"})

	assert.Error(t, err, `unknown profile "unknown"`)
}

func TestArchivematicaScannerIdentifiesExtendedSignatures(t *testing.T) {
	scanner, err := pygfried.NewScanner(pygfried.ScannerOptions{
		Profile: pygfried.ProfileArchivematica,
	})
	assert.NilError(t, err)
	assert.Equal(t, scanner.Profile(), pygfried.ProfileArchivematica)
	assert.Equal(t, scanner.Signature(), "archivematica.sig")

	dir := t.TempDir()
	cases := []struct {
		name       string
		content    []byte
		identifier string
	}{
		{"sample.ad1", []byte("ADSEGMENTEDFILE"), "archivematica-fmt/2"},
		{"encrypted.ad1", []byte("ADCRYPT"), "archivematica-fmt/3"},
		{"sample.dd", []byte("small raw image"), "archivematica-fmt/4"},
		{"sample.aff", []byte("AFF10\r\n\x00"), "archivematica-fmt/5"},
	}
	for _, test := range cases {
		path := filepath.Join(dir, test.name)
		assert.NilError(t, os.WriteFile(path, test.content, 0o644))
		assertIdentifier(t, scanner, path, test.identifier)
	}
}

func TestDefaultAndArchivematicaScannersAreIndependent(t *testing.T) {
	defaultScanner, err := pygfried.NewScanner(pygfried.ScannerOptions{})
	assert.NilError(t, err)
	archivematicaScanner, err := pygfried.NewScanner(pygfried.ScannerOptions{
		Profile: pygfried.ProfileArchivematica,
	})
	assert.NilError(t, err)

	path := filepath.Join(t.TempDir(), "sample.ad1")
	assert.NilError(t, os.WriteFile(path, []byte("ADSEGMENTEDFILE"), 0o644))

	assertIdentifier(t, defaultScanner, path, "fmt/842")
	assertIdentifier(t, archivematicaScanner, path, "archivematica-fmt/2")
	assertIdentifier(t, defaultScanner, path, "fmt/842")
}

func TestExternalScannerLoadsSignatureEagerly(t *testing.T) {
	signature := filepath.Join(t.TempDir(), "custom.sig")
	assert.NilError(t, os.WriteFile(signature, signatures.Archivematica, 0o644))

	scanner, err := pygfried.NewScanner(pygfried.ScannerOptions{Signature: signature})
	assert.NilError(t, err)
	assert.Equal(t, scanner.Profile(), "")
	assert.Equal(t, scanner.Signature(), "custom.sig")
	assert.NilError(t, os.Remove(signature))

	path := filepath.Join(t.TempDir(), "sample.ad1")
	assert.NilError(t, os.WriteFile(path, []byte("ADSEGMENTEDFILE"), 0o644))
	assertIdentifier(t, scanner, path, "archivematica-fmt/2")
}

func TestExternalScannerRejectsInvalidSignature(t *testing.T) {
	signature := filepath.Join(t.TempDir(), "invalid.sig")
	assert.NilError(t, os.WriteFile(signature, []byte("not a signature"), 0o644))

	_, err := pygfried.NewScanner(pygfried.ScannerOptions{Signature: signature})

	assert.ErrorContains(t, err, "not a siegfried signature file")
	assert.NilError(t, os.Remove(signature))
}

func TestExternalScannerRejectsMissingSignature(t *testing.T) {
	signature := filepath.Join(t.TempDir(), "missing.sig")

	_, err := pygfried.NewScanner(pygfried.ScannerOptions{Signature: signature})

	assert.ErrorContains(t, err, "error opening signature file")
}

func TestScannerJSONReportsSelectedSignature(t *testing.T) {
	scanner, err := pygfried.NewScanner(pygfried.ScannerOptions{
		Profile: pygfried.ProfileArchivematica,
	})
	assert.NilError(t, err)

	blob, err := scanner.IdentifyWithJSON("pyproject.toml")
	assert.NilError(t, err)

	var response struct {
		Signature string `json:"signature"`
	}
	assert.NilError(t, json.Unmarshal([]byte(blob), &response))
	assert.Equal(t, response.Signature, "archivematica.sig")
}

func TestScannerSupportsConcurrentUse(t *testing.T) {
	scanner, err := pygfried.NewScanner(pygfried.ScannerOptions{
		Profile: pygfried.ProfileArchivematica,
	})
	assert.NilError(t, err)

	path := filepath.Join(t.TempDir(), "sample.ad1")
	assert.NilError(t, os.WriteFile(path, []byte("ADSEGMENTEDFILE"), 0o644))

	const goroutines = 16
	const repetitions = 10
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range repetitions {
				result, err := scanner.Identify(path)
				if err != nil {
					errors <- err
					return
				}
				if len(result.Identifiers) != 1 ||
					result.Identifiers[0] != "archivematica-fmt/2" {
					errors <- &unexpectedIdentifiersError{result.Identifiers}
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}

type unexpectedIdentifiersError struct {
	identifiers []string
}

func (e *unexpectedIdentifiersError) Error() string {
	blob, _ := json.Marshal(e.identifiers)
	return "unexpected identifiers: " + string(blob)
}

func assertIdentifier(t *testing.T, scanner *pygfried.Scanner, path, want string) {
	t.Helper()
	result, err := scanner.Identify(path)
	assert.NilError(t, err)
	assert.DeepEqual(t, result.Identifiers, []string{want})
}
