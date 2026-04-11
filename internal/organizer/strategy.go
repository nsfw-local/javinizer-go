package organizer

import (
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
)

type OperationStrategy interface {
	Plan(match matcher.MatchResult, movie *models.Movie, destDir string, forceUpdate bool) (*OrganizePlan, error)
	Execute(plan *OrganizePlan) (*OrganizeResult, error)
}
