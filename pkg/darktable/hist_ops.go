package darktable

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
)

/* Generic or reused across functions */

func mkfloat(p []byte) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(p)) }
func mk64f(p []byte) float64   { return math.Float64frombits(binary.LittleEndian.Uint64(p)) }

type Point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// ----

type AShiftMode int

const (
	AShiftModeGeneric AShiftMode = iota
	AShiftModeSpecific
)

func (a AShiftMode) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }
func (a AShiftMode) String() string {
	if a == AShiftModeGeneric {
		return "generic"
	}
	return "specific"
}

type AShiftCropMode int

const (
	AShiftCropOff AShiftCropMode = iota
	AShiftCropLargest
	AShiftCropAspect
)

func (a AShiftCropMode) MarshalJSON() ([]byte, error) { return json.Marshal(a.String()) }
func (a AShiftCropMode) String() string {
	switch a {
	case AShiftCropOff:
		return "off"
	case AShiftCropLargest:
		return "largest"
	case AShiftCropAspect:
		return "aspect"
	}
	return "unknown"
}

type AShiftParams struct {
	Rotation   float32        `json:"rotation"`
	LensShiftV float32        `json:"lens_shift_v"`
	LensShiftH float32        `json:"lens_shift_h"`
	Shear      float32        `json:"shear"`
	FLength    float32        `json:"f_length"`
	CropFactor float32        `json:"crop_factor"`
	OrthoCorr  float32        `json:"ortho_corr"`
	Aspect     float32        `json:"aspect"`
	Mode       AShiftMode     `json:"mode"`
	Toggle     int            `json:"toggle"`
	Crop       AShiftCropMode `json:"crop"`
	CL         float32        `json:"cl"`
	CR         float32        `json:"cr"`
	CT         float32        `json:"ct"`
	CB         float32        `json:"cb"`
}

func ashift(v int, params string) (AShiftParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return AShiftParams{}, err
	}

	a := AShiftParams{
		Rotation:   mkfloat(p[0:4]),
		LensShiftV: mkfloat(p[4:8]),
		LensShiftH: mkfloat(p[8:12]),
	}
	if v == 1 {
		a.Toggle = int(binary.LittleEndian.Uint32(p[12:16]))
	}

	if v > 1 {
		i := 0

		if v > 3 {
			a.Shear = mkfloat(p[12:16])
			i = 4
		}
		a.FLength = mkfloat(p[12+i : 16+i])
		a.CropFactor = mkfloat(p[16+i : 20+i])
		a.Aspect = mkfloat(p[24+i : 28+i])
		a.Mode = AShiftMode(binary.LittleEndian.Uint32(p[28+i : 32+i]))
		a.Toggle = int(binary.LittleEndian.Uint32(p[: 32+i : 36+i]))
		if v > 2 {
			a.Crop = AShiftCropMode(binary.LittleEndian.Uint32(p[36+i : 40+i]))
			a.CL = mkfloat(p[40+i : 44+i])
			a.CR = mkfloat(p[44+i : 48+i])
			a.CT = mkfloat(p[48+i : 52+i])
			a.CB = mkfloat(p[52+i : 56+i])
		}
	}
	return a, nil
}

type AtrousParams struct {
	Octaves int32 `json:"octaves"`
	/* as laid out in atrous.c
	X       [5][6]float32 `json:"x"`
	Y       [5][6]float32 `json:"y"`
	*/
	Luminance   [6]Point `json:"luminance"`
	Chrominance [6]Point `json:"chrominance"`
	Sharpness   [6]Point `json:"sharpness"`
	LumNoise    [6]Point `json:"luminance_noise"`
	ChrNoise    [6]Point `json:"chrominance_noise"`
}

