package darkstorm

type IndividualCrash struct {
	Platform string
	Error    string
	Stack    string
	Count    int
}

type CrashReport struct {
	ID         string
	Error      string
	FirstLine  string
	Individual []IndividualCrash
}
