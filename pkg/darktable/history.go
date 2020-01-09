package darktable

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
)

type Op struct {
	Name           string      `json:"name"`
	OpName         string      `json:"op_name"`
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

func ParseHistory(num string, opname string, en string, ver string, rawparams string, multname string, multpri string, order string, bopv string, boparm string) Op {
	mv, err := strconv.Atoi(ver)
	if err != nil {
		mv = -1
	}
	mp, err := strconv.Atoi(multpri)
	if err != nil {
		mp = -1
	}
	bv, err := strconv.Atoi(bopv)
	if err != nil {
		bv = -1
	}
	op := Op{
		OpName:         opname,
		Name:           FriendlyHistoryName(opname),
		Enabled:        en == "1",
		ModVersion:     mv,
		RawParams:      rawparams,
		MultiName:      multname,
		MultiPriority:  mp,
		BlendOpVersion: bv,
		BlendOpParams:  boparm,
		Number:         num,
		IOPOrder:       order,
	}
	if prm, err := ParseOpParams(opname, mv, rawparams); err != nil {
		// @todo: log a low priority error, but don't block up XMP parsing for it
	} else {
		op.Params = prm
	}
	return op
}

func FriendlyHistoryName(s string) string {
	switch s {
	case "ashift":
		return "perspective correction"
	case "atrous":
		return "contrast equalizer"
	case "basecurve":
		return "base curve"
	case "basicadj":
		return "basic adjustments"
	case "bilat":
		return "local contrast"
	case "bilateral":
		return "denoise (bilateral filter)"
	case "borders":
		return "framing"
	case "cacorrect":
		return "chromatic aberrations"
	case "channelmixer":
		return "channel mixer"
	case "clahe":
		return "local contrast"
	case "clipping":
		return "crop and rotate"
	case "colisa":
		return "contrast brightness saturation"
	case "colorbalance":
		return "color balance"
	case "colorchecker":
		return "color look up table"
	case "colorcontrast":
		return "color contrast"
	case "colorcorrection":
		return "color correction"
	case "colorin":
		return "input color profile"
	case "colormapping":
		return "color mapping"
	case "colorout":
		return "output color profile"
	case "colorreconstruction":
		return "color reconstruction"
	case "colortransfer":
		return "color transfer"
	case "colorzones":
		return "color zones"
	case "denoiseprofile":
		return "denoise (profiled)"
	case "equalizer":
		return "legacy equalizer"
	case "filmicrgb":
		return "filmic rgb"
	case "finalscale":
		return "scale into final size"
	case "flip":
		return "orientation"
	case "globaltonemap":
		return "global tonemap"
	case "graduatednd":
		return "graduated density"
	case "hazeremoval":
		return "haze removal"
	case "highlights":
		return "hightlight reconstruction"
	case "hotpixels":
		return "hot pixels"
	case "lens":
		return "lens correction"
	case "lowlight":
		return "lowlight vision"
	case "lut3d":
		return "lut 3d"
	case "nlmeans":
		return "denoise (non-local means)"
	case "profile_gamma":
		return "unbreak input profile"
	case "rawdenoise":
		return "raw denoise"
	case "rawoverexposed":
		return "raw overexposed"
	case "rawprepare":
		return "raw black/white point"
	case "relight":
		return "fill light"
	case "rgbcurve":
		return "rgb curve"
	case "rgblevels":
		return "rgb levels"
	case "rotatepixels":
		return "rotate pixels"
	case "scalepixels":
		return "scale pixels"
	case "shadhi":
		return "shadows and highlights"
	case "splittoning":
		return "split-toning"
	case "spots":
		return "spot removal"
	case "temperature":
		return "white balance"
	case "tonecurve":
		return "tone curve"
	case "toneequal":
		return "tone equalizer"
	case "tonemap":
		return "tone mapping"
	case "zonesystem":
		return "zone system"
	case "bloom", "colorize", "defringe", "demosaic", "dither", "exposure", "filmic",
		"gamma", "grain", "highpass", "invert", "levels", "liquify", "lowpass", "monochrome",
		"overexposed", "retouch", "sharpen", "soften", "velvia", "vibrance", "vignette", "watermark":
		return s
	}
	return s
}

func ParseOpParams(name string, v int, params string) (interface{}, error) {
	switch name {
	case "bilat":
		return bilat(v, params)
	case "exposure":
		return exposure(v, params)
	case "vibrance":
		return vibrance(v, params)
	case "sharpen":
		return sharpen(v, params)
	case "soften":
		return soften(v, params)
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
