package backend

import "net/http"

// An application interface. Both LogTable and CrashTable are optional, if they return nil then requests will be forbidden.
type App interface {
	AppID() string
	CountTable() CountTable
	CrashTable() CrashTable
}

type ExtendedApp interface {
	// Extension is called for any calls to /{appID}/
	// Alternatively, use Backend.HandleFunc for more customizability
	Extension(http.ResponseWriter, *http.Request)
}
