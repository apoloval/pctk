package pctk

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Music is a entity that represents a music.
type Music struct {
	ref   ResourceRef
	track *MusicTrack
}

// LoadMusic loads the music from the given resource reference.
func NewMusic(ref ResourceRef) *Music {
	return &Music{ref: ref}
}

// Play plays the music.
func (m *Music) Play(app *App) {
	if m.track == nil {
		track := app.res.LoadMusic(m.ref)
		if track == nil {
			log.Fatalf("music not found: %s", m.ref)
		}
		m.track = track
	}

	rl.PlayMusicStream(m.track.raw)
}

// Stop stops the music.
func (m *Music) Stop(app *App) {
	rl.StopMusicStream(m.track.raw)
	rl.UnloadMusicStream(m.track.raw)
	m.track = nil
}

// MusicTrack is the data of a music entity.
type MusicTrack struct {
	data   []byte
	format [4]byte
	raw    rl.Music
}

// LoadMusicFromFile - Load music stream from a file path
func LoadMusicFromFile(path string) *MusicTrack {
	var err error
	music := new(MusicTrack)
	music.data, err = os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read music file: %v", err)
	}

	music.raw = rl.LoadMusicStreamFromMemory(filepath.Ext(path), music.data, int32(len(music.data)))
	copy(music.format[:], strings.ToUpper(filepath.Ext(path)))
	return music
}

// BinaryEncode encodes the music data to a binary stream. The format is:
//   - [4]byte: data format
//   - uint32: data length
//   - []byte: data
func (m *MusicTrack) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, m.format[:], uint32(len(m.data)), m.data)
}

// BinaryDecode decodes the music data from a binary stream. See Music.BinaryEncode for the format.
func (m *MusicTrack) BinaryDecode(r io.Reader) error {
	var format [4]byte
	var length uint32
	if err := BinaryDecode(r, &format, &length); err != nil {
		return err
	}

	data := make([]byte, length)
	if err := BinaryDecode(r, &data); err != nil {
		return err
	}

	m.format = format
	m.data = data
	m.raw = rl.LoadMusicStreamFromMemory(strings.ToLower(string(format[:])), data, int32(length))
	return nil
}

// PlayMusic plays the music.
func (a *App) PlayMusic(music *Music) {
	if a.music != nil {
		a.music.Stop(a)
	}
	a.music = music
	a.music.Play(a)
}

// StopMusic stops the music.
func (a *App) StopMusic() {
	if a.music != nil {
		a.music.Stop(a)
		a.music = nil
	}
}

func (a *App) updateMusic() {
	if a.music != nil {
		rl.UpdateMusicStream(a.music.track.raw)
	}
}