func atrous(v int, params string) (AtrousParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return AtrousParams{}, err
	}

	a := AtrousParams{
		Octaves: int32(binary.LittleEndian.Uint32(p[0:4])),
	}
	p = p[4:] // easy way to offset

	// memory layout is:
	// var x [5][6]float32
	// var y [5][6]float32
	// so 5*6*4 = 120 offset between a corresponding X and Y

	a.Luminance = [6]Point{
		Point{mkfloat(p[0:4]), mkfloat(p[120+0 : 120+4])},
		Point{mkfloat(p[4:8]), mkfloat(p[120+4 : 120+8])},
		Point{mkfloat(p[8:12]), mkfloat(p[120+8 : 120+12])},
		Point{mkfloat(p[12:16]), mkfloat(p[120+12 : 120+16])},
		Point{mkfloat(p[16:20]), mkfloat(p[120+16 : 120+20])},
		Point{mkfloat(p[20:24]), mkfloat(p[120+20 : 120+24])},
	}
	p = p[24:] // shift off luminance, 120-offset remains constant
	a.Chrominance = [6]Point{
		Point{mkfloat(p[0:4]), mkfloat(p[120+0 : 120+4])},
		Point{mkfloat(p[4:8]), mkfloat(p[120+4 : 120+8])},
		Point{mkfloat(p[8:12]), mkfloat(p[120+8 : 120+12])},
		Point{mkfloat(p[12:16]), mkfloat(p[120+12 : 120+16])},
		Point{mkfloat(p[16:20]), mkfloat(p[120+16 : 120+20])},
		Point{mkfloat(p[20:24]), mkfloat(p[120+20 : 120+24])},
	}
	p = p[24:]
	a.Sharpness = [6]Point{
		Point{mkfloat(p[0:4]), mkfloat(p[120+0 : 120+4])},
		Point{mkfloat(p[4:8]), mkfloat(p[120+4 : 120+8])},
		Point{mkfloat(p[8:12]), mkfloat(p[120+8 : 120+12])},
		Point{mkfloat(p[12:16]), mkfloat(p[120+12 : 120+16])},
		Point{mkfloat(p[16:20]), mkfloat(p[120+16 : 120+20])},
		Point{mkfloat(p[20:24]), mkfloat(p[120+20 : 120+24])},
	}
	p = p[24:]
	a.LumNoise = [6]Point{
		Point{mkfloat(p[0:4]), mkfloat(p[120+0 : 120+4])},
		Point{mkfloat(p[4:8]), mkfloat(p[120+4 : 120+8])},
		Point{mkfloat(p[8:12]), mkfloat(p[120+8 : 120+12])},
		Point{mkfloat(p[12:16]), mkfloat(p[120+12 : 120+16])},
		Point{mkfloat(p[16:20]), mkfloat(p[120+16 : 120+20])},
		Point{mkfloat(p[20:24]), mkfloat(p[120+20 : 120+24])},
	}
	p = p[24:]
	a.ChrNoise = [6]Point{
		Point{mkfloat(p[0:4]), mkfloat(p[120+0 : 120+4])},
		Point{mkfloat(p[4:8]), mkfloat(p[120+4 : 120+8])},
		Point{mkfloat(p[8:12]), mkfloat(p[120+8 : 120+12])},
		Point{mkfloat(p[12:16]), mkfloat(p[120+12 : 120+16])},
		Point{mkfloat(p[16:20]), mkfloat(p[120+16 : 120+20])},
		Point{mkfloat(p[20:24]), mkfloat(p[120+20 : 120+24])},
	}
	return a, nil

}

type ExposureMode int

const (
	ExposureModeManual ExposureMode = iota
	ExposureModeDeflicker
)

func (e ExposureMode) MarshalJSON() ([]byte, error) { return json.Marshal(e.String()) }
func (e ExposureMode) String() string {
	if e == ExposureModeManual {
		return "manual"
	}
	return "deflicker"
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
		Black:           mkfloat(p[4:8]),
		Exposure:        mkfloat(p[8:12]),
		DeflickerPerctl: mkfloat(p[12:16]),
		DeflickerTgt:    mkfloat(p[16:20]),
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
		Amount: mkfloat(p),
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
		Radius:    mkfloat(p[0:4]),
		Amount:    mkfloat(p[4:8]),
		Threshold: mkfloat(p[8:12]),
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
		Size:       mkfloat(p[0:4]),
		Saturation: mkfloat(p[4:8]),
		Brightness: mkfloat(p[8:12]),
		Amount:     mkfloat(p[12:16]),
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
		SigmaR:  mkfloat(p[i : i+4]),
		SigmaS:  mkfloat(p[i+4 : i+8]),
		Detail:  mkfloat(p[i+8 : i+12]),
		MidTone: 0.2,
	}

	if v >= 2 {
		b.Mode = BilatMode(binary.LittleEndian.Uint32(p[0:4]))
	}
	if v >= 3 {
		b.MidTone = mkfloat(p[i+12 : i+16])
	}

	return b, nil
}

