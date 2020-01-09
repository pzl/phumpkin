package photos

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

type Size int

const (
	SizeXS     Size = 10
	SizeSmall  Size = 200
	SizeMedium Size = 800
	SizeLarge  Size = 1200
	SizeXL     Size = 2000
	SizeFull   Size = 999999
)

func (s Size) String() string {
	switch s {
	case SizeXS:
		return "x-small"
	case SizeSmall:
		return "small"
	case SizeMedium:
		return "medium"
	case SizeLarge:
		return "large"
	case SizeXL:
		return "x-large"
	case SizeFull:
		return "full"
	}
	return "medium"
}

func ParseSize(s string) Size {
	switch strings.ToLower(s) {
	case "x-small", "xsmall", "xs":
		return SizeXS
	case "small", "sm", "s":
		return SizeSmall
	case "medium", "med", "m":
		return SizeMedium
	case "large", "lg", "l":
		return SizeLarge
	case "x-large", "xlarge", "xl":
		return SizeXL
	case "full", "f":
		return SizeFull
	}
	return SizeMedium
}

func (s Size) Int() int {
	if s != SizeFull {
		return int(s)
	}
	return 0
}

// https://stackoverflow.com/questions/52161555/how-to-custom-marshal-map-keys-in-json
func (s Size) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

/* -- XMP struct --- */

type XMP struct {
	DerivedFromFile string        `json:"derived_from"`
	Rating          int           `json:"rating"`
	Location        *Location     `json:"loc,omitempty"`
	AutoPresets     bool          `json:"auto_presets_applied"`
	XMPVersion      int           `json:"xmp_version"`
	ColorLabels     []string      `json:"color_labels,omitempty"`
	Creator         string        `json:"creator,omitempty"`
	History         []DTOperation `json:"history,omitempty"`
	Rights          string        `json:"rights"`
	Tags            []string      `json:"tags,omitempty"`
	Title           string        `json:"title,omitempty"`
}

type Location struct {
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Altitude string `json:"alt"`
}

type DTOperation struct {
	Name           string `json:"name"`
	Number         string `json:"num"`
	Enabled        bool   `json:"enabled"`
	ModVersion     int    `json:"modversion"`
	Params         string `json:"params"`
	MultiName      string `json:"multi_name"`
	MultiPriority  int    `json:"multi_priority"`
	BlendOpVersion int    `json:"blendop_version"`
	BlendOpParams  string `json:"blendop_params"`
	IOPOrder       string `json:"iop_order"`
}

/*
	data to track:
		- source file (jpg, raw)
		- XMP existence
		- XMP data
		- If a duplicate exists
		- EXIF data

	actions:
		- resize
			+ via darktable or vips
			+ deciding algorithm should live here
		- change ratings
		- change tags



	general actions that relate to photos:
		- listing photos (by dir, or other)
		- filtering, sorting

*/

type Photo struct {
	Src string

	exifRead bool
	exif     map[string]interface{}

	xmpRead bool // detecting zero value may not work, since xmp may not exist. Need to know if we tried
	xmp     XMP

	// cached fields
	filesize       int64
	xmpExists      bool
	searchedForXMP bool
	sourceModTime  time.Time
	xmppModTime    time.Time

	ctx context.Context // awkward way to fetch some external fields
}

// @todo: track XMP file name, and stop adding ".xmp" to Src everywhere

type Resource struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Create a photo instance using the full path to original source
func FromSrc(ctx context.Context, src string) (Photo, error) {
	fi, err := os.Stat(src)
	if err != nil {
		return Photo{}, err
	} else if fi.IsDir() {
		return Photo{}, errors.New("src is a directory")
	}

	return Photo{
		Src:           src,
		ctx:           ctx,
		sourceModTime: fi.ModTime(),
		filesize:      fi.Size(),
	}, nil
}

func FromThumb(ctx context.Context, path string) (Photo, error) {
	return Photo{
		ctx: ctx,
	}, nil
}

func (p *Photo) HasXMP() bool {
	if !p.searchedForXMP {
		p.searchedForXMP = true
		if _, err := os.Stat(p.Src + ".xmp"); err == nil {
			p.xmpExists = true
		} else {
			// look for duplicates?
			ext := filepath.Ext(p.Src)
			base := strings.TrimSuffix(p.Src, ext)
			if m, err := filepath.Glob(base + "_??" + ext); err == nil && len(m) > 0 {
				p.xmpExists = true
			}
		}
	}

	return p.xmpExists
}

// modification time of source image
func (p *Photo) ModTime() time.Time {
	if p.sourceModTime.IsZero() {
		fi, err := os.Stat(p.Src)
		if err != nil {
			// @todo: surface the error
			return time.Time{}
		}
		p.sourceModTime = fi.ModTime()
	}
	return p.sourceModTime
}

// modification time of XMP, if available
func (p *Photo) XModTime() time.Time {
	if !p.xmpExists { // no XMP, no mod time
		return time.Time{}
	}

	if p.xmppModTime.IsZero() {
		fi, err := os.Stat(p.Src + ".xmp")
		if err != nil {
			// @todo: surface the error
			return time.Time{}
		}
		p.xmppModTime = fi.ModTime()
	}
	return p.xmppModTime
}

