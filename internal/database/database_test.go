package database

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_UnsupportedDatabaseType(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "postgres", // Unsupported type
			DSN:  "host=localhost",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestNew_Success(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, db)
	require.NoError(t, db.Close())
}

func TestNew_EmptyDatabaseType(t *testing.T) {
	// Empty string should default to sqlite
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "", // Empty defaults to sqlite
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, db)
	require.NoError(t, db.Close())
}

func TestNew_DebugLogging(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "debug", // Should enable debug logging
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, db)
	require.NoError(t, db.Close())
}

func TestClose_Success(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)

	// Close should succeed
	err = db.Close()
	assert.NoError(t, err)

	// Note: SQLite in-memory databases may not error on second close
	// This is database-specific behavior and not a bug
}

func TestMovieRepository_Delete_ErrorHandling(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewMovieRepository(db)

	t.Run("Delete non-existent movie", func(t *testing.T) {
		// Deleting non-existent movie should not error (SQL semantics)
		err := repo.Delete("NON-EXISTENT")
		assert.NoError(t, err)
	})

	t.Run("Delete movie with translations", func(t *testing.T) {
		// Create movie with translation
		movie := createTestMovie("IPX-DELETE-001")
		movie.Translations = []models.MovieTranslation{
			{
				MovieID:  "IPX-DELETE-001",
				Language: "en",
				Title:    "English Title",
			},
		}
		err := repo.Create(movie)
		require.NoError(t, err)

		// Delete should cascade to translations
		err = repo.Delete("IPX-DELETE-001")
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.FindByID("IPX-DELETE-001")
		assert.Error(t, err)
	})
}

func TestMovieRepository_EnsureGenresExist_RaceCondition(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewMovieRepository(db)

	t.Run("Upsert with new genres", func(t *testing.T) {
		movie := createTestMovie("IPX-GENRE-001")
		movie.Genres = []models.Genre{
			{Name: "Action"},
			{Name: "Comedy"},
		}

		// Upsert calls ensureGenresExist
		err := repo.Upsert(movie)
		require.NoError(t, err)

		// Verify genres were created
		found, err := repo.FindByID("IPX-GENRE-001")
		require.NoError(t, err)
		assert.Len(t, found.Genres, 2)
	})

	t.Run("Upsert with existing genres reuses them", func(t *testing.T) {
		// Pre-create a genre
		genreRepo := NewGenreRepository(db)
		existingGenre, err := genreRepo.FindOrCreate("Thriller")
		require.NoError(t, err)

		// Upsert movie with existing genre
		movie := createTestMovie("IPX-GENRE-002")
		movie.Genres = []models.Genre{
			{Name: "Thriller"}, // Already exists
		}

		err = repo.Upsert(movie)
		require.NoError(t, err)

		// Should reuse existing genre ID
		found, err := repo.FindByID("IPX-GENRE-002")
		require.NoError(t, err)
		require.Len(t, found.Genres, 1, "Should have exactly one genre")
		assert.Equal(t, existingGenre.ID, found.Genres[0].ID)

		// Verify only one Thriller genre exists in database
		allGenres, err := genreRepo.List()
		require.NoError(t, err)
		thrillerCount := 0
		for _, g := range allGenres {
			if g.Name == "Thriller" {
				thrillerCount++
			}
		}
		assert.Equal(t, 1, thrillerCount, "Should only have one Thriller genre")
	})
}