type BloomParams struct {
	Size      float32 `json:"size"`
	Threshold float32 `json:"threshold"`
	Strength  float32 `json:"strength"`
}

func bloom(v int, params string) (BloomParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return BloomParams{}, err
	}

	return BloomParams{
		Size:      mkfloat(p[0:4]),
		Threshold: mkfloat(p[4:8]),
		Strength:  mkfloat(p[8:12]),
	}, nil
}

type LCLContrastParams struct {
	Radius float64 `json:"radius"`
	Slope  float64 `json:"slope"`
}

func clahe(v int, params string) (LCLContrastParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return LCLContrastParams{}, err
	}

	return LCLContrastParams{
		Radius: mk64f(p[0:8]),
		Slope:  mk64f(p[8:16]),
	}, nil
}

// contrast brightness saturation
type ColisaParams struct {
	Contrast   float32 `json:"contrast"`
	Brightness float32 `json:"brightness"`
	Saturation float32 `json:"saturation"`
}

func colisa(v int, params string) (ColisaParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ColisaParams{}, err
	}

	return ColisaParams{
		Contrast:   mkfloat(p[0:4]),
		Brightness: mkfloat(p[4:8]),
		Saturation: mkfloat(p[8:12]),
	}, nil
}

type ColorContrastParams struct {
	SteepA  float32 `json:"steep_a"`
	OffsetA float32 `json:"offset_a"`
	SteepB  float32 `json:"steep_b"`
	OffsetB float32 `json:"offset_b"`
	Unbound bool    `json:"unbound"`
}

func colorcontrast(v int, params string) (ColorContrastParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ColorContrastParams{}, err
	}

	unbound := false // v1 default
	if v > 1 && binary.LittleEndian.Uint32(p[16:20]) > 0 {
		unbound = true
	}
	return ColorContrastParams{
		SteepA:  mkfloat(p[0:4]),
		OffsetA: mkfloat(p[4:8]),
		SteepB:  mkfloat(p[8:12]),
		OffsetB: mkfloat(p[12:16]),
		Unbound: unbound,
	}, nil
}

type ColorCorrectionParams struct {
	HiA        float32 `json:"hi_a"`
	HiB        float32 `json:"hi_b"`
	LowA       float32 `json:"low_a"`
	LowB       float32 `json:"low_b"`
	Saturation float32 `json:"saturation"`
}

func colorcorrection(v int, params string) (ColorCorrectionParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ColorCorrectionParams{}, err
	}
	return ColorCorrectionParams{
		HiA:        mkfloat(p[0:4]),
		HiB:        mkfloat(p[4:8]),
		LowA:       mkfloat(p[8:12]),
		LowB:       mkfloat(p[12:16]),
		Saturation: mkfloat(p[16:20]),
	}, nil
}

type ColorizeParams struct {
	Hue                float32 `json:"hue"`
	Saturation         float32 `json:"saturation"`
	SourceLightnessMix float32 `json:"source_lightness_mix"`
	Lightness          float32 `json:"lightness"`
	Version            int     `json:"version"` // colorize v1 uses different math internally
}

func colorize(v int, params string) (ColorizeParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ColorizeParams{}, err
	}
	return ColorizeParams{
		Hue:                mkfloat(p[0:4]),
		Saturation:         mkfloat(p[4:8]),
		SourceLightnessMix: mkfloat(p[8:12]),
		Lightness:          mkfloat(p[12:16]),
		Version:            int(binary.LittleEndian.Uint32(p[16:20])),
	}, nil
}

type DemosaicGreenEQ int

const (
	DemosaicGreenEQNo DemosaicGreenEQ = iota
	DemosaicGreenEQLocal
	DemosaicGreenEQFull
	DemosaicGreenEQBoth
)