// get last modification time (source image or XMP)
func (p *Photo) LastMod() time.Time {
	src := p.ModTime()
	x := p.XModTime()
	if src.After(x) {
		return src
	}
	return x
}

// retrieve a value from XMP if available, falling back to exif
func (p *Photo) Meta(field string) (interface{}, error) {
	if p.HasXMP() {
		x, xload_err := p.XMP()
		switch field {
		case "DerivedFromFile":
			return x.DerivedFromFile, xload_err
		case "AutoPresets":
			return x.AutoPresets, xload_err
		case "XMPVersion":
			return x.XMPVersion, xload_err
		case "ColorLabels":
			return x.ColorLabels, xload_err
		case "History":
			return x.History, xload_err
		case "Tags":
			return x.Tags, xload_err
		case "Title":
			return x.Title, xload_err

		// the below cases can be backfilled from Exif data
		case "Location":
			if xload_err == nil && x.Location != nil {
				return x.Location, nil
			}
		case "Rating":
			if xload_err == nil && x.DerivedFromFile != "" { // DFF always populated if XMP read
				return x.Rating, nil
			}
		case "Rights", "Copyright":
			if xload_err == nil && x.Rights != "" {
				return x.Rights, nil
			}
			field = "Copyright" // translate to EXIF name
		case "Creator", "Artist":
			if xload_err == nil && x.Creator != "" {
				return x.Creator, nil
			}
			field = "Artist" // translate to EXIF name
		}
	}

	// could not look up via XMP, check EXIF

	ex, err := p.Exif()
	if err != nil {
		return nil, err
	}

	v, has := ex[field]
	if !has {
		// @todo: key not found? is that an error?
		return nil, nil
	}

	return v, nil
}

// retrieve a meta value as a string, swallowing errors and returning empty string
// if not exists, or error
func (p *Photo) MetaString(field string) string {
	v, _ := p.Meta(field)

	if v == nil {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// retrieve a meta value as an int, swallowing errors and returning 0
// if not exists, or error
func (p *Photo) MetaInt(field string) int {
	v, _ := p.Meta(field)

	if v == nil {
		return 0
	}

	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	case bool:
		if n {
			return 1
		} else {
			return 0
		}
	}

	return 0
}

// get photo's XMP data. Will use a cache if already populated.
// will return an empty XMP{} if there is no XMP file
func (p *Photo) XMP() (XMP, error) {
	if !p.xmpRead && p.HasXMP() {
		if err := p.loadXmp(); err != nil {
			return XMP{}, err
		}
	}
	return p.xmp, nil
}

// get associated Exif data, possibly from cache.
func (p *Photo) Exif() (map[string]interface{}, error) {
	if !p.exifRead {
		if err := p.loadExif(); err != nil {
			return nil, err
		}
	}
	return p.exif, nil
}

func (p *Photo) Size() (int, int) {
	// @todo: this doesn't account for darktable crops

	w, h := 0, 0
	p.Ex_if_int("ImageWidth", func(i int) { w = i })
	p.Ex_if_int("ImageHeight", func(i int) { h = i })

	if p.Rotation() == Portrait {
		w, h = h, w
	}

	return w, h
}

func (p *Photo) FileSize() (int64, error) {
	if p.filesize <= 0 {
		fi, err := os.Stat(p.Src)
		if err != nil {
			return -1, err
		}
		p.filesize = fi.Size()
	}
	return p.filesize, nil
}

func (p Photo) Relpath() string {
	photoDir := p.ctx.Value("photoDir").(string)
	return strings.TrimPrefix(p.Src, photoDir+"/")
}

func (p *Photo) Orientation() Orientation {
	o := OrientationInvalid
	p.Ex_if_string("Orientation", func(s string) {
		o = parseOrientation(s)
	})
	return o
}

// rotation convenience function
type RotMode int

const (
	Portrait RotMode = iota
	Landscape
)

func (r RotMode) String() string {
	if int(r) == 0 {
		return "portrait"
	}
	return "landscape"
}

func (p Photo) Rotation() RotMode {
	if p.Orientation() > Rot180 {
		return Portrait
	}
	return Landscape
}

/*  EXIF rotation values */

type Orientation int

// rotations are specified in a clockwise direction
// https://www.daveperrett.com/articles/2012/07/28/exif-orientation-handling-is-a-ghetto/
const (
	OrientationInvalid Orientation = -1
	RotNormal          Orientation = 0
	MirrorVert         Orientation = 1
	MirrorHoriz        Orientation = 2
	Rot180             Orientation = 3
	MirrorHorizRot270  Orientation = 4
	Rot90              Orientation = 5
	Rot270             Orientation = 6
	MirrorHorizRot90   Orientation = 7
)

func parseOrientation(s string) Orientation {
	switch s {
	case "Horizontal (normal)":
		return 0
	case "Mirror vertical":
		return 1
	case "Mirror horizontal":
		return 2
	case "Rotate 180":
		return 3
	case "Mirror horizontal and rotate 270 CW":
		return 4
	case "Rotate 90 CW":
		return 5
	case "Rotate 270 CW":
		return 6
	case "Mirror horizontal and rotate 90 CW":
		return 7
	}
	return OrientationInvalid
}

/* internal helpers, lazy loaders, etc */

// run callback if property exists, and successfully converts
func (p *Photo) Ex_if_int(prop string, f func(int)) {
	if !p.exifRead {
		p.loadExif() // @todo: surface this
	}
	if val, ok := p.exif[prop]; ok {
		if fl, ok := val.(float64); ok { // numbers read as float64 from JSON conversion
			f(int(fl))
		}
	}
}
func (p *Photo) Ex_if_string(prop string, f func(string)) {
	if !p.exifRead {
		p.loadExif() // @todo: surface this
	}
	if val, ok := p.exif[prop]; ok {
		if s, ok := val.(string); ok {
			f(s)
		}
	}
}

func (p *Photo) loadExif() error {
	db := p.ctx.Value("badger").(*badger.DB)

	t, err := ReadModTime(db, "EXIF", p.Relpath())
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	var loader func() (map[string]interface{}, error)
	if err != nil || p.ModTime().After(t) {
		loader = p.loadExifFromFile // @todo: write to db/index if used
	} else {
		loader = p.loadExifFromDB
	}

	ex, err := loader()
	if err != nil {
		return err
	}
	p.exif = ex
	p.exifRead = true
	return nil
}

func (p Photo) loadExifFromFile() (map[string]interface{}, error) { return ReadExif(p.Src) }
func (p Photo) loadExifFromDB() (map[string]interface{}, error) {
	return readExifDB(p.ctx.Value("badger").(*badger.DB), p.Relpath())
}

func (p *Photo) loadXmp() error {
	db := p.ctx.Value("badger").(*badger.DB)

	t, err := ReadModTime(db, "XMP", p.Relpath())
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}

	var loader func() (XMP, error)
	if err != nil || p.ModTime().After(t) {
		loader = p.loadXMPFromFile // @todo: write to DB if used?
	} else {
		loader = p.loadXMPFromDB
	}
	m, err := loader()
	if err != nil {
		return err
	}
	p.xmp = m
	p.xmpRead = true
	return nil
}

