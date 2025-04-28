package avcamx

import "fmt"

type AvItem struct {
	ID     int
	Url    string
	Config VideoConfig
	source VideoSource
	server *Server
}

func NewAvItem(id int, config *VideoConfig, source VideoSource) (item *AvItem) {
	item = &AvItem{
		ID:     id,
		Url:    fmt.Sprintf("/video%d", id),
		Config: *config,
		source: source,
	}
	return
}

func (item *AvItem) IsOpened() bool {
	if item.source == nil {
		return false
	}
	return item.source.IsOpened()
}

func (item *AvItem) IsRecording() bool {
	if item.server == nil {
		return false
	}
	return item.server.Recording
}

func (item *AvItem) RecordCmd(seconds int) {
	if item.server != nil {
		item.server.RecordCmd(seconds)
	}
}

func (item *AvItem) StopRecordCmd() {
	if item.server != nil {
		item.server.StopRecordCmd()
	}
}

func (item *AvItem) SetRecordListener(streamListener StreamListener) {
	if item.server != nil {
		item.server.Listener = streamListener
	}
}
