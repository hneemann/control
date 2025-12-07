package data

import (
	"encoding/json"
	"github.com/hneemann/parser2/funcGen"
	"github.com/hneemann/parser2/value"
	"io"
	"sort"
	"strings"
	"time"
)

type Script struct {
	Name string
	Code string
	Time time.Time
}

func (s Script) Size() int {
	return len(s.Code)
}

type Scripts []Script

func (s Scripts) Len() int {
	return len(s)
}

func (s Scripts) Less(i, j int) bool {
	return strings.ToLower(s[i].Name) < strings.ToLower(s[j].Name)
}

func (s Scripts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type UserData struct {
	Scripts    Scripts
	lastScript string
	lastFunc   funcGen.Func[value.Value]
}

func (d *UserData) Save(w io.Writer) error {
	return json.NewEncoder(w).Encode(d)
}

func (d *UserData) Add(name string, src string) {
	// Check if the script already exists
	for i, script := range d.Scripts {
		if script.Name == name {
			// Update the existing script
			d.Scripts[i].Code = src
			d.Scripts[i].Time = time.Now()
			return
		}
	}
	d.Scripts = append(d.Scripts, Script{Name: name, Code: src, Time: time.Now()})
	sort.Sort(d.Scripts)
}

func (d *UserData) Get(name string) (string, bool) {
	for _, script := range d.Scripts {
		if script.Name == name {
			return script.Code, true
		}
	}
	return "", false
}

func (d *UserData) Delete(name string) bool {
	for i, script := range d.Scripts {
		if script.Name == name {
			d.Scripts = append(d.Scripts[:i], d.Scripts[i+1:]...)
			return true
		}
	}
	return false
}

func (d *UserData) GetLastFu(src string) funcGen.Func[value.Value] {
	if d.lastScript == src {
		return d.lastFunc
	}
	return nil
}

func (d *UserData) SetLastFu(src string, fu funcGen.Func[value.Value]) {
	d.lastScript = src
	d.lastFunc = fu
}

func Load(r io.Reader) (*UserData, error) {
	var data UserData
	err := json.NewDecoder(r).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
