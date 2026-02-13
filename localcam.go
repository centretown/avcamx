package avcamx

import (
	"fmt"
	"log"
	"strings"

	"github.com/korandiz/v4l"
)

var _ VideoSource = (*LocalCam)(nil)

const UVCVideoDriver = "uvcvideo"

type LocalCam struct {
	Info        v4l.DeviceInfo
	device      v4l.Device
	Controls    map[string]v4l.ControlInfo
	Buffer      []byte
	videoConfig VideoConfig
	isOpened    bool
	readOnly    bool
}

func NewLocalCam(info *v4l.DeviceInfo) *LocalCam {
	cam := &LocalCam{
		Info:     *info,
		Buffer:   make([]byte, 0),
		Controls: make(map[string]v4l.ControlInfo, 0),
	}
	return cam
}

func FindLocalCams() (ws map[string]*LocalCam) {
	ws = make(map[string]*LocalCam, 0)
	list := v4l.FindDevices()
	for _, info := range list {
		if info.Camera && info.DriverName == UVCVideoDriver {
			ws[info.Path] = NewLocalCam(&info)
		}
	}
	return
}

func (cam *LocalCam) Reset() error {
	for _, control := range cam.Controls {
		val, err := cam.device.GetControl(control.CID)
		if err != nil {
			log.Printf("LocalCam Reset GetControl: %v, '%s', ==%d def %d, min %d, max %d, step %d",
				err, control.Name, val,
				control.Default, control.Min, control.Max, control.Step)
			continue
		}
		if val == control.Default {
			continue
		}

		err = cam.device.SetControl(control.CID, control.Default)
		if err != nil {
			log.Printf("LocalCam Reset SetControl: %v, '%s', ==%d def %d, min %d, max %d, step %d",
				err, control.Name, val,
				control.Default, control.Min, control.Max, control.Step)
			// return err
		}
	}
	return nil
}

func (cam *LocalCam) Path() string {
	return cam.Info.Path
}

func (cam *LocalCam) DeviceInfo() (info v4l.DeviceInfo, err error) {
	info = cam.Info
	return info, err
}

// if already in use becomes read only
func (cam *LocalCam) Open(videoConfig *VideoConfig) (err error) {
	var device *v4l.Device
	device, err = v4l.Open(cam.Info.Path)
	if err != nil {
		log.Println("Open", cam.Info.Path, err)
		return
	}

	cam.device = *device
	cam.isOpened = true
	deviceInfo, err := cam.device.DeviceInfo()
	if err != nil {
		log.Println("DeviceInfo", cam.Info.Path, err)
		cam.Close()
		return
	}

	err = cam.mapControls()
	if err != nil {
		return
	}

	cam.device.TurnOff()

	preferred := &v4l.DeviceConfig{
		Format: ToFourCC(videoConfig.Codec),
		Width:  videoConfig.Width, Height: videoConfig.Height,
		FPS: v4l.Frac{N: videoConfig.FPS, D: 1},
	}

	found := cam.findConfig(preferred)
	cam.videoConfig.Path = cam.Info.Path
	cam.videoConfig.Driver = deviceInfo.DriverName
	cam.videoConfig.Codec = FourCC(found.Format)
	cam.videoConfig.Width = found.Width
	cam.videoConfig.Height = found.Height
	cam.videoConfig.FPS = found.FPS.N

	err = cam.device.SetConfig(*found)
	if err != nil {
		cam.readOnly = true
		log.Println("readonly: ", err)
		return
	}

	var bufferInfo v4l.BufferInfo
	if bufferInfo, err = cam.device.BufferInfo(); err != nil {
		err = fmt.Errorf("bufferInfo %v", err)
		return
	}

	cam.Buffer = make([]byte, bufferInfo.BufferSize)

	if !cam.readOnly {
		if err = cam.device.TurnOn(); err != nil {
			err = fmt.Errorf("turn on %v", err)
			return
		}
	}

	return nil
}

func (cam *LocalCam) mapControls() (err error) {
	var controls []v4l.ControlInfo
	controls, err = cam.device.ListControls()
	if err != nil {
		log.Println("ListControls", cam.Info.Path, err)
		return
	}

	for _, control := range controls {
		cam.Controls[control.Name] = control
	}
	return
}

func (cam *LocalCam) GetControlInfo(key string) (info v4l.ControlInfo, err error) {
	var ok bool
	if info, ok = cam.Controls[strings.ToLower(key)]; !ok {
		err = fmt.Errorf("unknown control %s", key)
	}
	return
}

func (cam *LocalCam) GetControlValue(key string) (value int32) {
	control, ok := cam.Controls[strings.ToLower(key)]
	if !ok {
		log.Println("unknown control", key, value)
		return
	}

	value, err := cam.device.GetControl(control.CID)
	if err != nil {
		log.Println("GetControl", key, value, err)
		return
	}

	return
}

func (cam *LocalCam) SetControlValue(key string, value int32) {
	control, ok := cam.Controls[strings.ToLower(key)]
	if !ok {
		log.Println("unknown control", key, value)
		return
	}

	err := cam.device.SetControl(control.CID, value)
	if err != nil {
		log.Println("SetControl", key, value, err)
		return
	}

	log.Println("SetControl", key, value)
}

func (cam *LocalCam) IsOpened() bool {
	return cam.isOpened
}

func (cam *LocalCam) Close() {
	cam.device.TurnOff()
	cam.device.Close()
	cam.isOpened = false
}

func (cam *LocalCam) Read() (buf []byte, err error) {
	buf = cam.Buffer
	var (
		vbuf  *v4l.Buffer
		count int
	)
	vbuf, err = cam.device.Capture()
	if err != nil {
		log.Println("Webcam Capture", err)
		return
	}

	count, err = vbuf.Read(buf)
	if err != nil {
		log.Println("Webcam Read", err)
		return
	}
	// log.Println(count, "bytes read")
	buf = buf[:count]
	return
}

func (cam *LocalCam) findConfig(b *v4l.DeviceConfig) *v4l.DeviceConfig {
	var (
		selected int
		lowest   int = 1_000_000
		score    int
		configs  []v4l.DeviceConfig
		err      error
	)

	configs, err = cam.device.ListConfigs()
	if err != nil {
		log.Println("ListConfigs", err)
		return nil
	}

	for i := range configs {
		score = scoreConfig(b, &configs[i])
		if score < lowest {
			selected = i
			lowest = score
		}
	}
	// fmt.Println("lowest", lowest, "selected", selected)
	return &configs[selected]
}

func scoreConfig(a, b *v4l.DeviceConfig) (score int) {
	abs := func(a int) int {
		if a < 0 {
			return -a
		}
		return a
	}
	if a.Format != b.Format {
		score += 100
	}
	if a.Width != b.Width {
		score += abs(a.Width - b.Width)
	}
	if a.Height != b.Height {
		score += abs(a.Height - b.Height)
	}
	if a.FPS != b.FPS {
		score += abs(int(a.FPS.N) - int(b.FPS.N))
	}
	return
}
