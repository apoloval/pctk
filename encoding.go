package pctk

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	// ResourceFormatVersion
	ResourceFormatVersion uint16 = 0x0001
)

// BinaryEncode encodes objects to a writer using the binary format. If the object implements the
// BinaryEncoder interface, it will be used to encode the object. If the object is a string, it will
// be encoded with its bytes prefixed with a word indicating its size. Otherwise, the object will be
// encoded using the binary package assuming it has a fixed size.
//
// Use this function to encode only values that implement BinaryEncoder, are strings, or are
// fixed-size primitive types. Passing a custom struct might work, but it will fail if it has
// variable-size fields.
func BinaryEncode(w io.Writer, o ...any) (n int, err error) {
	for _, obj := range o {
		nn, err := binaryEncode(w, obj)
		n += nn
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func binaryEncode(w io.Writer, obj any) (n int, err error) {
	if m, ok := obj.(BinaryEncoder); ok {
		return m.BinaryEncode(w)
	}
	if str, ok := obj.(string); ok {
		return BinaryEncode(w, uint16(len(str)), []byte(str))
	}
	err = binary.Write(w, binary.LittleEndian, obj)
	n = int(binary.Size(obj))
	return
}

// BinaryDecode decodes objects from a reader using the binary format. If the object implements the
// BinaryDecoder interface, it will be used to decode the object. If the object is a string, it will
// be decoded from a word indicating its size followed by the string bytes. Otherwise, the object
// will be decoded using the binary package assuming it has a fixed size.
func BinaryDecode(r io.Reader, o ...any) error {
	for _, obj := range o {
		if dec, ok := obj.(BinaryDecoder); ok {
			if err := dec.BinaryDecode(r); err != nil {
				return err
			}
			continue
		}
		if str, ok := obj.(*string); ok {
			var size uint16
			if err := BinaryDecode(r, &size); err != nil {
				return err
			}
			buf := make([]byte, size)
			if _, err := r.Read(buf); err != nil {
				return err
			}
			*str = string(buf)
			continue
		}
		if err := binary.Read(r, binary.LittleEndian, obj); err != nil {
			return err
		}
	}
	return nil
}

// BinaryEncoder is a value that can encode itself to a binary format.
type BinaryEncoder interface {
	BinaryEncode(w io.Writer) (int, error)
}

// BinaryDecoder is a value that can decode itself from a binary format.
type BinaryDecoder interface {
	BinaryDecode(r io.Reader) error
}

// ResourceCompression is the type of compression used for resources while encoding.
type ResourceCompression byte

const (
	// CompressionNone is the no compression.
	CompressionNone ResourceCompression = iota

	// CompressionGzip is the gzip compression.
	CompressionGzip
)

// ResourceEncoder is a value that can encode resources to a writer.
type ResourceEncoder struct {
	index io.Writer
	data  io.Writer

	next int
}

// NewResourceEncoder creates a new resource encoder that writes the index and data to the given
// writers.
func NewResourceEncoder(index, data io.Writer) (*ResourceEncoder, error) {
	enc := &ResourceEncoder{
		index: index,
		data:  data,
	}
	err := enc.encodeHeaders()
	return enc, err
}

// DataBytesWritten returns the number of bytes written to the data writer.
func (e *ResourceEncoder) DataBytesWritten() int {
	return e.next
}

// EncodeCostume encodes a costume using the resource encoder.
func (e *ResourceEncoder) EncodeCostume(id ResourceID, c *Costume, comp ResourceCompression) error {
	return e.encodeResource(id, c, resourceHeader{
		Type:        resourceTypeCostume,
		Compression: comp,
	})
}

// EncodeImage encodes an image using the resource encoder.
func (e *ResourceEncoder) EncodeImage(id ResourceID, i *Image, comp ResourceCompression) error {
	return e.encodeResource(id, i, resourceHeader{
		Type:        resourceTypeImage,
		Compression: comp,
	})
}

// EncodeMusic encodes a music using the resource encoder.
func (e *ResourceEncoder) EncodeMusic(id ResourceID, m *MusicTrack, comp ResourceCompression) error {
	return e.encodeResource(id, m, resourceHeader{
		Type:        resourceTypeMusic,
		Compression: comp,
	})
}

// EncodeScript encodes a script using the resource encoder.
func (e *ResourceEncoder) EncodeScript(id ResourceID, s *Script, comp ResourceCompression) error {
	return e.encodeResource(id, s, resourceHeader{
		Type:        resourceTypeScript,
		Compression: comp,
	})
}

// EncodeSound encodes a sound using the resource encoder.
func (e *ResourceEncoder) EncodeSound(id ResourceID, s *SoundTrack, comp ResourceCompression) error {
	return e.encodeResource(id, s, resourceHeader{
		Type:        resourceTypeSound,
		Compression: comp,
	})
}

// EncodeSpriteSheet encodes a sprite sheet using the resource encoder.
func (e *ResourceEncoder) EncodeSpriteSheet(
	id ResourceID,
	s *SpriteSheet,
	comp ResourceCompression,
) error {
	return e.encodeResource(id, s, resourceHeader{
		Type:        resourceTypeSpriteSheet,
		Compression: comp,
	})
}

func (e *ResourceEncoder) encodeResource(id ResourceID, res BinaryEncoder, h resourceHeader) error {
	var n int
	var err error

	if n, err = h.BinaryEncode(e.data); err != nil {
		return err
	}

	switch h.Compression {
	case CompressionNone:
		nn, err := res.BinaryEncode(e.data)
		if err != nil {
			return err
		}
		n += nn
	case CompressionGzip:
		var buf bytes.Buffer
		zipper := gzip.NewWriter(&buf)
		if _, err = res.BinaryEncode(zipper); err != nil {
			return err
		}
		if err = zipper.Close(); err != nil {
			return err
		}
		nn, err := e.data.Write(buf.Bytes())
		if err != nil {
			return err
		}
		n += nn
	}

	if err := e.encodeIndexEntry(id, e.next, n); err != nil {
		return err
	}
	e.next += n
	return nil
}

func (e *ResourceEncoder) encodeHeaders() error {
	if err := e.encodeIndexHeader(); err != nil {
		return err
	}
	if err := e.encodeDataHeader(); err != nil {
		return err
	}
	return nil
}

func (e *ResourceEncoder) encodeIndexHeader() error {
	_, err := BinaryEncode(e.index, resourceFileHeader{
		Magic:   resourceIndexMagic,
		Version: ResourceFormatVersion,
	})
	return err
}

func (e *ResourceEncoder) encodeDataHeader() error {
	n, err := BinaryEncode(e.data, resourceFileHeader{
		Magic:   resourceDataMagic,
		Version: ResourceFormatVersion,
	})
	e.next += n
	return err
}

func (e *ResourceEncoder) encodeIndexEntry(id ResourceID, offset, size int) error {
	_, err := BinaryEncode(e.index, indexEntry{id, uint32(offset), uint32(size)})
	return err
}

// ResourceFileLoader is a value that can load resources from files.
type ResourceFileLoader struct {
	path    string
	indexes map[ResourcePackage]index
}

// NewResourceFileLoader creates a new resource file loader that loads resources from the
// filesystem.
func NewResourceFileLoader(path string) *ResourceFileLoader {
	return &ResourceFileLoader{
		path:    path,
		indexes: make(map[ResourcePackage]index),
	}
}

func (l *ResourceFileLoader) LoadCostume(ref ResourceRef) *Costume {
	c := new(Costume)
	l.decodeResource(ref, resourceTypeCostume, c)
	return c
}

func (l *ResourceFileLoader) LoadImage(ref ResourceRef) *Image {
	img := new(Image)
	l.decodeResource(ref, resourceTypeImage, img)
	return img
}

func (l *ResourceFileLoader) LoadMusic(ref ResourceRef) *MusicTrack {
	m := new(MusicTrack)
	l.decodeResource(ref, resourceTypeMusic, m)
	return m
}

func (l *ResourceFileLoader) LoadScript(ref ResourceRef) *Script {
	script := new(Script)
	script.ref = ref
	l.decodeResource(ref, resourceTypeScript, script)
	return script
}

func (l *ResourceFileLoader) LoadSound(ref ResourceRef) *SoundTrack {
	sound := new(SoundTrack)
	l.decodeResource(ref, resourceTypeSound, sound)
	return sound
}

func (l *ResourceFileLoader) LoadSpriteSheet(ref ResourceRef) *SpriteSheet {
	ss := new(SpriteSheet)
	l.decodeResource(ref, resourceTypeSpriteSheet, ss)
	return ss
}

func (l *ResourceFileLoader) decodeResource(ref ResourceRef, t resourceType, res BinaryDecoder) {
	data := bytes.NewReader(l.getResource(ref, t))
	if err := BinaryDecode(data, res); err != nil {
		log.Fatalf("error decoding resource: %v", err)
	}
}

func (l *ResourceFileLoader) getResource(ref ResourceRef, t resourceType) []byte {
	entry := l.getIndexEntry(ref)
	data := make([]byte, int(entry.Size)-resourceHeaderSize)

	file, err := os.Open(filepath.Join(l.path, ref.Package().String()+".dat"))
	if err != nil {
		log.Fatalf("error opening data file: %v", err)
	}
	defer file.Close()

	if _, err := file.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		log.Fatalf("error seeking data file: %v", err)
	}

	var h resourceHeader
	if err := BinaryDecode(file, &h); err != nil {
		log.Fatalf("error decoding resource header: %v", err)
	}
	if h.Type != t {
		log.Fatalf("wrong resource type: %v != %v", h.Type, t)
	}

	if n, err := file.Read(data); err != nil {
		log.Fatalf("error reading data file: %v", err)
	} else if n != len(data) {
		log.Fatalf("short read: %d != %d", n, len(data))
	}

	switch h.Compression {
	case CompressionNone:
		return data
	case CompressionGzip:
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			log.Fatalf("error creating gzip reader: %v", err)
		}
		defer r.Close()

		data, err := io.ReadAll(r)
		if err != nil {
			log.Fatalf("error reading gzip data: %v", err)
		}
		return data
	default:
		log.Fatalf("unsupported compression: %v", h.Compression)
		return nil
	}
}