func (p *Photo) loadXMPFromFile() (XMP, error) { return ReadXMP(p.Src + ".xmp") }
func (p *Photo) loadXMPFromDB() (XMP, error) {
	return readXMPDB(p.ctx.Value("badger").(*badger.DB), p.Relpath())
}

/* JSON */

func (p Photo) ThumbSizes() map[Size]Resource {
	sizes := []Size{SizeXS, SizeSmall, SizeMedium, SizeLarge, SizeXL, SizeFull}
	host := p.ctx.Value("host").(string)
	relpath := p.Relpath()
	jpg := strings.TrimSuffix(relpath, filepath.Ext(relpath)) + ".jpg"
	w, h := p.Size()
	thumbs := make(map[Size]Resource, len(sizes))
	for _, s := range sizes {
		var rw int
		var rh int

		if w > h {
			rw = int(s)
			rh = int(float64(h) / (float64(w) / float64(rw)))
		} else {
			rh = int(s)
			rw = int(float64(w) / (float64(h) / float64(rh)))
		}

		if s == SizeFull {
			rw = w
			rh = h
		}
		thumbs[s] = Resource{
			Width:  rw,
			Height: rh,
			URL:    "http://" + host + "/api/v1/thumb/" + s.String() + "/" + jpg,
		}
	}
	return thumbs
}

func (p Photo) MarshalJSON() ([]byte, error) {

	// output
	type PhotoJSON struct {
		Name        string                 `json:"name"`
		Size        int64                  `json:"size"`
		Rotation    string                 `json:"rotation"`
		Orientation int                    `json:"orientation"`
		XMP         XMP                    `json:"xmp"`
		Exif        map[string]interface{} `json:"exif"`
		Thumbs      map[Size]Resource      `json:"thumbs"`
		Original    Resource               `json:"original"`
	}

	fs, err := p.FileSize()
	if err != nil {
		return nil, err
	}

	exif, err := p.Exif()
	if err != nil {
		return nil, err
	}

	xmp, err := p.XMP()
	if err != nil {
		return nil, err
	}

	w, h := p.Size()
	host := p.ctx.Value("host").(string)
	relpath := p.Relpath()
	j := PhotoJSON{
		Name:        relpath,
		Size:        fs,
		XMP:         xmp,
		Exif:        exif,
		Rotation:    p.Rotation().String(),
		Orientation: int(p.Orientation()),
		Thumbs:      p.ThumbSizes(),
		Original: Resource{
			Width:  w,
			Height: h,
			URL:    "http://" + host + "/api/v1/photos/" + relpath,
		},
	}

	return json.Marshal(j)
}
