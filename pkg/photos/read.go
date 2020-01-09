package photos

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

// ---------------- XMP parsing -----------
type DTXMP struct {
	Description *struct {
		DerivedFrom          string   `xml:"DerivedFrom,attr"`
		Rating               string   `xml:"Rating,attr"`
		DTAutoPresetsApplied string   `xml:"auto_presets_applied,attr"`
		DTHistoryEnd         string   `xml:"history_end,attr"`
		DTRawParams          string   `xml:"raw_params,attr"`
		DTXMPVersion         string   `xml:"xmp_version,attr"`
		IOPOrderVersion      string   `xml:"iop_order_version,attr"`
		DTColorLabels        []string `xml:"colorlabels>Seq>li"`
		Creator              []string `xml:"creator>Seq>li"`
		Title                []string `xml:"title>Alt>li,omitempty"`
		Altitude             string   `xml:"GPSAltitude,attr"`
		AltitudeRef          string   `xml:"GPSAltitudeRef,attr"`
		Latitude             string   `xml:"GPSLatitude,attr"`
		Longitude            string   `xml:"GPSLongitude,attr"`
		GPSVerID             string   `xml:"GPSVersionID,attr"`
		DTHistory            []*struct {
			Num            string `xml:"num,attr"`
			Operation      string `xml:"operation,attr"`
			Enabled        string `xml:"enabled,attr"`
			ModVersion     string `xml:"modversion,attr"`
			Params         string `xml:"params,attr"`
			MultiName      string `xml:"multi_name,attr"`
			MultiPriority  string `xml:"multi_priority,attr"`
			IOPOrder       string `xml:"iop_order,attr"`
			BlendOpVersion string `xml:"blendop_version,attr"`
			BlendOpParams  string `xml:"blendop_params,attr"`
		} `xml:"history>Seq>li,omitempty"`
		DTTags        []string `xml:"hierarchicalSubject>Seq>li,omitempty"`
		DTMask        []string `xml:"mask>Seq>li,omitempty"`
		DTMaskID      []string `xml:"mask_id>Seq>li,omitempty"`
		DTMaskName    []string `xml:"mask_name>Seq>li,omitempty"`
		DTMaskNB      []string `xml:"mask_nb>Seq>li,omitempty"`
		DTMaskSrc     []string `xml:"mask_src>Seq>li,omitempty"`
		DTMaskType    []string `xml:"mask_type>Seq>li,omitempty"`
		DTMaskVersion []string `xml:"mask_version>Seq>li,omitempty"`
		Rights        []string `xml:"rights>Alt>li,omitempty"`
	} `xml:"RDF>Description"`
}

// read the contents of an XMP file. Absolute path expected
func ReadXMP(file string) (XMP, error) {

	f, err := ioutil.ReadFile(file)
	if err != nil {
		return XMP{}, err
	}

	var d DTXMP

	if err := xml.Unmarshal(f, &d); err != nil {
		return XMP{}, err
	}

	rating, err := strconv.Atoi(d.Description.Rating)
	if err != nil {
		return XMP{}, err
	}
	xmpV, err := strconv.Atoi(d.Description.DTXMPVersion)
	if err != nil {
		return XMP{}, err
	}

	ops := make([]DTOperation, len(d.Description.DTHistory))
	for i, h := range d.Description.DTHistory {
		mv, err := strconv.Atoi(h.ModVersion)
		if err != nil {
			return XMP{}, err
		}
		mp, err := strconv.Atoi(h.MultiPriority)
		if err != nil {
			return XMP{}, err
		}
		bv, err := strconv.Atoi(h.BlendOpVersion)
		if err != nil {
			return XMP{}, err
		}
		ops[i] = DTOperation{
			Name:           h.Operation,
			Enabled:        h.Enabled == "1",
			ModVersion:     mv,
			Params:         h.Params,
			MultiName:      h.MultiName,
			MultiPriority:  mp,
			BlendOpVersion: bv,
			BlendOpParams:  h.BlendOpParams,
			Number:         h.Num,
			IOPOrder:       h.IOPOrder,
		}
	}

	var l *Location
	if d.Description.Latitude != "" && d.Description.Longitude != "" {
		l = &Location{
			Lat:      d.Description.Latitude,
			Lon:      d.Description.Longitude,
			Altitude: d.Description.Altitude,
		}
	}

	return XMP{
		DerivedFromFile: d.Description.DerivedFrom,
		Rating:          rating,
		AutoPresets:     d.Description.DTAutoPresetsApplied == "1",
		XMPVersion:      xmpV,
		ColorLabels:     d.Description.DTColorLabels,
		Creator:         strings.Join(d.Description.Creator, ", "),
		Rights:          strings.Join(d.Description.Rights, ", "),
		History:         ops,
		Location:        l,
		Title:           strings.Join(d.Description.Title, ", "),
		Tags:            d.Description.DTTags,
	}, nil
}

/* ------------ EXIF parsing --------------- */

// read the exif properties of the given path. Absolute file path expected
func ReadExif(file string) (map[string]interface{}, error) {

	cmd := exec.Command("exiftool", "-j", file)

	var so bytes.Buffer
	var se bytes.Buffer

	op, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	ep, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	io.Copy(&so, op) // nolint
	io.Copy(&se, ep) // nolint

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	if se.Len() > 0 {
		fmt.Printf("got exiftool stderr: %s\n", se.String())
	}

	var j interface{}
	if err := json.Unmarshal(so.Bytes(), &j); err != nil {
		return nil, err
	}

	exif := make(map[string]interface{})
	if si, ok := j.([]interface{}); ok {
		if m, ok2 := si[0].(map[string]interface{}); ok2 {
			exif = m
		} else {
			return nil, errors.New("unable to parse exiftool output into map")
		}
	} else {
		return nil, errors.New("unable to parse exiftool output")
	}

	return exif, nil
}
