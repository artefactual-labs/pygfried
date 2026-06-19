package pygfried

import (
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

type detailedJSONResult struct {
	Siegfried   string                   `json:"siegfried"`
	ScanDate    string                   `json:"scandate"`
	Signature   string                   `json:"signature"`
	Created     string                   `json:"created"`
	Identifiers []detailedJSONIdentifier `json:"identifiers"`
	Files       []detailedJSONFile       `json:"files"`
}

type detailedJSONIdentifier struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

type detailedJSONFile struct {
	Filename string              `json:"filename"`
	FileSize int64               `json:"filesize"`
	Modified string              `json:"modified"`
	Errors   string              `json:"errors"`
	Matches  []map[string]string `json:"matches"`
}

func TestBuildDetailedResultMatchesDecodedJSON(t *testing.T) {
	scanDate := time.Date(2026, 6, 19, 8, 30, 0, 0, time.UTC)
	results, err := identifyAll([]string{"README.md", "pyproject.toml"}, IdentifyOptions{Workers: 2})
	assert.NilError(t, err)

	want := decodeDetailedJSON(t, writeJSONResults(results, scanDate))
	have := detailedResultToJSONShape(buildDetailedResult(results, scanDate))

	assert.DeepEqual(t, have, want)
}

func TestBuildDetailedResultMatchesDecodedJSONErrors(t *testing.T) {
	scanDate := time.Date(2026, 6, 19, 8, 30, 0, 0, time.UTC)
	results, err := identifyAll([]string{`C:\Users\j472\Documents\missing`}, IdentifyOptions{Workers: 1})
	assert.NilError(t, err)

	want := decodeDetailedJSON(t, writeJSONResults(results, scanDate))
	have := detailedResultToJSONShape(buildDetailedResult(results, scanDate))

	assert.DeepEqual(t, have, want)
}

func decodeDetailedJSON(t *testing.T, blob string) detailedJSONResult {
	t.Helper()

	var result detailedJSONResult
	err := json.Unmarshal([]byte(blob), &result)
	assert.NilError(t, err)
	return result
}

func detailedResultToJSONShape(result *DetailedResult) detailedJSONResult {
	out := detailedJSONResult{
		Siegfried:   result.Siegfried,
		ScanDate:    result.ScanDate,
		Signature:   result.Signature,
		Created:     result.Created,
		Identifiers: make([]detailedJSONIdentifier, len(result.Identifiers)),
		Files:       make([]detailedJSONFile, len(result.Files)),
	}
	for idx, identifier := range result.Identifiers {
		out.Identifiers[idx] = detailedJSONIdentifier{
			Name:    identifier.Name,
			Details: identifier.Details,
		}
	}
	for idx, file := range result.Files {
		out.Files[idx] = detailedJSONFile{
			Filename: file.Filename,
			FileSize: file.FileSize,
			Modified: file.Modified,
			Errors:   file.Errors,
			Matches:  make([]map[string]string, len(file.Matches)),
		}
		for matchIdx, match := range file.Matches {
			fields := make(map[string]string, len(match.Fields))
			for _, field := range match.Fields {
				fields[field.Name] = field.Value
			}
			out.Files[idx].Matches[matchIdx] = fields
		}
	}
	return out
}
