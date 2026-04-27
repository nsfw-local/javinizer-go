package organizer

import (
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/scanner"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadataOnlyStrategy(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{}
	strategy := NewMetadataOnlyStrategy(fs, cfg)
	assert.NotNil(t, strategy)
	assert.NotNil(t, strategy.fs)
	assert.NotNil(t, strategy.config)
}

func TestMetadataOnlyStrategy_ImplementsInterface(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{}
	var _ OperationStrategy = NewMetadataOnlyStrategy(fs, cfg)
}

func TestMetadataOnlyStrategy_Plan(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{
		FileFormat: "<ID>",
		RenameFile: true,
	}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	match := matcher.MatchResult{
		ID: "ABC-123",
		File: scanner.FileInfo{
			Path:      "/source/ABC-123.mp4",
			Name:      "ABC-123.mp4",
			Extension: ".mp4",
		},
	}
	movie := &models.Movie{
		ID: "ABC-123",
	}

	plan, err := strategy.Plan(match, movie, "/dest", false)
	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, filepath.ToSlash("/source"), filepath.ToSlash(plan.TargetDir), "Should keep file in source directory")
	assert.Equal(t, filepath.ToSlash("/source/ABC-123.mp4"), filepath.ToSlash(plan.TargetPath), "Should preserve original filename even with RenameFile=true")
	assert.False(t, plan.WillMove, "Metadata-only mode should never set WillMove=true")
	assert.False(t, plan.InPlace, "MetadataOnlyStrategy should never set InPlace=true")
	assert.False(t, plan.IsDedicated, "MetadataOnlyStrategy should never set IsDedicated=true")
	assert.Contains(t, plan.SkipInPlaceReason, "metadata-only")
}

func TestMetadataOnlyStrategy_Plan_NoRename(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{
		RenameFile: false,
	}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	match := matcher.MatchResult{
		ID: "ABC-123",
		File: scanner.FileInfo{
			Path:      "/source/original-name.mp4",
			Name:      "original-name.mp4",
			Extension: ".mp4",
		},
	}
	movie := &models.Movie{
		ID: "ABC-123",
	}

	plan, err := strategy.Plan(match, movie, "/dest", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.ToSlash("/source/original-name.mp4"), filepath.ToSlash(plan.TargetPath))
	assert.False(t, plan.WillMove, "Metadata-only mode should never set WillMove=true")
}

func TestMetadataOnlyStrategy_Plan_IgnoresRenameFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{
		FileFormat: "<ID> <TITLE>",
		RenameFile: true,
	}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	match := matcher.MatchResult{
		ID: "ABC-123",
		File: scanner.FileInfo{
			Path:      "/source/original-name.mp4",
			Name:      "original-name.mp4",
			Extension: ".mp4",
		},
	}
	movie := &models.Movie{
		ID:    "ABC-123",
		Title: "Test Movie",
	}

	plan, err := strategy.Plan(match, movie, "/dest", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.ToSlash("/source/original-name.mp4"), filepath.ToSlash(plan.TargetPath), "Metadata-only mode should preserve original filename even with RenameFile=true")
	assert.False(t, plan.WillMove, "Metadata-only mode should never set WillMove=true")
	assert.Equal(t, filepath.ToSlash("/source"), filepath.ToSlash(plan.TargetDir))
}

func TestMetadataOnlyStrategy_Execute_NoMove(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	plan := &OrganizePlan{
		SourcePath: "/source/ABC-123.mp4",
		TargetDir:  "/source",
		TargetFile: "ABC-123.mp4",
		TargetPath: "/source/ABC-123.mp4",
		WillMove:   false,
		Conflicts:  []string{},
	}

	result, err := strategy.Execute(plan)
	require.NoError(t, err)
	assert.False(t, result.Moved, "Metadata-only should not move files")
	assert.Equal(t, filepath.ToSlash("/source/ABC-123.mp4"), filepath.ToSlash(result.NewPath))
}

func TestMetadataOnlyStrategy_Execute_NoMoveNoError(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	plan := &OrganizePlan{
		SourcePath: "/source/ABC-123.mp4",
		TargetDir:  "/source",
		TargetFile: "ABC-123.mp4",
		TargetPath: "/source/ABC-123.mp4",
		WillMove:   false,
		Conflicts:  nil,
	}

	result, err := strategy.Execute(plan)
	require.NoError(t, err)
	assert.False(t, result.Moved, "Metadata-only should not move files")
	assert.Equal(t, filepath.ToSlash("/source/ABC-123.mp4"), filepath.ToSlash(result.NewPath))
	assert.Equal(t, filepath.ToSlash("/source/ABC-123.mp4"), filepath.ToSlash(result.OriginalPath))
}

func TestMetadataOnlyStrategy_Plan_AlwaysNoConflicts(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := &config.OutputConfig{
		RenameFile: true,
	}
	strategy := NewMetadataOnlyStrategy(fs, cfg)

	_ = fs.MkdirAll("/source", 0777)
	_ = afero.WriteFile(fs, "/source/ABC-123.mp4", []byte("existing"), 0644)

	match := matcher.MatchResult{
		ID: "ABC-123",
		File: scanner.FileInfo{
			Path:      "/source/original.mp4",
			Name:      "original.mp4",
			Extension: ".mp4",
		},
	}
	movie := &models.Movie{
		ID: "ABC-123",
	}

	plan, err := strategy.Plan(match, movie, "/dest", false)
	require.NoError(t, err)
	assert.Nil(t, plan.Conflicts, "Metadata-only mode should never produce conflicts since it never renames")
	assert.False(t, plan.WillMove)

	planWithForce, err := strategy.Plan(match, movie, "/dest", true)
	require.NoError(t, err)
	assert.Nil(t, planWithForce.Conflicts, "ForceUpdate should also have no conflicts in metadata-only mode")
}
