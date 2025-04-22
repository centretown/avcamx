package avcamx

import (
	"encoding/json"
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(host)
		if err != nil {
			buf = ([]byte)(err.Error())
		}
		w.Write(buf)
	})

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
		avItem.server = NewVideoServer(id, webcam, &avItem.Config, nil, nil)
		mux.Handle(avItem.Url, avItem.server.Stream())

		controller := AvControllers[config.Driver]
		mux.HandleFunc(avItem.Url+"/reset", func(http.ResponseWriter, *http.Request) {
			for key, v4lCtrl := range webcam.Controls {
				_, keyFound := controller[key]
				if !keyFound {
					continue
				}
				webcam.device.SetControl(v4lCtrl.CID, v4lCtrl.Default)
			}
		})

		for key, v4lCtrl := range webcam.Controls {
			avCtrls, keyFound := controller[key]
			if !keyFound {
				continue
			}

			for _, avCtrl := range avCtrls {
				mux.HandleFunc(avItem.Url+avCtrl.Url, func(http.ResponseWriter, *http.Request) {
					value, err := webcam.device.GetControl(v4lCtrl.CID)
					if err != nil {
						log.Print(err)
						return
					}

					newValue := value + v4lCtrl.Step*avCtrl.Multiplier
					if newValue >= v4lCtrl.Min && newValue <= v4lCtrl.Max {
						value = newValue
						webcam.device.SetControl(v4lCtrl.CID, value)
					}
				})
			}

		}

		host.Items = append(host.Items, avItem)
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
		log.Printf("Stopping '%s'\n", avItem.source.Path())
		avItem.server.Quit()
	}
}
