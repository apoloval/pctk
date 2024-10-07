package pack

import (
	"path/filepath"

	"github.com/apoloval/pctk"
	"gopkg.in/yaml.v3"
)

// MusicData is the data for a music resource.
type MusicData struct {
	Resource *pctk.MusicTrack

	workingDir string
}

// NewMusicData creates a new music data associated with a working directory.
func NewMusicData(workingDir string) *MusicData {
	return &MusicData{workingDir: workingDir}
}

func (d *MusicData) UnmarshalYAML(n *yaml.Node) error {
	var data struct {
		Source string
	}
	if err := n.Decode(&data); err != nil {
		return err
	}

	d.Resource = pctk.LoadMusicFromFile(filepath.Join(d.workingDir, data.Source))
	return nil
}
