package darktable

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
)

type Op struct {
	Name           string      `json:"name"`
	Number         string      `json:"num"`
	Enabled        bool        `json:"enabled"`
	ModVersion     int         `json:"modversion"`
	RawParams      string      `json:"raw_params"`
	Params         interface{} `json:"params"`
	MultiName      string      `json:"multi_name"`
	MultiPriority  int         `json:"multi_priority"`
	BlendOpVersion int         `json:"blendop_version"`
	BlendOpParams  string      `json:"blendop_params"`
	IOPOrder       string      `json:"iop_order"`
}

func ParseOpParams(name string, v int, params string) (interface{}, error) {
	switch name {
	case "exposure":
		return exposure(v, params)
	}

	return nil, nil
}

type ExposureMode int

const (
	ExposureModeManual ExposureMode = iota
	ExposureModeDeflicker
)

func (e ExposureMode) MarshalJSON() ([]byte, error) {
	if e == ExposureModeManual {
		return json.Marshal("manual")
	}
	return json.Marshal("deflicker")
}

type ExposureParams struct {
	Mode            ExposureMode `json:"mode"`
	Black           float32      `json:"black"`
	Exposure        float32      `json:"exposure"`
	DeflickerPerctl float32      `json:"deflicker_percentile"`
	DeflickerTgt    float32      `json:"deflicker_target_level"`
}

func exposure(v int, params string) (ExposureParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ExposureParams{}, err
	}
	return ExposureParams{
		Mode:            ExposureMode(uint8(p[0])),
		Black:           math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Exposure:        math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		DeflickerPerctl: math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
		DeflickerTgt:    math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
	}, nil
}
