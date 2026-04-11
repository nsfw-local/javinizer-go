package organizer

import (
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
)

type PreviewStrategy struct {
	delegate OperationStrategy
}

var _ OperationStrategy = (*PreviewStrategy)(nil)

func NewPreviewStrategy(delegate OperationStrategy) *PreviewStrategy {
	return &PreviewStrategy{
		delegate: delegate,
	}
}

func (s *PreviewStrategy) Plan(match matcher.MatchResult, movie *models.Movie, destDir string, forceUpdate bool) (*OrganizePlan, error) {
	return s.delegate.Plan(match, movie, destDir, forceUpdate)
}

func (s *PreviewStrategy) Execute(plan *OrganizePlan) (*OrganizeResult, error) {
	return &OrganizeResult{
		OriginalPath: plan.SourcePath,
		NewPath:      plan.TargetPath,
		FolderPath:   plan.TargetDir,
		FileName:     plan.TargetFile,
		Moved:        false,
		Subtitles:    nil,
		Error:        nil,
	}, nil
}
