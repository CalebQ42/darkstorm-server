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

type CrashTable interface {
	Table[CrashReport]
	InsertCrash(ID string, report IndividualCrash) error
}