func (l *ResourceFileLoader) getIndexEntry(ref ResourceRef) indexEntry {
	idx, ok := l.indexes[ref.Package()]
	if !ok {
		idx = l.loadIndex(ref)
		l.indexes[ref.Package()] = idx
	}
	entry, ok := idx[ref.ID()]
	if !ok {
		log.Fatalf("resource not found: %s", ref)
	}
	return entry
}

func (l *ResourceFileLoader) loadIndex(ref ResourceRef) index {
	idxPath := filepath.Join(l.path, ref.Package().String()+".idx")
	idxFile, err := os.Open(idxPath)
	if err != nil {
		log.Fatalf("error opening index file for ref %s: %v", ref, err)
	}
	defer idxFile.Close()

	var h resourceFileHeader
	if err := BinaryDecode(idxFile, &h); err != nil {
		log.Fatalf("error decoding index header: %v", err)
	}
	if !bytes.Equal(h.Magic[:], resourceIndexMagic[:]) {
		log.Fatalf("wrong magic number in index file %s: %v", idxPath, h.Magic)
	}

	idx := make(index)
	for {
		var entry indexEntry
		if err := BinaryDecode(idxFile, &entry); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("error decoding index entry: %v", err)
		}
		idx[entry.ID] = entry
	}
	return idx
}

