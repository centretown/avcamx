package avcamx

type AvControl struct {
	Url        string
	Icon       string
	Multiplier int32
}

type AvUrl struct {
	Name    string
	Control *AvControl
}

var AvUrlToName map[string]AvUrl = make(map[string]AvUrl)

func init() {
	for ctrlName, avCtrlList := range UCVVIDEO {
		for _, ctrl := range avCtrlList {
			AvUrlToName[ctrl.Url] = AvUrl{Name: ctrlName, Control: &ctrl}
		}
	}
}

// keys correspond to v4l Control Names
var UCVVIDEO = map[string][]AvControl{
	"Zoom, Absolute": {
		AvControl{
			Url:        "/zoomin",
			Icon:       "zoom_in",
			Multiplier: 1,
		},
		AvControl{
			Url:        "/zoomout",
			Icon:       "zoom_out",
			Multiplier: -1,
		},
	},
	"Pan, Absolute": {
		AvControl{
			Url:        "/panleft",
			Icon:       "arrow_back",
			Multiplier: -1,
		},
		AvControl{
			Url:        "/panright",
			Icon:       "arrow_forward",
			Multiplier: 1,
		},
	},
	"Tilt, Absolute": {
		AvControl{
			Url:        "/tiltup",
			Icon:       "arrow_upward",
			Multiplier: 1,
		},
		AvControl{
			Url:        "/tiltdown",
			Icon:       "arrow_downward",
			Multiplier: -1,
		},
	},
	"Brightness": {
		AvControl{
			Url:        "/brightnessup",
			Icon:       "brightness_high",
			Multiplier: 10,
		},
		AvControl{
			Url:        "/brightnessdown",
			Icon:       "brightness_low",
			Multiplier: -10,
		},
	},
	"Contrast": {
		AvControl{
			Url:        "/contrastup",
			Icon:       "contrast_square",
			Multiplier: 10,
		},
		AvControl{
			Url:        "/contrastdown",
			Icon:       "exposure",
			Multiplier: -10,
		},
	},
	"Saturation": {
		AvControl{
			Url:        "/saturationup",
			Icon:       "backlight_high",
			Multiplier: 10,
		},
		AvControl{
			Url:        "/saturationdown",
			Icon:       "backlight_low",
			Multiplier: -10,
		},
	},
}

var AvControllers = map[string]map[string][]AvControl{
	"uvcvideo": UCVVIDEO,
}
