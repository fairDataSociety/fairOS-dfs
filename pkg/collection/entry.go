package collection

type Manifest struct {
	Name         string   `json:"name"`
	CreationTime int64    `json:"creation_time"`
	Entries      []*Entry `json:"entries,omitempty"`
	dirtyFlag    bool
}

type Entry struct {
	Name     string `json:"name"`
	EType    string `json:"type"`
	Ref      []byte `json:"ref,omitempty"`
	manifest *Manifest
}

func NewManifest(name string, time int64) *Manifest {
	return &Manifest{
		Name:         name,
		CreationTime: time,
		dirtyFlag:    true,
	}
}
