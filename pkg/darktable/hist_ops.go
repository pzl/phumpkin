package darktable

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
)

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

type SingleFloatAmount struct {
	Amount float32 `json:"amount"`
}

func vibrance(v int, params string) (SingleFloatAmount, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return SingleFloatAmount{}, err
	}

	return SingleFloatAmount{
		Amount: math.Float32frombits(binary.LittleEndian.Uint32(p)),
	}, nil
}

type SharpenParams struct {
	Radius    float32 `json:"radius"`
	Amount    float32 `json:"amount"`
	Threshold float32 `json:"threshold"`
}

func sharpen(v int, params string) (SharpenParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return SharpenParams{}, err
	}

	return SharpenParams{
		Radius:    math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Amount:    math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Threshold: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
	}, nil
}

type SoftenParams struct {
	Size       float32 `json:"size"`
	Saturation float32 `json:"saturation"`
	Brightness float32 `json:"brightness"`
	Amount     float32 `json:"amount"`
}

func soften(v int, params string) (SoftenParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return SoftenParams{}, err
	}
	return SoftenParams{
		Size:       math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Saturation: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Brightness: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		Amount:     math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
	}, nil
}

type BilatMode uint32

const (
	BilatModeBilateral BilatMode = iota
	BilatModeLocalLaplacian
)

func (b BilatMode) MarshalJSON() ([]byte, error) {
	if b == BilatModeBilateral {
		return json.Marshal("bilateral")
	}
	return json.Marshal("local laplacian")
}

type BilatParams struct {
	Mode    BilatMode `json:"mode"`
	SigmaR  float32   `json:"sigma_r"`
	SigmaS  float32   `json:"sigma_s"`
	Detail  float32   `json:"detail"`
	MidTone float32   `json:"midtone"`
}

func bilat(v int, params string) (BilatParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return BilatParams{}, err
	}

	i := 0 // starting byte for common fields
	if v >= 2 {
		i = 4 // Mode comes first after v1
	}
	b := BilatParams{ // defaults & common fields
		Mode:    BilatModeBilateral,
		SigmaR:  math.Float32frombits(binary.LittleEndian.Uint32(p[i : i+4])),
		SigmaS:  math.Float32frombits(binary.LittleEndian.Uint32(p[i+4 : i+8])),
		Detail:  math.Float32frombits(binary.LittleEndian.Uint32(p[i+8 : i+12])),
		MidTone: 0.2,
	}

	if v >= 2 {
		b.Mode = BilatMode(binary.LittleEndian.Uint32(p[0:4]))
	}
	if v >= 3 {
		b.MidTone = math.Float32frombits(binary.LittleEndian.Uint32(p[i+12 : i+16]))
	}

	return b, nil
}
