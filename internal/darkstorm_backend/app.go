package darkstorm

type App interface {
	LogTable() Table[Log]
	CrashTable() CrashTable
}
