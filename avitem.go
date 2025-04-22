package avcam

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
