package darkstorm

import "errors"

var (
	ErrNotFound = errors.New("no matches found in table")
)

type IDStruct interface {
	GetID() string
}

type Table[T IDStruct] interface {
	Get(ID string) (data T, err error)
	Find(values map[string]any) ([]T, error)
	Insert(data T) error
	Remove(ID string) error
	FullUpdate(ID string, data T) error
	PartUpdate(ID string, update map[string]any) error
}

type CountTable interface {
	Table[CountLog]
	// Remove all Log items that have a CountLog.Date value less then the given value.
	RemoveOldLogs(date int)
	// Get count. If platform is an empty string or "all", the full count should be given
	Count(platform string) int
}

type CrashTable interface {
	Table[CrashReport]
	// Move a crash type to archive. Crashes that match the archived crash will be automatically removed from the CrashTable.
	Archive(ArchivedCrash) error
	IsArchived(IndividualCrash) bool
	// Add the IndividualCrash report to the crash table. If a CrashReport exists that matches, then it gets added to CrashReport.Individual.
	// If an IndividualCrash exists that is a perfect match, Count is incremented instead of adding it to the array.
	InsertCrash(IndividualCrash) error
}
