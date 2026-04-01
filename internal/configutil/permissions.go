package configutil

// File and directory permission constants
// Centralized to ensure consistency across the codebase
const (
	// DirPermConfig is the permission mode for configuration directories (owner + group read/execute)
	DirPermConfig = 0755
	// DirPermTemp is the permission mode for temporary/sensitive directories (owner-only access)
	DirPermTemp = 0700
	// FilePermConfig is the permission mode for configuration files
	FilePermConfig = 0644

	// CurrentConfigVersion tracks compatibility breakpoints for on-disk config.
	// Do not bump for additive/default-only fields; those are handled by loading
	// into DefaultConfig() and idempotent normalization rules.
	CurrentConfigVersion = 3
)
