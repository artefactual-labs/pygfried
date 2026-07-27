// Command update refreshes pygfried's bundled Archivematica signature database
// from the exact Siegfried module version selected in go.mod.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	modulePath    = "github.com/richardlehane/siegfried"
	signaturePath = "cmd/roy/data/archivematica.sig"
)

type module struct {
	Path    string
	Version string
	Dir     string
}

type provenance struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	Source  string `json:"source"`
	SHA256  string `json:"sha256"`
}

func main() {
	check := flag.Bool("check", false, "verify that the bundled files are current")
	flag.Parse()

	if err := run(*check); err != nil {
		fmt.Fprintf(os.Stderr, "update Archivematica signature: %v\n", err)
		os.Exit(1)
	}
}

func run(check bool) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	mod, err := downloadModule(repoRoot)
	if err != nil {
		return err
	}
	if mod.Path != modulePath || mod.Version == "" || mod.Dir == "" {
		return fmt.Errorf("incomplete module information for %s", modulePath)
	}

	source := filepath.Join(mod.Dir, filepath.FromSlash(signaturePath))
	signature, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("open %s at %s: %w", modulePath, mod.Version, err)
	}

	hash := sha256.Sum256(signature)
	metadata := provenance{
		Module:  mod.Path,
		Version: mod.Version,
		Source:  signaturePath,
		SHA256:  hex.EncodeToString(hash[:]),
	}
	blob, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode provenance: %w", err)
	}
	blob = append(blob, '\n')

	destinationDir := filepath.Join(repoRoot, "internal", "signatures")
	destination := filepath.Join(destinationDir, "archivematica.sig")
	metadataPath := filepath.Join(destinationDir, "archivematica.json")

	if check {
		if err := checkFile(destination, signature); err != nil {
			return err
		}
		if err := checkFile(metadataPath, blob); err != nil {
			return err
		}
		fmt.Printf("Bundled Archivematica signature matches %s %s\n", mod.Path, mod.Version)
		return nil
	}

	if err := os.WriteFile(destination, signature, 0o644); err != nil {
		return fmt.Errorf("write signature database: %w", err)
	}
	if err := os.WriteFile(metadataPath, blob, 0o644); err != nil {
		return fmt.Errorf("write provenance: %w", err)
	}

	fmt.Printf("Updated %s from %s %s\n", destination, mod.Path, mod.Version)
	return nil
}

func checkFile(path string, want []byte) error {
	have, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w; run go run ./internal/signatures/cmd/update", path, err)
	}
	if !bytes.Equal(have, want) {
		return fmt.Errorf("%s is stale; run go run ./internal/signatures/cmd/update", path)
	}
	return nil
}

func findRepoRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate updater source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")), nil
}

func downloadModule(repoRoot string) (module, error) {
	cmd := exec.Command("go", "mod", "download", "-json", modulePath)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return module{}, fmt.Errorf("download module: %s", exitErr.Stderr)
		}
		return module{}, fmt.Errorf("download module: %w", err)
	}

	var mod module
	if err := json.Unmarshal(output, &mod); err != nil {
		return module{}, fmt.Errorf("decode module information: %w", err)
	}
	return mod, nil
}
