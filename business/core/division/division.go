package division

import (
	"context"
	"errors"
	"github.com/nhaancs/bhms/foundation/logger"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound = errors.New("division not found")
)

// =============================================================================

type Storer interface {
	QueryByID(ctx context.Context, divisionID int) (Divison, error)
	QueryByParentID(ctx context.Context, parentID int) ([]Divison, error)
	QueryLevel1s(ctx context.Context) ([]Divison, error)
}

type Core struct {
	store Storer
	log   *logger.Logger
}

func NewCore(log *logger.Logger, store Storer) *Core {
	return &Core{
		store: store,
		log:   log,
	}
}