func (d DemosaicGreenEQ) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }
func (d DemosaicGreenEQ) String() string {
	switch d {
	case DemosaicGreenEQNo:
		return "no"
	case DemosaicGreenEQLocal:
		return "local"
	case DemosaicGreenEQFull:
		return "full"
	case DemosaicGreenEQBoth:
		return "both"
	}
	return "unknown"
}

type DemosaicMethod int

const (
	DemosaicPPG             DemosaicMethod = 0
	DemosaicAmaze           DemosaicMethod = 1
	DemosaicVNG4            DemosaicMethod = 2
	DemosaicPassthroughMono DemosaicMethod = 3
	DemosaicVNG             DemosaicMethod = 1024 | 0
	DemosaicMarkesteijn     DemosaicMethod = 1024 | 1
	DemosaicMarkesteijn3    DemosaicMethod = 1024 | 2
	DemosaicFDC             DemosaicMethod = 1024 | 4
)

func (d DemosaicMethod) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }
func (d DemosaicMethod) String() string {
	switch d {
	case DemosaicPPG:
		return "PPG"
	case DemosaicAmaze:
		return "AMaZE"
	case DemosaicVNG4:
		return "VNG4"
	case DemosaicPassthroughMono:
		return "passthrough monochrome"
	case DemosaicVNG:
		return "VNG (xtrans)"
	case DemosaicMarkesteijn:
		return "Markesteijn-1 (xtrans)"
	case DemosaicMarkesteijn3:
		return "Markesteijn-3 (xtrans)"
	case DemosaicFDC:
		return "Frequency Domain Chroma (xtrans)"
	}
	return "unknown"
}

type DemosaicParams struct {
	GreenEQ         DemosaicGreenEQ `json:"green_eq"`
	MedianThreshold float32         `json:"median_threshold"`
	ColorSmoothing  uint32          `json:"color_smoothing"`
	Method          DemosaicMethod  `json:"method"`
	Unused          uint32          `json:"-"`
}

func demosaic(v int, params string) (DemosaicParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return DemosaicParams{}, err
	}

	return DemosaicParams{
		GreenEQ:         DemosaicGreenEQ(binary.LittleEndian.Uint32(p[0:4])),
		MedianThreshold: mkfloat(p[4:8]),
		ColorSmoothing:  binary.LittleEndian.Uint32(p[8:12]),
		Method:          DemosaicMethod(binary.LittleEndian.Uint32(p[12:16])),
		Unused:          binary.LittleEndian.Uint32(p[16:20]),
	}, nil
}

type FilmicInterpolator int

const (
	FilmicInterpolateCubicSpline FilmicInterpolator = iota
	FilmicInterpolateCatmullRom
	FilmicInterpolateMonotoneHermite
	FilmicInterpolateOptimized
)

func (f FilmicInterpolator) MarshalJSON() ([]byte, error) { return json.Marshal(f.String()) }
func (f FilmicInterpolator) String() string {
	switch f {
	case FilmicInterpolateCubicSpline:
		return "Cubic Spline"
	case FilmicInterpolateCatmullRom:
		return "Catmull-Rom"
	case FilmicInterpolateMonotoneHermite:
		return "Monotone"
	case FilmicInterpolateOptimized:
		return "Optimized"
	}
	return "unknown"
}

// common to both Filmic and Filmic RGB modules
type FilmicCommonParams struct {
	GreyPtSource  float32 `json:"grey_point_source"`
	BlackPtSource float32 `json:"black_point_source"`
	WhitePtSource float32 `json:"white_point_source"`
	Security      float32 `json:"security"`
	GreyPtTarget  float32 `json:"grey_point_target"`
	BlackPtTarget float32 `json:"black_point_target"`
	WhitePtTarget float32 `json:"white_point_target"`
	Output        float32 `json:"output_power"`
	Latitude      float32 `json:"latitude"`
	Contrast      float32 `json:"contrast"`
	Saturation    float32 `json:"saturation"`
	Balance       float32 `json:"balance"`
	PreserveColor bool    `json:"preserve_color"`
}

type FilmicParams struct {
	FilmicCommonParams
	Latitude         float32            `json:"latitude_stops"` // override json name
	GlobalSaturation float32            `json:"global_saturation"`
	Interpolator     FilmicInterpolator `json:"interpolator"`
}

