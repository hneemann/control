package data

import (
	"encoding/json"
	"io"
	"sort"
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
	return s[i].Name < s[j].Name
}

func (s Scripts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type UserData struct {
	Scripts Scripts
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

func Load(r io.Reader) (*UserData, error) {
	var data UserData
	err := json.NewDecoder(r).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
