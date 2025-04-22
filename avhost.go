package avcam

import (
	"log"
	"net/http"
)

type AvHost struct {
	Url    string
	Items  []*AvItem
	server *http.Server
}

func NewAvHost(address string, port string) (host *AvHost) {
	host = &AvHost{
		Items: make([]*AvItem, 0),
	}
	if len(port) == 0 {
		port = "8080"
	}

	if len(address) == 0 {
		address = GetOutboundIP()
	}

	host.Url = address + ":" + port
	return
}

func (host *AvHost) Load() {
	var (
		err error
		mux = &http.ServeMux{}
	)

	host.server = &http.Server{
		Addr:    host.Url,
		Handler: mux,
	}

	webcams := FindWebcams()
	for id, webcam := range webcams {
		// requested configuration
		config := &VideoConfig{
			Codec:  "MJPG",
			Width:  1920,
			Height: 1080,
			FPS:    30,
		}

		err = webcam.Open(config)
		if err != nil {
			log.Print(err)
			continue
		}

		avItem := NewAvItem(id, config, webcam)
		host.Items = append(host.Items, avItem)

		avItem.server = NewVideoServer(id, webcam, &avItem.Config, nil, nil)
		mux.Handle(avItem.Url, avItem.server.Stream())
	}
}

func (host *AvHost) ListenAndServe() error {
	for _, avItem := range host.Items {
		go avItem.server.Serve()
	}
	return host.server.ListenAndServe()
}

func (host *AvHost) Quit() {
	for _, avItem := range host.Items {
		avItem.server.quit <- 1
		avItem.source.Close()
	}
}
