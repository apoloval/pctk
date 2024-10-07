package pctk

import (
	"io"
	"log"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Sound is a resource that represents a sound.
type Sound struct {
	ref   ResourceRef
	track *SoundTrack
}

// LoadSound loads the sound from the given resource reference.
func NewSound(ref ResourceRef) *Sound {
	return &Sound{ref: ref}
}

// Play plays the sound.
func (s *Sound) Play(app *App) {
	if s.track == nil {
		track := app.res.LoadSound(s.ref)
		if track == nil {
			log.Fatalf("sound not found: %s", s.ref)
		}
		s.track = track
	}
	rl.PlaySound(s.track.raw)
}

// Stop stops the sound.
func (s *Sound) Stop(app *App) {
	rl.StopSound(s.track.raw)
}

// SoundTrack source type
type SoundTrack struct {
	data   []byte
	raw    rl.Sound
	format [4]byte
}

// LoadSoundFromFile - Load sound stream from a file path
func LoadSoundFromFile(path string) *SoundTrack {
	var err error
	sound := new(SoundTrack)

	sound.data, err = os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read sound file: %v", err)
	}
	wav := rl.LoadWaveFromMemory(filepath.Ext(path), sound.data, int32(len(sound.data)))
	sound.raw = rl.LoadSoundFromWave(wav)
	copy(sound.format[:], filepath.Ext(path))
	return sound
}

// BinaryEncode encodes the sound data to a binary stream. The format is:
//   - [4]byte: data format
//   - uint32: data length
//   - []byte: data
func (s *SoundTrack) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, s.format[:], uint32(len(s.data)), s.data)
}

// BinaryDecode decodes the sound data from a binary stream. See Sound.BinaryEncode for the format.
func (s *SoundTrack) BinaryDecode(r io.Reader) error {
	var format [4]byte
	var length uint32
	if err := BinaryDecode(r, &format, &length); err != nil {
		return err
	}

	data := make([]byte, length)
	if err := BinaryDecode(r, &data); err != nil {
		return err
	}

	s.format = format
	s.data = data
	wav := rl.LoadWaveFromMemory(string(format[:]), data, int32(length))
	s.raw = rl.LoadSoundFromWave(wav)
	return nil
}
