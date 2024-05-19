package darkstorm

import "errors"

var (
	ErrIDNotFound = errors.New("id not found in table")
)

type IDStruct interface {
	GetID() string
}

type Table[T IDStruct] interface {
	Get(ID string) (data T, err error)
	Insert(data T) error
	Update(data T) error
	Remove(ID string)
}

type CrashTable interface {
	Table[CrashReport]
	InsertCrash(report IndividualCrash) error
}