func filmic(v int, params string) (FilmicParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return FilmicParams{}, err
	}

	// v2 added preserve color
	// v3 added global saturation
	i := 44
	if v > 2 {
		i = 48 // global saturation is 44:48 in v3+
	}

	f := FilmicParams{
		FilmicCommonParams: FilmicCommonParams{
			GreyPtSource:  mkfloat(p[0:4]),
			BlackPtSource: mkfloat(p[4:8]),
			WhitePtSource: mkfloat(p[8:12]),
			Security:      mkfloat(p[12:16]),
			GreyPtTarget:  mkfloat(p[16:20]),
			BlackPtTarget: mkfloat(p[20:24]),
			WhitePtTarget: mkfloat(p[24:28]),
			Output:        mkfloat(p[28:32]),
			Latitude:      mkfloat(p[32:36]),
			Contrast:      mkfloat(p[36:40]),
			Saturation:    mkfloat(p[40:44]),
			Balance:       mkfloat(p[i : i+4]),
			PreserveColor: false,
		},
		GlobalSaturation: 100, // default for old versions
		Interpolator:     FilmicInterpolator(binary.LittleEndian.Uint32(p[i+4 : i+8])),
	}

	if v > 2 {
		f.GlobalSaturation = mkfloat(p[44:48])
	}
	if v > 1 {
		f.PreserveColor = binary.LittleEndian.Uint32(p[i+8:i+12]) > 0
	}

	return f, nil
}

func filmicrgb(v int, params string) (FilmicCommonParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return FilmicCommonParams{}, err
	}

	return FilmicCommonParams{
		GreyPtSource:  mkfloat(p[0:4]),
		BlackPtSource: mkfloat(p[4:8]),
		WhitePtSource: mkfloat(p[8:12]),
		Security:      mkfloat(p[12:16]),
		GreyPtTarget:  mkfloat(p[16:20]),
		BlackPtTarget: mkfloat(p[20:24]),
		WhitePtTarget: mkfloat(p[24:28]),
		Output:        mkfloat(p[28:32]),
		Latitude:      mkfloat(p[32:36]),
		Contrast:      mkfloat(p[36:40]),
		Saturation:    mkfloat(p[40:44]),
		Balance:       mkfloat(p[44:48]),
		PreserveColor: binary.LittleEndian.Uint32(p[48:52]) > 0,
	}, nil

}

type GammaParams struct {
	Gamma  float32 `json:"gamma"`
	Linear float32 `json:"linear"`
}

func gamma(v int, params string) (GammaParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return GammaParams{}, err
	}

	return GammaParams{
		Gamma:  mkfloat(p[0:4]),
		Linear: mkfloat(p[4:8]),
	}, nil
}

type GraduatedNDparams struct {
	Density    float32 `json:"density"`  // density of filter, 0-8EV
	Hardness   float32 `json:"hardness"` // 0% soft, 100% hard
	Rotation   float32 `json:"rotation"` // 2*Pi  -180 <-> 180
	Offset     float32 `json:"offset"`   // default 50%, can be offset
	Hue        float32 `json:"hue"`
	Saturation float32 `json:"saturation"`
}

func graduatednd(v int, params string) (GraduatedNDparams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return GraduatedNDparams{}, err
	}
	return GraduatedNDparams{
		Density:    mkfloat(p[0:4]),
		Hardness:   mkfloat(p[4:8]),
		Rotation:   mkfloat(p[8:12]),
		Offset:     mkfloat(p[12:16]),
		Hue:        mkfloat(p[16:20]),
		Saturation: mkfloat(p[20:24]),
	}, nil
}

type GrainChannel int

const (
	GrainChannelHue GrainChannel = iota
	GrainChannelSaturation
	GrainChannelLightness
	GrainChannelRGB
)

func (g GrainChannel) MarshalJSON() ([]byte, error) { return json.Marshal(g.String()) }
func (g GrainChannel) String() string {
	switch g {
	case GrainChannelHue:
		return "hue"
	case GrainChannelSaturation:
		return "saturation"
	case GrainChannelLightness:
		return "lightness"
	case GrainChannelRGB:
		return "rgb"
	}
	return "unknown"
}

