package darkstorm

type Key struct {
	Perm  map[string]bool
	ID    string
	AppID string
	Death int
}

func (k Key) GetID() string {
	return k.ID
}
