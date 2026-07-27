package pygfried

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func TestScannerJSONEscapesSignature(t *testing.T) {
	scanner, err := NewScanner(ScannerOptions{})
	assert.NilError(t, err)
	scanner.signature = "custom\"\n.sig"

	blob, err := scanner.IdentifyWithJSON("pyproject.toml")
	assert.NilError(t, err)

	var response struct {
		Signature string `json:"signature"`
	}
	assert.NilError(t, json.Unmarshal([]byte(blob), &response))
	assert.Equal(t, response.Signature, scanner.signature)
}