type GrainParam struct {
	Channel     GrainChannel `json:"channel"`
	Scale       float32      `json:"scale"`
	Strength    float32      `json:"strength"`
	MidtoneBias float32      `json:"midtone_bias"`
}

func grain(v int, params string) (GrainParam, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return GrainParam{}, err
	}
	g := GrainParam{
		Channel:  GrainChannel(binary.LittleEndian.Uint32(p[0:4])),
		Scale:    mkfloat(p[4:8]),
		Strength: mkfloat(p[8:12]),
	}
	if v > 1 {
		g.MidtoneBias = mkfloat(p[12:16])
	}
	return g, nil
}

type HazeParams struct {
	Strength float32 `json:"strength"`
	Distance float32 `json:"distance"`
}

func hazeremoval(v int, params string) (HazeParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return HazeParams{}, err
	}
	return HazeParams{
		Strength: mkfloat(p[0:4]),
		Distance: mkfloat(p[4:8]),
	}, nil
}

type HighlightsMode int

const (
	HighlightsClip HighlightsMode = iota
	HighlightsLCh
	HighlightsInpaint
)

func (h HighlightsMode) MarshalJSON() ([]byte, error) { return json.Marshal(h.String()) }
func (h HighlightsMode) String() string {
	switch h {
	case HighlightsClip:
		return "clip hightlights"
	case HighlightsLCh:
		return "reconstruct in LCh"
	case HighlightsInpaint:
		return "reconstruct in color"
	}
	return "unknown"
}

type HighlightsParams struct {
	Mode   HighlightsMode `json:"mode"`
	BlendL float32        `json:"-"` // unused
	BlendC float32        `json:"-"`
	BlendH float32        `json:"-"`
	Clip   float32        `json:"clip"`
}

func highlights(v int, params string) (HighlightsParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return HighlightsParams{}, err
	}
	h := HighlightsParams{
		Mode: HighlightsMode(binary.LittleEndian.Uint32(p[0:4])),
		Clip: 1.0,
	}
	if v > 1 {
		h.Clip = mkfloat(p[16:20])
	}
	return h, nil
}

type HighPassParams struct {
	Sharpness float32 `json:"sharpness"`
	Contrast  float32 `json:"contrast"`
}

func highpass(v int, params string) (HighPassParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return HighPassParams{}, err
	}
	return HighPassParams{
		Sharpness: mkfloat(p[0:4]),
		Contrast:  mkfloat(p[4:8]),
	}, nil
}

type LevelsMode int

const (
	LevelsManual LevelsMode = iota
	LevelsAutomatic
)

func (l LevelsMode) MarshalJSON() ([]byte, error) { return json.Marshal(l.String()) }
func (l LevelsMode) String() string {
	if l == LevelsManual {
		return "manual"
	}
	return "automatic"
}

type LevelsParams struct {
	Mode        LevelsMode `json:"mode"`
	Percentiles [3]float32 `json:"percentiles"`
	Levels      [3]float32 `json:"levels"`
}

func levels(v int, params string) (LevelsParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return LevelsParams{}, err
	}

	return LevelsParams{
		Mode: LevelsMode(binary.LittleEndian.Uint32(p[0:4])),
		Percentiles: [3]float32{
			mkfloat(p[4:8]), mkfloat(p[8:12]), mkfloat(p[12:16]),
		},
		Levels: [3]float32{
			mkfloat(p[16:20]), mkfloat(p[20:24]), mkfloat(p[24:28]),
		},
	}, nil
}

type LowlightParams struct {
	Blueness     float32    `json:"blueness"`
	TransitionX  [6]float32 `json:"transition_x"`
	TransistionY [6]float32 `json:"transition_y"`
}

func lowlight(v int, params string) (LowlightParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return LowlightParams{}, err
	}

	return LowlightParams{
		Blueness: mkfloat(p[0:4]),
		TransitionX: [6]float32{
			mkfloat(p[4:8]), mkfloat(p[8:12]), mkfloat(p[12:16]),
			mkfloat(p[16:20]), mkfloat(p[20:24]), mkfloat(p[24:28]),
		},
		TransistionY: [6]float32{
			mkfloat(p[28:32]), mkfloat(p[32:36]), mkfloat(p[36:40]),
			mkfloat(p[40:44]), mkfloat(p[44:48]), mkfloat(p[48:52]),
		},
	}, nil
}

