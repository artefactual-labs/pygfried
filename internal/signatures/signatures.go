// Package signatures provides signature databases bundled with pygfried.
package signatures

import _ "embed"

// Archivematica is the complete Archivematica signature database shipped with
// the version of Siegfried required by this module.
//
//go:embed archivematica.sig
var Archivematica []byte

// ArchivematicaProvenance records where the bundled database came from.
//
//go:embed archivematica.json
var ArchivematicaProvenance []byte
