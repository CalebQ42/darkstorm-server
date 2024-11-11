package backend

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("no matches found in table")
)

type IDStruct interface {
	GetID() string
}

type Table[T IDStruct] interface {
	Get(ctx context.Context, ID string) (data T, err error)
	Find(ctx context.Context, values map[string]any) ([]T, error)
	Insert(ctx context.Context, data T) error
	Remove(ctx context.Context, ID string) error
	FullUpdate(ctx context.Context, ID string, data T) error
	PartUpdate(ctx context.Context, ID string, update map[string]any) error
}

type CountTable interface {
	Table[CountLog]
	// Remove all Log items that have a CountLog.Date value less then the given value.
	RemoveOldLogs(ctx context.Context, date int) error
	// Get count. If platform is an empty string or "all", the full count should be given
	Count(ctx context.Context, platform string) (int, error)
}

type CrashTable interface {
	Table[CrashReport]
	// Move a crash type to archive. Crashes that match the archived crash will be automatically removed from the CrashTable.
	Archive(context.Context, ArchivedCrash) error
	IsArchived(context.Context, IndividualCrash) bool
	// Add the IndividualCrash report to the crash table. If a CrashReport exists that matches, then it gets added to CrashReport.Individual.
	// If an IndividualCrash exists that is a perfect match, Count is incremented instead of adding it to the array.
	InsertCrash(context.Context, IndividualCrash) error
}
