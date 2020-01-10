package darktable

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
)

/* Generic or reused across functions */

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
		Size:      math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Threshold: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Strength:  math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
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
		Radius: math.Float64frombits(binary.LittleEndian.Uint64(p[0:8])),
		Slope:  math.Float64frombits(binary.LittleEndian.Uint64(p[8:16])),
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
		Contrast:   math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Brightness: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Saturation: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
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
		SteepA:  math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		OffsetA: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		SteepB:  math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		OffsetB: math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
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
		HiA:        math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		HiB:        math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		LowA:       math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		LowB:       math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
		Saturation: math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
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
		Hue:                math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Saturation:         math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		SourceLightnessMix: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		Lightness:          math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
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
		MedianThreshold: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
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
			GreyPtSource:  math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
			BlackPtSource: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
			WhitePtSource: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
			Security:      math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
			GreyPtTarget:  math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
			BlackPtTarget: math.Float32frombits(binary.LittleEndian.Uint32(p[20:24])),
			WhitePtTarget: math.Float32frombits(binary.LittleEndian.Uint32(p[24:28])),
			Output:        math.Float32frombits(binary.LittleEndian.Uint32(p[28:32])),
			Latitude:      math.Float32frombits(binary.LittleEndian.Uint32(p[32:36])),
			Contrast:      math.Float32frombits(binary.LittleEndian.Uint32(p[36:40])),
			Saturation:    math.Float32frombits(binary.LittleEndian.Uint32(p[40:44])),
			Balance:       math.Float32frombits(binary.LittleEndian.Uint32(p[i : i+4])),
			PreserveColor: false,
		},
		GlobalSaturation: 100, // default for old versions
		Interpolator:     FilmicInterpolator(binary.LittleEndian.Uint32(p[i+4 : i+8])),
	}

	if v > 2 {
		f.GlobalSaturation = math.Float32frombits(binary.LittleEndian.Uint32(p[44:48]))
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
		GreyPtSource:  math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		BlackPtSource: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		WhitePtSource: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		Security:      math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
		GreyPtTarget:  math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
		BlackPtTarget: math.Float32frombits(binary.LittleEndian.Uint32(p[20:24])),
		WhitePtTarget: math.Float32frombits(binary.LittleEndian.Uint32(p[24:28])),
		Output:        math.Float32frombits(binary.LittleEndian.Uint32(p[28:32])),
		Latitude:      math.Float32frombits(binary.LittleEndian.Uint32(p[32:36])),
		Contrast:      math.Float32frombits(binary.LittleEndian.Uint32(p[36:40])),
		Saturation:    math.Float32frombits(binary.LittleEndian.Uint32(p[40:44])),
		Balance:       math.Float32frombits(binary.LittleEndian.Uint32(p[44:48])),
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
		Gamma:  math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Linear: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
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
		Density:    math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Hardness:   math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Rotation:   math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
		Offset:     math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
		Hue:        math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
		Saturation: math.Float32frombits(binary.LittleEndian.Uint32(p[20:24])),
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
		Scale:    math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
		Strength: math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
	}
	if v > 1 {
		g.MidtoneBias = math.Float32frombits(binary.LittleEndian.Uint32(p[12:16]))
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
		Strength: math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Distance: math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
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
		h.Clip = math.Float32frombits(binary.LittleEndian.Uint32(p[16:20]))
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
		Sharpness: math.Float32frombits(binary.LittleEndian.Uint32(p[0:4])),
		Contrast:  math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
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
			math.Float32frombits(binary.LittleEndian.Uint32(p[4:8])),
			math.Float32frombits(binary.LittleEndian.Uint32(p[8:12])),
			math.Float32frombits(binary.LittleEndian.Uint32(p[12:16])),
		},
		Levels: [3]float32{
			math.Float32frombits(binary.LittleEndian.Uint32(p[16:20])),
			math.Float32frombits(binary.LittleEndian.Uint32(p[20:24])),
			math.Float32frombits(binary.LittleEndian.Uint32(p[24:28])),
		},
	}, nil
}