type LowpassAlgo int

const (
	LowpassGaussian LowpassAlgo = iota
	LowpassBilateral
)

func (l LowpassAlgo) MarshalJSON() ([]byte, error) { return json.Marshal(l.String()) }
func (l LowpassAlgo) String() string {
	if l == LowpassGaussian {
		return "gaussian"
	}
	return "bilateral"
}

type LowpassParams struct {
	Order      uint32      `json:"order"`
	Radius     float32     `json:"radius"`
	Contrast   float32     `json:"contrast"`
	Brightness float32     `json:"brightness"`
	Saturation float32     `json:"saturation"`
	Algorithm  LowpassAlgo `json:"algo"`
	Unbound    bool        `json:"unbound"`
}

func lowpass(v int, params string) (LowpassParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return LowpassParams{}, err
	}

	l := LowpassParams{
		Order:      binary.LittleEndian.Uint32(p[0:4]),
		Radius:     mkfloat(p[4:8]),
		Contrast:   mkfloat(p[8:12]),
		Saturation: mkfloat(p[12:16]),
	}

	if v > 1 {
		l.Brightness = mkfloat(p[12:16])
		l.Saturation = mkfloat(p[16:20])
	}
	if v == 3 {
		l.Unbound = binary.LittleEndian.Uint32(p[20:24]) > 0
	}
	if v == 4 {
		l.Algorithm = LowpassAlgo(binary.LittleEndian.Uint32(p[20:24]))
		l.Unbound = binary.LittleEndian.Uint32(p[24:28]) > 0
	}

	return l, nil
}

type MonochromeParams struct {
	A          float32 `json:"a"`
	B          float32 `json:"b"`
	Size       float32 `json:"size"`
	Highlights float32 `json:"highlights"`
}

func monochrome(v int, params string) (MonochromeParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return MonochromeParams{}, err
	}

	m := MonochromeParams{
		A:          mkfloat(p[0:4]),
		B:          mkfloat(p[4:8]),
		Size:       mkfloat(p[8:12]),
		Highlights: 0,
	}

	if v > 1 {
		m.Highlights = mkfloat(p[12:16])
	}

	return m, nil
}

type NLMeansParams struct {
	Radius   float32 `json:"radius"`
	Strength float32 `json:"strength"`
	Luma     float32 `json:"luma"`
	Chroma   float32 `json:"chroma"`
}

func nlmeans(v int, params string) (NLMeansParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return NLMeansParams{}, err
	}

	i := 8
	if v == 1 {
		i = 0
	}
	n := NLMeansParams{
		Luma:     mkfloat(p[i : i+4]),
		Chroma:   mkfloat(p[i+4 : i+8]),
		Radius:   3, // v1 defaults
		Strength: 100,
	}

	if v > 1 {
		n.Radius = mkfloat(p[0:4])
		n.Strength = mkfloat(p[4:8])
	}

	return n, nil
}

type RelightAlgo int

const (
	RelightGaussian RelightAlgo = iota
	RelightBilateral
)

func (r RelightAlgo) MarshalJSON() ([]byte, error) { return json.Marshal(r.String()) }
func (r RelightAlgo) String() string {
	if r == RelightGaussian {
		return "gaussian"
	}
	return "bilateral"
}

type RelightParams struct {
	EV     float32 `json:"ev"`
	Center float32 `json:"center"`
	Width  float32 `json:"width"`
}

func relight(v int, params string) (RelightParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return RelightParams{}, err
	}
	return RelightParams{
		EV:     mkfloat(p[0:4]),
		Center: mkfloat(p[4:8]),
		Width:  mkfloat(p[8:12]),
	}, nil
}

type ShadhiAlgo int

const (
	ShadhiGaussian ShadhiAlgo = iota
	ShadhiBilateral
)

func (s ShadhiAlgo) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }
func (s ShadhiAlgo) String() string {
	if s == ShadhiGaussian {
		return "gaussian"
	}
	return "bilateral"
}

