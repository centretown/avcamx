package avcamx

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/korandiz/v4l"
)

type AvHost struct {
	Url    string
	Items  []*AvItem
	server *http.Server
	tmpl   *template.Template
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

	host.server = &http.Server{
		Addr:    host.Url,
		Handler: &http.ServeMux{},
	}

	host.tmpl, _ = template.New("response").Parse(`{{ define "layout.response" }}
<div id="response-div" class="fade-it">{{.}}</div>
{{ end }}`)

	return
}

func (host *AvHost) MakeLocal() {
	var (
		err error
		mux = host.server.Handler.(*http.ServeMux)
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(host)
		if err != nil {
			buf = ([]byte)(err.Error())
		}
		w.Write(buf)
	})

	webcams := FindWebcams()
	for id, webcam := range webcams {

		// requested configuration, actual configuration determined
		// when opened depending on what's available for that camera
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
				if _, keyFound := controller[key]; keyFound {
					webcam.device.SetControl(v4lCtrl.CID, v4lCtrl.Default)
				}
			}
		})

		for key, v4lCtrl := range webcam.Controls {
			if avCtrls, keyFound := controller[key]; keyFound {
				for _, avCtrl := range avCtrls {
					mux.HandleFunc(avItem.Url+avCtrl.Url,
						host.LocalHandler(webcam, v4lCtrl, avCtrl))
				}
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

func (host *AvHost) FetchRemote(remoteAddr string) (remote *AvHost, err error) {
	var (
		resp *http.Response
	)

	resp, err = http.Get(remoteAddr)
	if err != nil {
		log.Print(err)
		return
	}

	remote, err = ReadRemote(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	return
}

func (host *AvHost) MakeProxy(remote *AvHost) {
	var (
		err error
		id  = len(host.Items)
		mux = host.server.Handler.(*http.ServeMux)
	)

	for index, remoteItem := range remote.Items {
		var (
			remoteItemUrl = "http://" + remote.Url + remoteItem.Url
			config        = remoteItem.Config
			ipcam         = NewIpcam(remoteItemUrl)
			avItem        = NewAvItem(id+index, &config, ipcam)
		)

		config.Path = remoteItem.Url
		err = ipcam.Open(&avItem.Config)
		if err != nil {
			log.Print(err)
			continue
		}

		avItem.server = NewVideoServer(id, ipcam, &avItem.Config, nil, nil)
		mux.Handle(avItem.Url, avItem.server.Stream())
		mux.HandleFunc(avItem.Url+"/reset",
			host.RemoteHandler(remoteItemUrl, "/reset"))

		controller := AvControllers[config.Driver]
		for _, controls := range controller {
			for _, control := range controls {
				mux.HandleFunc(avItem.Url+control.Url,
					host.RemoteHandler(remoteItemUrl, control.Url))
			}
		}

		host.Items = append(host.Items, avItem)
		log.Printf("added proxy at %s%s\n\tfor remote %s\n",
			host.Url, avItem.Url, remoteItemUrl)
	}
	return
}

func (host *AvHost) RemoteHandler(remoteItemUrl string, command string) func(http.ResponseWriter, *http.Request) {
	return func(http.ResponseWriter, *http.Request) {
		resp, err := http.Get(remoteItemUrl + command)
		if err != nil {
			log.Print(err)
			return
		}
		defer resp.Body.Close()
		var buf []byte
		buf, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("Received '%s' from remote: %s command: %s\n",
			string(buf), remoteItemUrl, command)
	}
}

func (host *AvHost) LocalHandler(webcam *Webcam, v4lCtrl v4l.ControlInfo, avCtrl AvControl) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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
		host.tmpl.Execute(w, value)
	}
}
