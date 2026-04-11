package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/javinizer/javinizer-go/internal/types"
)

func init() {
	RegisterTestScraperConfigs()
}

func TestNormalize_OperationModeMigration(t *testing.T) {
	testCases := []struct {
		name            string
		renameInPlace   bool
		moveToFolder    bool
		renameFile      bool
		explicitMode    string
		wantMode        types.OperationMode
		wantModeChanged bool
	}{
		{
			name:            "rename true takes priority over move true",
			renameInPlace:   true,
			moveToFolder:    true,
			renameFile:      true,
			explicitMode:    "",
			wantMode:        types.OperationModeInPlace,
			wantModeChanged: true,
		},
		{
			name:            "rename true with move false",
			renameInPlace:   true,
			moveToFolder:    false,
			renameFile:      true,
			explicitMode:    "",
			wantMode:        types.OperationModeInPlace,
			wantModeChanged: true,
		},
		{
			name:            "move true with rename false",
			renameInPlace:   false,
			moveToFolder:    true,
			renameFile:      false,
			explicitMode:    "",
			wantMode:        types.OperationModeOrganize,
			wantModeChanged: true,
		},
		{
			name:            "both false with rename true gives norenamefolder",
			renameInPlace:   false,
			moveToFolder:    false,
			renameFile:      true,
			explicitMode:    "",
			wantMode:        types.OperationModeInPlaceNoRenameFolder,
			wantModeChanged: true,
		},
		{
			name:            "both false with rename false defaults to metadata-only",
			renameInPlace:   false,
			moveToFolder:    false,
			renameFile:      false,
			explicitMode:    "",
			wantMode:        types.OperationModeMetadataOnly,
			wantModeChanged: true,
		},
		{
			name:            "explicit organize mode not overridden",
			renameInPlace:   true,
			moveToFolder:    false,
			renameFile:      true,
			explicitMode:    "organize",
			wantMode:        types.OperationModeOrganize,
			wantModeChanged: false,
		},
		{
			name:            "explicit in-place mode not overridden",
			renameInPlace:   false,
			moveToFolder:    true,
			renameFile:      true,
			explicitMode:    "in-place",
			wantMode:        types.OperationModeInPlace,
			wantModeChanged: false,
		},
		{
			name:            "explicit preview mode not overridden even with booleans set",
			renameInPlace:   true,
			moveToFolder:    true,
			renameFile:      true,
			explicitMode:    "preview",
			wantMode:        types.OperationModePreview,
			wantModeChanged: false,
		},
		{
			name:            "explicit metadata-only mode not overridden",
			renameInPlace:   true,
			moveToFolder:    true,
			renameFile:      false,
			explicitMode:    "metadata-only",
			wantMode:        types.OperationModeMetadataOnly,
			wantModeChanged: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Output.RenameFolderInPlace = tc.renameInPlace
			cfg.Output.MoveToFolder = tc.moveToFolder
			cfg.Output.RenameFile = tc.renameFile
			cfg.Output.OperationMode = types.OperationMode(tc.explicitMode)

			modeBefore := cfg.Output.OperationMode
			Normalize(cfg)

			assert.Equal(t, tc.wantMode, cfg.Output.OperationMode, "OperationMode mismatch")

			if tc.wantModeChanged {
				assert.NotEqual(t, modeBefore, cfg.Output.OperationMode,
					"expected OperationMode to change from %q to %q", modeBefore, cfg.Output.OperationMode)
			} else {
				assert.Equal(t, tc.wantMode, modeBefore,
					"explicit OperationMode should not be changed, was %q before normalization", modeBefore)
			}
		})
	}
}

func TestNormalize_OperationModeMigration_ReturnsChanged(t *testing.T) {
	t.Run("migration with rename true sets OperationMode to in-place", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = true
		cfg.Output.MoveToFolder = false
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.Equal(t, types.OperationModeInPlace, cfg.Output.OperationMode, "OperationMode should be migrated to in-place")
	})

	t.Run("migration with move true sets OperationMode to organize", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = false
		cfg.Output.MoveToFolder = true
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.Equal(t, types.OperationModeOrganize, cfg.Output.OperationMode, "OperationMode should be migrated to organize")
	})

	t.Run("migration with both false and renameFile true sets OperationMode to in-place-norenamefolder", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = false
		cfg.Output.MoveToFolder = false
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.Equal(t, types.OperationModeInPlaceNoRenameFolder, cfg.Output.OperationMode, "OperationMode should be migrated to in-place-norenamefolder")
	})

	t.Run("migration with both false and renameFile false sets OperationMode to metadata-only", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = false
		cfg.Output.MoveToFolder = false
		cfg.Output.RenameFile = false

		Normalize(cfg)

		assert.Equal(t, types.OperationModeMetadataOnly, cfg.Output.OperationMode, "OperationMode should be migrated to metadata-only")
	})
}

func TestNormalize_OperationModeMigration_NoChangeWhenSet(t *testing.T) {
	t.Run("already set OperationMode is preserved through normalization", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = types.OperationModePreview
		cfg.Output.RenameFolderInPlace = true
		cfg.Output.MoveToFolder = true
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.Equal(t, types.OperationModePreview, cfg.Output.OperationMode, "explicit mode should not be overridden")
	})

	t.Run("empty OperationMode is always derived from booleans", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = false
		cfg.Output.MoveToFolder = false
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.NotEqual(t, types.OperationMode(""), cfg.Output.OperationMode, "empty OperationMode should be derived from boolean flags")
	})
}

func TestNormalize_OperationModeMigration_EdgeCaseBothTrue(t *testing.T) {
	t.Run("both booleans true gives in-place due to rename priority", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Output.OperationMode = ""
		cfg.Output.RenameFolderInPlace = true
		cfg.Output.MoveToFolder = true
		cfg.Output.RenameFile = true

		Normalize(cfg)

		assert.Equal(t, types.OperationModeInPlace, cfg.Output.OperationMode,
			"rename_folder_in_place=true should take priority over move_to_folder=true")
	})
}
