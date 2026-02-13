package avcamx

import (
	"fmt"
)

type AvStream struct {
	ID     int
	Url    string
	Config VideoConfig
	Source VideoSource `json:"-"`
	Server *AvServer   `json:"-"`
}

func NewAvStream(id int, config *VideoConfig, source VideoSource) (item *AvStream) {
	item = &AvStream{
		ID:     id,
		Url:    fmt.Sprintf("/video%d", id),
		Config: *config,
		Source: source,
	}
	return
}

func (item *AvStream) IsOpened() bool {
	if item.Source == nil {
		return false
	}
	return item.Source.IsOpened()
}

func (item *AvStream) IsRecording() bool {
	if item.Server == nil {
		return false
	}
	return item.Server.Recording
}

func (item *AvStream) RecordCmd(seconds int) {
	if item.Server != nil {
		item.Server.RecordCmd(seconds)
	}
}

func (item *AvStream) StopRecordCmd() {
	if item.Server != nil {
		item.Server.StopRecordCmd()
	}
}

func (item *AvStream) SetRecordListener(streamListener StreamListener) {
	if item.Server != nil {
		item.Server.Listener = streamListener
	}
}
