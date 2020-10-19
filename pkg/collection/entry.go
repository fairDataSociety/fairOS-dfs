package collection

type Manifest struct {
	Name         string  `json:"name"`
	CreationTime int64   `json:"creation_time"`
	Entries      []Entry `json:"entries,omitempty"`
}

type Entry struct {
	Name  string `json:"name"`
	EType string `json:"type"`
	Ref   []byte `json:"ref,omitempty"`
}

func NewManifest(name string, time int64) *Manifest {
	return &Manifest{
		Name:         name,
		CreationTime: time,
	}
}

func (m *Manifest) addEntry(name, eType string, ref []byte) {
	entry := Entry{
		Name:  name,
		EType: eType,
		Ref:   ref,
	}
	m.Entries = append(m.Entries, entry)
}
