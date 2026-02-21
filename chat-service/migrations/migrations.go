package migrations

import "embed"

// FS holds all migration SQL files embedded at compile time.
// golang-migrate reads directly from this FS — no CLI or filesystem access needed at runtime.
//
//go:embed *.sql
var FS embed.FS
