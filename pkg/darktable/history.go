package darktable

import (
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
		return "highlight reconstruction"
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
	case "ashift":
		return ashift(v, params)
	case "atrous":
		return atrous(v, params)
	case "bilat":
		return bilat(v, params)
	case "bilateral":
		return bilateral(v, params)
	case "bloom":
		return bloom(v, params)
	case "cacorrect":
		return cacorrect(v, params)
	case "channelmixer":
		return channelmixer(v, params)
	case "clahe":
		return clahe(v, params)
	case "clipping":
		return clipping(v, params)
	case "colisa":
		return colisa(v, params)
	case "colorbalance":
		return colorbalance(v, params)
	case "colorcontrast":
		return colorcontrast(v, params)
	case "colorcorrection":
		return colorcorrection(v, params)
	case "colorize":
		return colorize(v, params)
	case "colorzones":
		return colorzones(v, params)
	case "defringe":
		return defringe(v, params)
	case "demosaic":
		return demosaic(v, params)
	case "exposure":
		return exposure(v, params)
	case "filmic":
		return filmic(v, params)
	case "filmicrgb":
		return filmicrgb(v, params)
	case "flip":
		return flip(v, params)
	case "gamma":
		return gamma(v, params)
	case "graduatednd":
		return graduatednd(v, params)
	case "grain":
		return grain(v, params)
	case "hazeremoval":
		return hazeremoval(v, params)
	case "highlights":
		return highlights(v, params)
	case "highpass":
		return highpass(v, params)
	case "invert":
		return invert(v, params)
	case "levels":
		return levels(v, params)
	case "lowlight":
		return lowlight(v, params)
	case "lowpass":
		return lowpass(v, params)
	case "monochrome":
		return monochrome(v, params)
	case "nlmeans":
		return nlmeans(v, params)
	case "relight":
		return relight(v, params)
	case "shadhi":
		return shadhi(v, params)
	case "sharpen":
		return sharpen(v, params)
	case "soften":
		return soften(v, params)
	case "splittoning":
		return splittoning(v, params)
	case "tonemap":
		return tonemap(v, params)
	case "velvia":
		return velvia(v, params)
	case "vibrance":
		return vibrance(v, params)
	case "zonesystem":
		return zonesystem(v, params)

	}

	return nil, nil
}
