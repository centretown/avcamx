package avcamx

import (
	"fmt"
	"log"

	"github.com/korandiz/v4l"
)

type AvStream struct {
	ID         int
	Url        string
	DeviceName string
	Config     VideoConfig
	Configs    []v4l.DeviceConfig
	Controls   []v4l.ControlInfo
	Source     VideoSource `json:"-"`
	Server     *AvServer   `json:"-"`
}

func NewAvStream(id int, config *VideoConfig, source VideoSource) (stream *AvStream) {
	stream = &AvStream{
		ID:       id,
		Url:      fmt.Sprintf("/video%d", id),
		Config:   *config,
		Source:   source,
		Configs:  make([]v4l.DeviceConfig, 0),
		Controls: make([]v4l.ControlInfo, 0),
	}
	stream.copyConfigs()
	return
}

func (stream *AvStream) copyConfigs() {
	if local, ok := stream.Source.(*LocalCam); ok {
		info, err := local.DeviceInfo()
		if err == nil {
			stream.DeviceName = info.DeviceName
		} else {
			log.Println(err)
		}

		configs, err := local.device.ListConfigs()
		if err == nil {
			stream.Configs = make([]v4l.DeviceConfig, len(configs))
			copy(stream.Configs, configs)
		} else {
			log.Println(err)
		}
		controls, err := local.device.ListControls()
		if err == nil {
			stream.Controls = make([]v4l.ControlInfo, len(controls))
			copy(stream.Controls, controls)
		} else {
			log.Println(err)
		}
	}
}

func (stream *AvStream) copyStream() (s *AvStream) {
	s = &AvStream{
		ID:         stream.ID,
		Url:        stream.Url,
		Config:     stream.Config,
		Source:     stream.Source,
		Server:     stream.Server,
		DeviceName: stream.DeviceName,
		Configs:    make([]v4l.DeviceConfig, len(stream.Configs)),
		Controls:   make([]v4l.ControlInfo, len(stream.Controls)),
	}
	copy(s.Configs, stream.Configs)
	copy(s.Controls, stream.Controls)
	return
}

func (stream *AvStream) IsOpened() bool {
	if stream.Source == nil {
		return false
	}
	return stream.Source.IsOpened()
}

func (stream *AvStream) IsRecording() bool {
	if stream.Server == nil {
		return false
	}
	return stream.Server.Recording
}

func (stream *AvStream) RecordCmd(seconds int) {
	if stream.Server == nil {
		log.Print("RecordCmd No server")
		return
	}
	stream.Server.RecordCmd(seconds)
}

func (stream *AvStream) StopRecordCmd() {
	if stream.Server == nil {
		log.Print("StopRecordCmd No server")
		return
	}
	stream.Server.StopRecordCmd()
}

func (stream *AvStream) SetRecordListener(streamListener StreamListener) {
	if stream.Server != nil {
		stream.Server.Listener = streamListener
	}
}
