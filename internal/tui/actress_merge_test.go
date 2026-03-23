package tui

import (
	"strconv"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTUI_BrowserKeyOpenActressMergeModal(t *testing.T) {
	_, cfg := testutil.CreateTestConfig(t, nil)
	model := New(cfg)
	model.SetActressRepo(&database.ActressRepository{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	got := updated.(*Model)

	assert.True(t, got.showingActressMerge)
	assert.Equal(t, actressMergeStepInput, got.actressMergeStep)
	assert.Equal(t, 0, got.actressMergeFocus)
}

func TestTUI_BrowserKeyOpenActressMergeModalWithoutRepo(t *testing.T) {
	_, cfg := testutil.CreateTestConfig(t, nil)
	model := New(cfg)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	got := updated.(*Model)

	assert.False(t, got.showingActressMerge)
}

func newActressRepoForTUIMergeTest(t *testing.T) *database.ActressRepository {
	t.Helper()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := database.New(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	require.NoError(t, db.AutoMigrate())
	return database.NewActressRepository(db)
}

func TestTUI_ActressMergeConflictSelectionAndApply(t *testing.T) {
	repo := newActressRepoForTUIMergeTest(t)

	target := &models.Actress{
		DMMID:        74001,
		FirstName:    "Target",
		LastName:     "Actress",
		JapaneseName: "ターゲット",
	}
	source := &models.Actress{
		DMMID:        74002,
		FirstName:    "Source",
		LastName:     "Actress",
		JapaneseName: "ソース",
	}
	require.NoError(t, repo.Create(target))
	require.NoError(t, repo.Create(source))

	_, cfg := testutil.CreateTestConfig(t, nil)
	model := New(cfg)
	model.SetActressRepo(repo)
	model.openActressMergeModal()
	model.actressMergeTargetInput.SetValue(strconv.FormatUint(uint64(target.ID), 10))
	model.actressMergeSourceInput.SetValue(strconv.FormatUint(uint64(source.ID), 10))

	require.NoError(t, model.loadActressMergePreview())
	require.Equal(t, actressMergeStepConflict, model.actressMergeStep)
	require.NotNil(t, model.actressMergePreview)
	require.NotEmpty(t, model.actressMergePreview.Conflicts)

	// Select a deterministic conflict field and choose "source".
	firstNameConflictIdx := -1
	for i, conflict := range model.actressMergePreview.Conflicts {
		if conflict.Field == "first_name" {
			firstNameConflictIdx = i
			break
		}
	}
	require.NotEqual(t, -1, firstNameConflictIdx)
	model.actressMergeConflictCursor = firstNameConflictIdx

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updated.(*Model)
	assert.Equal(t, "source", model.actressMergeResolutions["first_name"])

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	assert.Equal(t, actressMergeStepResult, model.actressMergeStep)
	require.NotNil(t, model.actressMergeResult)

	merged, err := repo.FindByID(target.ID)
	require.NoError(t, err)
	assert.Equal(t, "Source", merged.FirstName)

	_, err = repo.FindByID(source.ID)
	require.Error(t, err)
}
