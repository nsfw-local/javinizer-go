package batch

import (
	"fmt"
	"path/filepath"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/downloader"
	"github.com/javinizer/javinizer-go/internal/history"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/nfo"
	"github.com/javinizer/javinizer-go/internal/organizer"
	"github.com/javinizer/javinizer-go/internal/scanner"
	"github.com/javinizer/javinizer-go/internal/types"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/spf13/afero"
)

// processOrganizeJob processes file organization for a completed scrape job
func processOrganizeJob(job *worker.BatchJob, jobQueue *worker.JobQueue, destination string, copyOnly bool, linkModeRaw string, db *database.DB, cfg *config.Config, registry *models.ScraperRegistry) {
	// Determine effective operation mode and apply overrides
	outputConfig := cfg.Output
	if job.OperationModeOverride != "" {
		parsed, err := types.ParseOperationMode(job.OperationModeOverride)
		if err != nil {
			logging.Warnf("Invalid operation mode override %q: %v, using config default", job.OperationModeOverride, err)
		} else {
			outputConfig.OperationMode = parsed
			outputConfig.MoveToFolder = parsed == types.OperationModeOrganize
			outputConfig.RenameFolderInPlace = parsed == types.OperationModeInPlace
		}
	} else {
		// Apply legacy boolean overrides only when no explicit mode override
		if job.MoveToFolderOverride != nil {
			outputConfig.MoveToFolder = *job.MoveToFolderOverride
		}
		if job.RenameFolderInPlaceOverride != nil {
			outputConfig.RenameFolderInPlace = *job.RenameFolderInPlaceOverride
		}
		effectiveMode := outputConfig.GetOperationMode()
		outputConfig.OperationMode = effectiveMode
		outputConfig.MoveToFolder = effectiveMode == types.OperationModeOrganize
		outputConfig.RenameFolderInPlace = effectiveMode == types.OperationModeInPlace
	}
	effectiveMode := outputConfig.GetOperationMode()

	// Initialize organizer, downloader, NFO generator, and history logger
	org := organizer.NewOrganizer(afero.NewOsFs(), &outputConfig)
	fileMatcher, err := matcher.NewMatcher(&cfg.Matching)
	if err != nil {
		logging.Warnf("Failed to create matcher: %v (in-place rename disabled for this job)", err)
	} else {
		org.SetMatcher(fileMatcher)
	}

	// Select strategy based on effective operation mode
	var strategy organizer.OperationStrategy
	fs := afero.NewOsFs()
	switch effectiveMode {
	case types.OperationModeOrganize:
		strategy = organizer.NewOrganizeStrategy(fs, &outputConfig)
	case types.OperationModeInPlace:
		if fileMatcher != nil {
			strategy = organizer.NewInPlaceStrategy(fs, &outputConfig, fileMatcher)
		} else {
			logging.Warnf("No matcher available for in-place mode, falling back to organize")
			strategy = organizer.NewOrganizeStrategy(fs, &outputConfig)
		}
	case types.OperationModeInPlaceNoRenameFolder:
		if fileMatcher != nil {
			strategy = organizer.NewInPlaceNoRenameFolderStrategy(fs, &outputConfig, fileMatcher)
		} else {
			logging.Warnf("No matcher available for in-place-norenamefolder mode, falling back to organize")
			strategy = organizer.NewOrganizeStrategy(fs, &outputConfig)
		}
	case types.OperationModeMetadataOnly:
		strategy = organizer.NewMetadataOnlyStrategy(fs, &outputConfig)
	case types.OperationModePreview:
		// Preview mode wraps organize strategy — shouldn't reach here in organize job
		// but handle gracefully by falling back to organize
		logging.Warnf("Preview mode reached in organize job, falling back to organize")
		strategy = organizer.NewOrganizeStrategy(fs, &outputConfig)
	default:
		strategy = organizer.NewOrganizeStrategy(fs, &outputConfig)
	}

	historyLogger := history.NewLogger(db)
	linkMode, err := organizer.ParseLinkMode(linkModeRaw)
	if err != nil {
		broadcastProgress(&ws.ProgressMessage{
			JobID:    job.ID,
			Status:   "error",
			Progress: 0,
			Message:  fmt.Sprintf("Invalid link mode: %v", err),
		})
		job.MarkFailed()
		if jobQueue != nil {
			jobQueue.PersistJob(job)
		}
		return
	}

	// Initialize HTTP client for downloader
	httpClient, err := downloader.NewHTTPClientForDownloaderWithRegistry(cfg, registry)
	if err != nil {
		broadcastProgress(&ws.ProgressMessage{
			JobID:    job.ID,
			Status:   "error",
			Progress: 0,
			Message:  fmt.Sprintf("Failed to create HTTP client: %v", err),
		})
		job.MarkFailed()
		if jobQueue != nil {
			jobQueue.PersistJob(job)
		}
		return
	}
	dl := downloader.NewDownloaderWithNFOConfig(httpClient, afero.NewOsFs(), &cfg.Output, "Javinizer/1.0", cfg.Metadata.NFO.ActressLanguageJA, cfg.Metadata.NFO.FirstNameOrder)
	nfoGen := nfo.NewGenerator(afero.NewOsFs(), nfo.ConfigFromAppConfig(&cfg.Metadata.NFO, &cfg.Output, &cfg.Metadata, db))

	// Broadcast organization started
	broadcastProgress(&ws.ProgressMessage{
		JobID:    job.ID,
		Status:   "organizing",
		Progress: 0,
		Message:  "Starting file organization",
	})

	status := job.GetStatus()
	organized := 0
	failed := 0

	for filePath, fileResult := range status.Results {
		// Skip files that failed during scraping
		if fileResult.Status != worker.JobStatusCompleted || fileResult.Data == nil {
			continue
		}

		// Skip files excluded by user
		if job.IsExcluded(filePath) {
			logging.Infof("Skipping excluded file: %s", filePath)
			continue
		}

		movie, ok := fileResult.Data.(*models.Movie)
		if !ok {
			logging.Errorf("Invalid movie data type for file: %s", filePath)
			failed++
			continue
		}

		// Create match result for organizer using multipart metadata from job results
		// This preserves IsMultiPart status for letter patterns (-A, -B) that would
		// be lost if we re-matched files individually
		ext := filepath.Ext(filePath)
		match := matcher.MatchResult{
			File: scanner.FileInfo{
				Path:      filePath,
				Name:      filepath.Base(filePath),
				Extension: ext,
				Dir:       filepath.Dir(filePath),
			},
			ID:          movie.ID,
			IsMultiPart: fileResult.IsMultiPart,
			PartNumber:  fileResult.PartNumber,
			PartSuffix:  fileResult.PartSuffix,
		}

		// Organize file using selected strategy
		var result *organizer.OrganizeResult
		var organizeErr error

		if effectiveMode == types.OperationModeOrganize {
			// Use existing Organizer for organize mode (trusted code path)
			result, organizeErr = org.OrganizeWithLinkMode(match, movie, destination, false, false, copyOnly, linkMode)
		} else {
			// Use strategy for non-organize modes
			plan, planErr := strategy.Plan(match, movie, destination, false)
			if planErr != nil {
				organizeErr = planErr
			} else {
				result, organizeErr = strategy.Execute(plan)
				if result == nil && organizeErr == nil {
					result = &organizer.OrganizeResult{}
				}
			}
		}

		if organizeErr != nil {
			logging.Errorf("Failed to organize %s: %v", filePath, organizeErr)
			failed++

			// Log failed organize operation
			if logErr := historyLogger.LogOrganize(movie.ID, filePath, "", false, organizeErr); logErr != nil {
				logging.Warnf("Failed to log history for %s: %v", filePath, logErr)
			}

			broadcastProgress(&ws.ProgressMessage{
				JobID:    job.ID,
				FilePath: filePath,
				Status:   "failed",
				Progress: float64(organized+failed) / float64(len(status.Results)) * 100,
				Error:    organizeErr.Error(),
			})
			continue
		}

		// Log successful organize operation (includes metadata-only and in-place modes)
		if result.Moved || result.ShouldGenerateMetadata {
			newPath := result.NewPath
			if newPath == "" {
				newPath = result.OriginalPath // For metadata-only, log original path
			}
			if logErr := historyLogger.LogOrganize(movie.ID, filePath, newPath, false, nil); logErr != nil {
				logging.Warnf("Failed to log history for %s: %v", filePath, logErr)
			}
		}

		postMoveIssueCount := 0

		// Surface subtitle move failures clearly in logs for support/debug workflows.
		for _, subtitle := range result.Subtitles {
			if subtitle.Error != nil {
				postMoveIssueCount++
				logging.Warnf("[post-move] mode=Organize movie=%s file=%s stage=subtitle_move src=%s dst=%s err=%v", movie.ID, filePath, subtitle.OriginalPath, subtitle.NewPath, subtitle.Error)
			}
		}

		// Copy temp cropped poster and download all media files
		if result.ShouldGenerateMetadata {
			// Create multipart info from match for template conditionals
			var multipart *downloader.MultipartInfo
			if match.IsMultiPart {
				multipart = &downloader.MultipartInfo{
					IsMultiPart: match.IsMultiPart,
					PartNumber:  match.PartNumber,
					PartSuffix:  match.PartSuffix,
				}
			}

			// Copy temp cropped poster BEFORE downloads (so downloader skips it)
			copyTempCroppedPoster(job, movie, result.FolderPath, cfg, "Organize", multipart)

			// Download all media files and log to history
			downloadMediaFilesWithHistory(dl, movie, result.FolderPath, cfg, historyLogger, multipart)
		}

		// Generate NFO file
		if result.ShouldGenerateMetadata && cfg.Metadata.NFO.Enabled {
			// Determine part suffix for multi-part files (only if per_file is enabled)
			partSuffix := ""
			if cfg.Metadata.NFO.PerFile && match.IsMultiPart {
				partSuffix = match.PartSuffix
			}

			// Pass the video file path for stream details extraction
			// Use NewPath (destination) after move/copy, fall back to OriginalPath
			videoFilePath := result.NewPath
			if videoFilePath == "" {
				videoFilePath = result.OriginalPath
			}
			nfoErr := nfoGen.Generate(movie, result.FolderPath, partSuffix, videoFilePath)
			if nfoErr != nil {
				postMoveIssueCount++
				logging.Warnf("[post-move] mode=Organize movie=%s file=%s stage=nfo_generate folder=%s video=%s part_suffix=%q err=%v", movie.ID, filePath, result.FolderPath, videoFilePath, partSuffix, nfoErr)
			}

			// Log NFO generation to history
			nfoPath := filepath.Join(result.FolderPath, movie.ID+".nfo")
			if logErr := historyLogger.LogNFO(movie.ID, nfoPath, nfoErr); logErr != nil {
				logging.Warnf("Failed to log NFO history for %s: %v", movie.ID, logErr)
			}
		} else if result.ShouldGenerateMetadata && !cfg.Metadata.NFO.Enabled {
			logging.Debugf("NFO generation disabled in config, skipping for %s", movie.ID)
		}

		if postMoveIssueCount > 0 {
			logging.Warnf("[post-move] mode=Organize movie=%s file=%s stage=summary issues=%d moved_path=%s folder=%s", movie.ID, filePath, postMoveIssueCount, result.NewPath, result.FolderPath)
		}

		organized++

		broadcastProgress(&ws.ProgressMessage{
			JobID:    job.ID,
			FilePath: filePath,
			Status:   "organized",
			Progress: float64(organized+failed) / float64(len(status.Results)) * 100,
			Message:  fmt.Sprintf("Organized %s", movie.ID),
		})
	}

	// Broadcast final completion
	broadcastProgress(&ws.ProgressMessage{
		JobID:    job.ID,
		Status:   "organization_completed",
		Progress: 100,
		Message:  fmt.Sprintf("Organized %d files, %d failed", organized, failed),
	})

	// Only transition to "Organized" state if ALL files organized successfully
	// If any files failed, keep job in "Completed" state to enable retry
	// If no files were processed (all skipped/excluded), stay in "Completed" for inspection
	// State machine: Pending → Running → Completed → Organized (only on full success with actual work)
	//                Pending → Running → Completed (stays here if failed > 0 or organized == 0)
	if failed == 0 && organized > 0 {
		job.MarkOrganized()
	} else {
		// Re-mark as completed to ensure job is in retryable state
		// (MarkStarted was called at the beginning of organization)
		job.MarkCompleted()
	}
	if jobQueue != nil {
		jobQueue.PersistJob(job)
	}
}