var (
	resourceIndexMagic = [8]byte{'P', 'C', 'T', 'K', ':', 'I', 'D', 'X'}
	resourceDataMagic  = [8]byte{'P', 'C', 'T', 'K', ':', 'D', 'A', 'T'}
)

type resourceFileHeader struct {
	Magic   [8]byte
	Version uint16
}

func (h resourceFileHeader) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, h.Magic, h.Version)
}

func (h *resourceFileHeader) BinaryDecode(r io.Reader) error {
	return BinaryDecode(r, &h.Magic, &h.Version)
}

type index map[ResourceID]indexEntry

type indexEntry struct {
	ID     ResourceID
	Offset uint32
	Size   uint32
}

func (e indexEntry) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, e.ID, e.Offset, e.Size)
}

func (e *indexEntry) BinaryDecode(r io.Reader) error {
	return BinaryDecode(r, &e.ID, &e.Offset, &e.Size)
}

const resourceHeaderSize = 16

type resourceHeader struct {
	Type        resourceType
	Compression ResourceCompression
}

func (h resourceHeader) BinaryEncode(w io.Writer) (int, error) {
	return BinaryEncode(w, h.Type, h.Compression, [14]byte{})
}

func (h *resourceHeader) BinaryDecode(r io.Reader) error {
	var unused [14]byte
	return BinaryDecode(r, &h.Type, &h.Compression, &unused)
}

type resourceType byte

const (
	resourceTypeUndefined resourceType = iota
	resourceTypeCostume
	resourceTypeImage
	resourceTypeMusic
	resourceTypeScript
	resourceTypeSound
	resourceTypeSpriteSheet
)
