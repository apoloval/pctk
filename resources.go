package pctk

import (
	"fmt"
	"io"
	"strings"
)

// ResourcePackage is the package of a resource. This is used to group resources together.
type ResourcePackage string

// String returns the string representation of the resource package.
func (r ResourcePackage) String() string {
	return string(r)
}

// BinaryEncode encodes the resource package to a binary format.
func (r ResourcePackage) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, r.String())
}

// ResourceID is the identifier of a resource in a package.
type ResourceID string

// String returns the string representation of the resource ID.
func (i ResourceID) String() string {
	return string(i)
}

// BinaryEncode encodes the resource ID to a binary format.
func (i ResourceID) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, i.String())
}

// BinaryDecode decodes the resource ID from a binary format.
func (i *ResourceID) BinaryDecode(r io.Reader) error {
	var id string
	if err := BinaryDecode(r, &id); err != nil {
		return err
	}
	*i = ResourceID(id)
	return nil
}

// ResourceRef is a reference to a resource. This is typically used from resources to refer to other
// resources.
type ResourceRef struct {
	pkg ResourcePackage
	id  ResourceID
}

// ResourceRefNull is a null resource reference.
var ResourceRefNull = ResourceRef{}

// NewResourceRef creates a new resource reference.
func NewResourceRef(pkg ResourcePackage, id ResourceID) ResourceRef {
	return ResourceRef{pkg, id}
}

// ParseResourceRef parses a resource reference from a string. The string must be in the format
// "pkg:id".
func ParseResourceRef(s string) (ResourceRef, error) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 2:
		return ResourceRef{ResourcePackage(parts[0]), ResourceID(parts[1])}, nil
	default:
		return ResourceRef{}, fmt.Errorf("invalid resource reference: %s", s)
	}
}

// String returns the string representation of the resource reference.
func (r ResourceRef) String() string {
	if r.pkg == "" {
		return r.id.String()
	}
	return r.pkg.String() + ":" + r.id.String()
}

func (r ResourceRef) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, r.String())
}

// IsNull returns true if the reference is null.
func (r ResourceRef) IsNull() bool {
	return len(r.pkg) == 0 && len(r.id) == 0
}

// Package returns the part of the reference that indicates the resource package.
func (r ResourceRef) Package() ResourcePackage {
	return r.pkg
}

// ID returns the part of the reference that indicates the resource ID in its package.
func (r ResourceRef) ID() ResourceID {
	return r.id
}

// ResourceLoader is a value that can load game resources.
type ResourceLoader interface {
	// LoadCostume loads a costume from the given ref. It returns nil if the costume is not
	// found.
	LoadCostume(ref ResourceRef) *Costume

	// LoadImage loads an image from the given ref. It returns nil if the image is not found.
	LoadImage(ref ResourceRef) *Image

	// LoadMusic loads a music song from the given ref. It returns nil if the music is not
	// found.
	LoadMusic(ref ResourceRef) *MusicTrack

	// LoadScript loads a script from the given ref. It returns nil if the script is not found.
	LoadScript(ref ResourceRef) *Script

	// LoadSound loads a sound effect from he given ref. It returns nil if the sound is not
	// found.
	LoadSound(ref ResourceRef) *SoundTrack

	// LoadSpriteSheet loads a sprite sheet from the given ref. It returns nil if the sprite
	// sheet is not found.
	LoadSpriteSheet(ref ResourceRef) *SpriteSheet
}

// ResourceBundle is a bundle of resources that are loaded in memory. This can be used for
// testing purposes mainly.
type ResourceBundle struct {
	costumes map[ResourceRef]*Costume
	images   map[ResourceRef]*Image
	music    map[ResourceRef]*MusicTrack
	scripts  map[ResourceRef]*Script
	sounds   map[ResourceRef]*SoundTrack
	sprites  map[ResourceRef]*SpriteSheet
}

// NewResourceBundle creates a new resource bundle that can be used as resource loader.
func NewResourceBundle() *ResourceBundle {
	return &ResourceBundle{
		costumes: make(map[ResourceRef]*Costume),
		images:   make(map[ResourceRef]*Image),
		music:    make(map[ResourceRef]*MusicTrack),
		scripts:  make(map[ResourceRef]*Script),
		sounds:   make(map[ResourceRef]*SoundTrack),
		sprites:  make(map[ResourceRef]*SpriteSheet),
	}
}

// PutCostume adds a costume to the bundle.
func (c *ResourceBundle) PutCostume(ref ResourceRef, cos *Costume) {
	c.costumes[ref] = cos
}

// PutImage adds an image to the bundle.
func (c *ResourceBundle) PutImage(ref ResourceRef, img *Image) {
	c.images[ref] = img
}

// PutMusic adds a music song to the bundle.
func (c *ResourceBundle) PutMusic(ref ResourceRef, m *MusicTrack) {
	c.music[ref] = m
}

// PutScript adds a script to the bundle.
func (c *ResourceBundle) PutScript(ref ResourceRef, s *Script) {
	c.scripts[ref] = s
}

// PutSound adds a sound to the bundle.
func (c *ResourceBundle) PutSound(ref ResourceRef, s *SoundTrack) {
	c.sounds[ref] = s
}

// PutSpriteSheet adds a sprite sheet to the bundle.
func (c *ResourceBundle) PutSpriteSheet(ref ResourceRef, s *SpriteSheet) {
	c.sprites[ref] = s
}

// LoadCostume loads a costume from the given ref. It returns nil if the costume is not found.
func (c *ResourceBundle) LoadCostume(ref ResourceRef) *Costume {
	return c.costumes[ref]
}

// LoadImage loads an image from the given ref. It returns nil if the image is not found.
func (c *ResourceBundle) LoadImage(ref ResourceRef) *Image {
	return c.images[ref]
}

// LoadMusic loads a music song from the given ref. It returns nil if the music is not found.
func (c *ResourceBundle) LoadMusic(ref ResourceRef) *MusicTrack {
	return c.music[ref]
}

// LoadScript loads a script from the given ref. It returns nil if the script is not found.
func (c *ResourceBundle) LoadScript(ref ResourceRef) *Script {
	return c.scripts[ref]
}

// LoadSound loads a sound effect from he given ref. It returns nil if the sound is not found.
func (c *ResourceBundle) LoadSound(ref ResourceRef) *SoundTrack {
	return c.sounds[ref]
}

// LoadSpriteSheet loads a sprite sheet from the given ref. It returns nil if the sprite sheet is
// not found.
func (c *ResourceBundle) LoadSpriteSheet(ref ResourceRef) *SpriteSheet {
	return c.sprites[ref]
}