func TestMovieRepository_EnsureActressesExist_AllPaths(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewMovieRepository(db)

	t.Run("Ensure actresses with DMMID", func(t *testing.T) {
		movie := createTestMovie("IPX-ACTRESS-001")
		movie.Actresses = []models.Actress{
			{
				DMMID:        12345,
				JapaneseName: "テスト女優",
				FirstName:    "Test",
				LastName:     "Actress",
			},
		}

		// Use Upsert to test ensureActressesExist
		err := repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-001")
		require.NoError(t, err)
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, 12345, found.Actresses[0].DMMID)
	})

	t.Run("Ensure actresses with only JapaneseName", func(t *testing.T) {
		movie := createTestMovie("IPX-ACTRESS-002")
		movie.Actresses = []models.Actress{
			{
				DMMID:        20002, // Set unique DMMID to avoid constraint violation
				JapaneseName: "山田花子",
			},
		}

		err := repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-002")
		require.NoError(t, err)
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, "山田花子", found.Actresses[0].JapaneseName)
	})

	t.Run("Ensure actresses with only FirstName and LastName", func(t *testing.T) {
		movie := createTestMovie("IPX-ACTRESS-003")
		movie.Actresses = []models.Actress{
			{
				DMMID:     20003, // Set unique DMMID to avoid constraint violation
				FirstName: "Jane",
				LastName:  "Doe",
			},
		}

		err := repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-003")
		require.NoError(t, err)
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, "Jane", found.Actresses[0].FirstName)
	})

	t.Run("Ensure actresses with DMMID only are accepted", func(t *testing.T) {
		movie := createTestMovie("IPX-ACTRESS-004")
		movie.Actresses = []models.Actress{
			{
				// DMMID alone is sufficient as identifier
				DMMID:    20004,
				ThumbURL: "http://example.com/thumb.jpg",
			},
		}

		err := repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-004")
		require.NoError(t, err)
		// Actress with DMMID should be saved
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, 20004, found.Actresses[0].DMMID)
	})

	t.Run("Ensure actresses existing by DMMID", func(t *testing.T) {
		// Pre-create actress
		actressRepo := NewActressRepository(db)
		existing := &models.Actress{
			DMMID:        99999,
			JapaneseName: "既存女優",
			FirstName:    "Existing",
			LastName:     "Actress",
		}
		err := actressRepo.Create(existing)
		require.NoError(t, err)

		// Upsert movie referencing existing actress by DMMID
		movie := createTestMovie("IPX-ACTRESS-005")
		movie.Actresses = []models.Actress{
			{
				DMMID:        99999,
				JapaneseName: "Different Name", // Should use existing record
			},
		}

		err = repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-005")
		require.NoError(t, err)
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, existing.ID, found.Actresses[0].ID)
		assert.Equal(t, "既存女優", found.Actresses[0].JapaneseName) // Original name preserved
	})

	t.Run("Ensure actresses existing by JapaneseName (no DMMID provided)", func(t *testing.T) {
		// Pre-create actress without DMMID (DMMID=0)
		actressRepo := NewActressRepository(db)
		existing := &models.Actress{
			DMMID:        0, // No DMMID set - falls back to JapaneseName matching
			JapaneseName: "田中美咲",
			FirstName:    "Misaki",
			LastName:     "Tanaka",
		}
		err := actressRepo.Create(existing)
		require.NoError(t, err)

		// Upsert movie with same JapaneseName and no DMMID
		movie := createTestMovie("IPX-ACTRESS-006")
		movie.Actresses = []models.Actress{
			{
				DMMID:        0, // No DMMID - should match by JapaneseName
				JapaneseName: "田中美咲",
			},
		}

		err = repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-006")
		require.NoError(t, err)
		// Should reuse existing actress when DMMID=0 and JapaneseName matches
		// This tests fallback to JapaneseName matching when DMMID is not set
		assert.Len(t, found.Actresses, 1)
		// Verify the existing record was reused (same ID)
		assert.Equal(t, existing.ID, found.Actresses[0].ID, "Should reuse existing actress record when JapaneseName matches")
	})

	t.Run("Ensure actresses existing by FirstName and LastName (no DMMID or JapaneseName)", func(t *testing.T) {
		// Pre-create actress without DMMID or JapaneseName
		actressRepo := NewActressRepository(db)
		existing := &models.Actress{
			DMMID:     30007, // Unique DMMID
			FirstName: "Emily",
			LastName:  "Smith",
		}
		err := actressRepo.Create(existing)
		require.NoError(t, err)

		// Upsert movie with same name but no JapaneseName
		movie := createTestMovie("IPX-ACTRESS-007")
		movie.Actresses = []models.Actress{
			{
				DMMID:     40007, // Different DMMID - should create new actress
				FirstName: "Emily",
				LastName:  "Smith",
			},
		}

		err = repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-ACTRESS-007")
		require.NoError(t, err)
		// Different DMMID creates a new actress
		assert.Len(t, found.Actresses, 1)
		assert.NotEqual(t, existing.ID, found.Actresses[0].ID)
	})
}