type ShadhiParams struct {
	Order              uint32     `json:"order"`
	Radius             float32    `json:"radius"`
	Shadows            float32    `json:"shadows"`
	Whitepoint         float32    `json:"whitepoint"` // reserved1
	Highlights         float32    `json:"highlights"`
	Reserved2          float32    `json:"-"`
	Compress           float32    `json:"compress"`
	ShadowsCCorrect    float32    `json:"shadows_ccorrect"`
	HighlightsCCorrect float32    `json:"highlights_ccorrect"`
	Flags              uint32     `json:"flags"`
	LowApprox          float32    `json:"low_approximation"`
	Algorithm          ShadhiAlgo `json:"algo"`
}

func shadhi(v int, params string) (ShadhiParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ShadhiParams{}, err
	}

	s := ShadhiParams{
		Order:           binary.LittleEndian.Uint32(p[0:4]),
		Radius:          mkfloat(p[4:8]),
		Shadows:         mkfloat(p[8:12]),
		Whitepoint:      mkfloat(p[12:16]), // reserved1 / ignored for v3 and below
		Highlights:      mkfloat(p[16:20]),
		Compress:        mkfloat(p[24:28]), // note skipped space for reserved2
		LowApprox:       0.01,
		ShadowsCCorrect: 100,
	}

	if v > 1 {
		s.ShadowsCCorrect = mkfloat(p[28:32])
		s.HighlightsCCorrect = mkfloat(p[32:36])
	}
	if v > 2 {
		s.Flags = binary.LittleEndian.Uint32(p[36:40])
	}
	if v > 3 {
		s.LowApprox = mkfloat(p[40:44])
	}
	if v > 4 {
		s.Algorithm = ShadhiAlgo(binary.LittleEndian.Uint32(p[44:48]))
	}

	return s, nil
}

type SplitToneParams struct {
	ShadowHue           float32 `json:"shadow_hue"`
	ShadowSaturation    float32 `json:"shadow_saturation"`
	HighlightHue        float32 `json:"highlight_hue"`
	HighlightSaturation float32 `json:"highlight_saturation"`
	Balance             float32 `json:"balance"`
	Compress            float32 `json:"compress"`
}

func splittoning(v int, params string) (SplitToneParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return SplitToneParams{}, err
	}

	return SplitToneParams{
		ShadowHue:           mkfloat(p[0:4]),
		ShadowSaturation:    mkfloat(p[4:8]),
		HighlightHue:        mkfloat(p[8:12]),
		HighlightSaturation: mkfloat(p[12:16]),
		Balance:             mkfloat(p[16:20]),
		Compress:            mkfloat(p[20:24]),
	}, nil
}

type ToneMapParams struct {
	Contrast float32 `json:"contrast"`
	FSize    float32 `json:"f_size"`
}

func tonemap(v int, params string) (ToneMapParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ToneMapParams{}, err
	}

	return ToneMapParams{
		Contrast: mkfloat(p[0:4]),
		FSize:    mkfloat(p[4:8]),
	}, nil
}

func velvia(v int, params string) (interface{}, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return nil, err
	}

	if v == 1 {
		return struct {
			Saturation float32 `json:"saturation"`
			Vibrance   float32 `json:"vibrance"`
			Luminance  float32 `json:"luminance"`
			Clarity    float32 `json:"clarity"`
		}{
			Saturation: mkfloat(p[0:4]),
			Vibrance:   mkfloat(p[4:8]),
			Luminance:  mkfloat(p[8:12]),
			Clarity:    mkfloat(p[12:16]),
		}, nil
	}

	return struct {
		Strength float32 `json:"strength"`
		Bias     float32 `json:"bias"`
	}{
		Strength: mkfloat(p[0:4]),
		Bias:     mkfloat(p[4:8]),
	}, nil
}

type ZoneSystemParams struct {
	Size int         `json:"size"`
	Zone [25]float32 `json:"zone"`
}

func zonesystem(v int, params string) (ZoneSystemParams, error) {
	p, err := hex.DecodeString(params)
	if err != nil {
		return ZoneSystemParams{}, err
	}

	z := ZoneSystemParams{
		Size: int(binary.LittleEndian.Uint32(p[0:4])),
	}

	for i := 4; i < len(p); i += 4 {
		z.Zone[(i-4)/4] = mkfloat(p[i : i+4])
	}

	return z, nil
}
