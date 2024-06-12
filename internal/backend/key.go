package backend

type ApiKey struct {
	Perm  map[string]bool `json:"perm" bson:"perm"`
	ID    string          `json:"id" bson:"_id" valkey:",key"`
	AppID string          `json:"appID" bson:"appID"`
	Death int64           `json:"death" bson:"death"`
}

func (k ApiKey) GetID() string {
	return k.ID
}
