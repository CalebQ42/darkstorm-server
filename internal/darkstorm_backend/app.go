package darkstorm

type App interface {
	//TODO
}

type CrashApp interface {
	App
	AddCrash(CrashReport)
	// TODO
}
