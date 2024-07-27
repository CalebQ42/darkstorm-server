package backend

import "net/http"

// An application interface. Both LogTable and CrashTable are optional, if they return nil then requests will be forbidden.
type App interface {
	AppID() string
	CountTable() CountTable
	CrashTable() CrashTable
}

// Allows for an App to filter crashes before they get added to the DB, such as making sure the crash is from the correct version.
type CrashFilterApp interface {
	App
	AddCrash(IndividualCrash) bool
}

// Allows an app more flexibility by directly interfacing with the backend's mux
type ExtendedApp interface {
	Extension(*http.ServeMux)
}

type simpleApp struct {
	countTab CountTable
	crashTab CrashTable
	appID    string
}

func NewSimpleApp(appID string, countTable CountTable, crashTable CrashTable) App {
	return &simpleApp{
		appID:    appID,
		countTab: countTable,
		crashTab: crashTable,
	}
}

func (s *simpleApp) AppID() string {
	return s.appID
}
func (s *simpleApp) CountTable() CountTable {
	return s.countTab
}
func (s *simpleApp) CrashTable() CrashTable {
	return s.crashTab
}
