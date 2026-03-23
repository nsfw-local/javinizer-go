package actress_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/javinizer/javinizer-go/cmd/javinizer/commands/actress"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		outC <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	fn()

	require.NoError(t, wOut.Close())
	require.NoError(t, wErr.Close())

	return <-outC, <-errC
}

func setupActressTestDB(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data", "test.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))

	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	configPath := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, config.Save(testCfg, configPath))

	db, err := database.New(testCfg)
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate())
	require.NoError(t, db.Close())

	return configPath, dbPath
}

func seedActresses(t *testing.T, configPath string, actresses ...*models.Actress) {
	t.Helper()
	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	db, err := database.New(cfg)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := database.NewActressRepository(db)
	for _, actress := range actresses {
		require.NoError(t, repo.Create(actress))
	}
}

func TestActressCommand_MergeFlags(t *testing.T) {
	cmd := actress.NewCommand()
	mergeCmd, _, err := cmd.Find([]string{"merge"})
	require.NoError(t, err)
	require.NotNil(t, mergeCmd)

	assert.NotNil(t, mergeCmd.Flags().Lookup("target"))
	assert.NotNil(t, mergeCmd.Flags().Lookup("source"))
	assert.NotNil(t, mergeCmd.Flags().Lookup("non-interactive"))
	assert.NotNil(t, mergeCmd.Flags().Lookup("prefer"))
	assert.NotNil(t, mergeCmd.Flags().Lookup("yes"))
}

func TestActressCommand_MergeNonInteractive(t *testing.T) {
	configPath, _ := setupActressTestDB(t)
	target := &models.Actress{DMMID: 80001, FirstName: "Target", LastName: "Actor", JapaneseName: "ターゲット"}
	source := &models.Actress{DMMID: 80002, FirstName: "Source", LastName: "Actor", JapaneseName: "ソース"}
	seedActresses(t, configPath, target, source)

	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.PersistentFlags().String("config", configPath, "config file")
	rootCmd.AddCommand(actress.NewCommand())
	rootCmd.SetArgs([]string{
		"actress", "merge",
		"--target", strconv.FormatUint(uint64(target.ID), 10),
		"--source", strconv.FormatUint(uint64(source.ID), 10),
		"--non-interactive",
		"--prefer", "source",
	})

	stdout, _ := captureOutput(t, func() {
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Merged actress")

	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	db, err := database.New(cfg)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := database.NewActressRepository(db)

	merged, err := repo.FindByID(target.ID)
	require.NoError(t, err)
	assert.Equal(t, 80002, merged.DMMID)

	_, err = repo.FindByID(source.ID)
	require.Error(t, err)
}

func TestActressCommand_MergeInteractive(t *testing.T) {
	configPath, _ := setupActressTestDB(t)
	target := &models.Actress{DMMID: 0, FirstName: "Target", LastName: "Actor", JapaneseName: "共通"}
	source := &models.Actress{DMMID: 1, FirstName: "Source", LastName: "Actor", JapaneseName: "共通"}
	seedActresses(t, configPath, target, source)

	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.PersistentFlags().String("config", configPath, "config file")
	rootCmd.AddCommand(actress.NewCommand())
	rootCmd.SetArgs([]string{
		"actress", "merge",
		"--target", strconv.FormatUint(uint64(target.ID), 10),
		"--source", strconv.FormatUint(uint64(source.ID), 10),
		"--yes",
	})
	rootCmd.SetIn(strings.NewReader("s\n"))

	stdout, _ := captureOutput(t, func() {
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Detected 1 conflicting field")
	assert.Contains(t, stdout, "Merged actress")

	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	db, err := database.New(cfg)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	repo := database.NewActressRepository(db)

	merged, err := repo.FindByID(target.ID)
	require.NoError(t, err)
	assert.Equal(t, "Source", merged.FirstName)
}
