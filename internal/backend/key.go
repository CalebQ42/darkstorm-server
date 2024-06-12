package backend

type ApiKey struct {
	Perm  map[string]bool
	ID    string
	AppID string
	Death int64
}

func (k ApiKey) GetID() string {
	return k.ID
}