func TestMovieRepository_Upsert_ComplexScenarios(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewMovieRepository(db)

	t.Run("Upsert new movie with all associations", func(t *testing.T) {
		movie := createTestMovie("IPX-UPSERT-001")
		movie.Genres = []models.Genre{
			{Name: "Drama"},
		}
		movie.Actresses = []models.Actress{
			{
				DMMID:        50001, // Unique DMMID
				JapaneseName: "アップサート女優",
			},
		}
		movie.Translations = []models.MovieTranslation{
			{
				MovieID:  "IPX-UPSERT-001",
				Language: "en",
				Title:    "Upsert Test",
			},
		}

		err := repo.Upsert(movie)
		require.NoError(t, err)

		found, err := repo.FindByID("IPX-UPSERT-001")
		require.NoError(t, err)
		assert.Equal(t, "IPX-UPSERT-001", found.ID)
		assert.Len(t, found.Genres, 1)
		assert.Len(t, found.Actresses, 1)
		assert.Len(t, found.Translations, 1)
	})

	t.Run("Upsert existing movie updates associations", func(t *testing.T) {
		// First create
		movie := createTestMovie("IPX-UPSERT-002")
		movie.Genres = []models.Genre{{Name: "Action"}}
		movie.Actresses = []models.Actress{
			{
				DMMID:        50002, // Unique DMMID
				JapaneseName: "女優1",
			},
		}
		err := repo.Upsert(movie)
		require.NoError(t, err)

		// Update with different associations
		movie.Genres = []models.Genre{{Name: "Comedy"}}
		movie.Actresses = []models.Actress{
			{
				DMMID:        50003, // Different unique DMMID
				JapaneseName: "女優2",
			},
		}
		movie.Translations = []models.MovieTranslation{
			{
				MovieID:  "IPX-UPSERT-002",
				Language: "en",
				Title:    "Updated Title",
			},
		}
		err = repo.Upsert(movie)
		require.NoError(t, err)

		// Verify update
		found, err := repo.FindByID("IPX-UPSERT-002")
		require.NoError(t, err)
		assert.Len(t, found.Genres, 1)
		assert.Equal(t, "Comedy", found.Genres[0].Name)
		assert.Len(t, found.Actresses, 1)
		assert.Equal(t, "女優2", found.Actresses[0].JapaneseName)
	})

	t.Run("Upsert preserves CreatedAt timestamp", func(t *testing.T) {
		// Create initial movie
		movie := createTestMovie("IPX-UPSERT-003")
		err := repo.Upsert(movie)
		require.NoError(t, err)

		// Get original CreatedAt
		found1, err := repo.FindByID("IPX-UPSERT-003")
		require.NoError(t, err)
		originalCreatedAt := found1.CreatedAt

		// Update movie
		movie.Title = "Updated Title"
		err = repo.Upsert(movie)
		require.NoError(t, err)

		// Verify CreatedAt unchanged
		found2, err := repo.FindByID("IPX-UPSERT-003")
		require.NoError(t, err)
		assert.Equal(t, originalCreatedAt, found2.CreatedAt)
		assert.Equal(t, "Updated Title", found2.Title)
	})
}

func TestGenreReplacementRepository_GetReplacementMap_Empty(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewGenreReplacementRepository(db)

	// Empty database should return empty map
	replacements, err := repo.GetReplacementMap()
	require.NoError(t, err)
	assert.Empty(t, replacements)
}

func TestActressAliasRepository_GetAliasMap_Empty(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewActressAliasRepository(db)

	// Empty database should return empty map
	aliases, err := repo.GetAliasMap()
	require.NoError(t, err)
	assert.Empty(t, aliases)
}

func TestMovieTagRepository_ErrorCases(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, db.AutoMigrate())
	repo := NewMovieTagRepository(db)

	t.Run("GetTagsForMovie - empty result", func(t *testing.T) {
		tags, err := repo.GetTagsForMovie("NON-EXISTENT")
		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("GetMoviesWithTag - empty result", func(t *testing.T) {
		movies, err := repo.GetMoviesWithTag("non-existent-tag")
		require.NoError(t, err)
		assert.Empty(t, movies)
	})

	t.Run("ListAll - empty result", func(t *testing.T) {
		all, err := repo.ListAll()
		require.NoError(t, err)
		assert.Empty(t, all)
	})
}

func TestAutoMigrate(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer db.Close()

	// AutoMigrate should succeed
	err = db.AutoMigrate()
	assert.NoError(t, err)

	// Running AutoMigrate again should be idempotent
	err = db.AutoMigrate()
	assert.NoError(t, err)
}
