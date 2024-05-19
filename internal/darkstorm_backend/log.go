package darkstorm

type Log struct {
	ID       string
	Platform string
	Date     int
}

func (l Log) GetID() string {
	return l.ID
}
