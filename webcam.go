package avcam

import (
	"fmt"
	"log"
	"strings"

	"github.com/korandiz/v4l"
)

var _ VideoSource = (*Webcam)(nil)

type Webcam struct {
	path     string
	device   *v4l.Device
	Controls map[string]v4l.ControlInfo
	isOpened bool
	readOnly bool
	Buffer   []byte
}

func NewWebcam(path string) *Webcam {
	cam := &Webcam{
		path:     path,
		Buffer:   make([]byte, 0),
		Controls: make(map[string]v4l.ControlInfo, 0),
	}
	return cam
}

func FindWebcams() (ws []*Webcam) {
	ws = make([]*Webcam, 0)
	list := v4l.FindDevices()
	for _, info := range list {
		ws = append(ws, NewWebcam(info.Path))
	}
	return
}

func (cam *Webcam) Path() string {
	return cam.path
}

func (cam *Webcam) DeviceInfo() (info v4l.DeviceInfo, err error) {
	if cam.device == nil {
		err = fmt.Errorf("webcam not opened")
		return
	}
	info, err = cam.device.DeviceInfo()
	return
}

// if already in use becomes read only
func (cam *Webcam) Open(videoConfig *VideoConfig) (err error) {

	cam.isOpened = false
	cam.device, err = v4l.Open(cam.path)
	if err != nil {
		log.Println("Open", cam.path, err)
		return err
	}

	cam.isOpened = true
	deviceInfo, err := cam.device.DeviceInfo()
	if err != nil {
		log.Println("DeviceInfo", cam.path, err)
		return err
	}

	log.Printf("DeviceName:'%s' DriverName: '%s'\n",
		deviceInfo.DeviceName, deviceInfo.DriverName)

	cam.listControls()
	cam.device.TurnOff()

	preferred := &v4l.DeviceConfig{
		Format: ToFourCC(videoConfig.Codec),
		Width:  videoConfig.Width, Height: videoConfig.Height,
		FPS: v4l.Frac{N: videoConfig.FPS, D: 1},
	}

	configErr := cam.device.SetConfig(*cam.findConfig(preferred))
	cam.readOnly = configErr != nil

	var bufferInfo v4l.BufferInfo
	if bufferInfo, err = cam.device.BufferInfo(); err != nil {
		err = fmt.Errorf("bufferInfo %v", err)
		return
	}

	log.Printf("buffer size %x\n", bufferInfo.BufferSize)
	cam.Buffer = make([]byte, bufferInfo.BufferSize)

	if !cam.readOnly {
		if err = cam.device.TurnOn(); err != nil {
			err = fmt.Errorf("turn on %v", err)
			return
		}
	}
	return
}

func (cam *Webcam) listControls() {
	controls, err := cam.device.ListControls()
	if err != nil {
		log.Println("ListControls", cam.path, err)
		return
	}

	log.Println("Controls:")
	for _, c := range controls {
		cam.Controls[strings.ToLower(c.Name)] = c
		val, _ := cam.device.GetControl(c.CID)
		log.Printf("CID='%v', Name='%s', Type=%v, Default=%v, Max=%v, Min=%v, Step=%v Value=%v\n",
			c.CID, c.Name, c.Type, c.Default, c.Max, c.Min, c.Step, val)
	}
}

func (cam *Webcam) GetControlInfo(key string) (info v4l.ControlInfo, err error) {
	var ok bool
	if info, ok = cam.Controls[strings.ToLower(key)]; !ok {
		err = fmt.Errorf("unknown control %s", key)
	}
	return
}

func (cam *Webcam) GetControlValue(key string) (value int32) {
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

func (cam *Webcam) SetControlValue(key string, value int32) {
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

func (cam *Webcam) IsOpened() bool {
	return cam.isOpened
}

func (cam *Webcam) Close() {
	cam.device.TurnOff()
	cam.device.Close()
	cam.isOpened = false
}

func (cam *Webcam) Read() (buf []byte, err error) {
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

func (cam *Webcam) findConfig(b *v4l.DeviceConfig) *v4l.DeviceConfig {
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
