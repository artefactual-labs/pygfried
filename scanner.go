package pygfried

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/artefactual-labs/pygfried/internal/signatures"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/static"
)

const (
	ProfileDefault       = "default"
	ProfileArchivematica = "archivematica"
)

// ScannerOptions selects the signature database used by a Scanner.
//
// Profile and Signature are mutually exclusive. An empty Profile selects the
// default profile. Signature names a complete external Siegfried database.
type ScannerOptions struct {
	Profile   string
	Signature string
}

// Scanner identifies files using one immutable Siegfried signature database.
//
// A Scanner may be reused by multiple goroutines. Its Siegfried engine is
// fully loaded before NewScanner returns and is never mutated afterward.
type Scanner struct {
	sf        *siegfried.Siegfried
	profile   string
	signature string
}

var (
	defaultScanner     *Scanner
	defaultScannerOnce sync.Once
)

// NewScanner constructs a Scanner and eagerly loads its signature database.
func NewScanner(opts ScannerOptions) (*Scanner, error) {
	if opts.Profile != "" && opts.Signature != "" {
		return nil, fmt.Errorf("profile and signature are mutually exclusive")
	}

	if opts.Signature != "" {
		file, err := os.Open(opts.Signature)
		if err != nil {
			return nil, fmt.Errorf("error opening signature file: %w", err)
		}
		defer file.Close()

		sf, err := siegfried.LoadReader(file)
		if err != nil {
			return nil, err
		}
		return &Scanner{
			sf:        sf,
			signature: filepath.Base(opts.Signature),
		}, nil
	}

	profile := opts.Profile
	if profile == "" {
		profile = ProfileDefault
	}

	switch profile {
	case ProfileDefault:
		return &Scanner{
			sf:        static.New(),
			profile:   ProfileDefault,
			signature: "default.sig",
		}, nil
	case ProfileArchivematica:
		sf, err := siegfried.LoadReader(bytes.NewReader(signatures.Archivematica))
		if err != nil {
			return nil, fmt.Errorf("load embedded Archivematica signature: %w", err)
		}
		return &Scanner{
			sf:        sf,
			profile:   ProfileArchivematica,
			signature: "archivematica.sig",
		}, nil
	default:
		return nil, fmt.Errorf("unknown profile %q", profile)
	}
}

// Profile returns the selected bundled profile. It is empty when an external
// signature database is in use.
func (s *Scanner) Profile() string {
	return s.profile
}

// Signature returns the logical filename of the selected signature database.
func (s *Scanner) Signature() string {
	return s.signature
}

func loadDefaultScanner() *Scanner {
	defaultScannerOnce.Do(func() {
		scanner, err := NewScanner(ScannerOptions{})
		if err != nil {
			panic(err)
		}
		defaultScanner = scanner
	})
	return defaultScanner
}
