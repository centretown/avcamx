package avcamx

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/korandiz/v4l"
)

type AvHost struct {
	Url      string
	Streams  []*AvStream
	Remotes  []string
	Interval int
	Server   *http.Server       `json:"-"`
	tmpl     *template.Template `json:"-"`
	mux      *http.ServeMux     `json:"-"`
	cmd      chan int           `json:"-"`
	buf      chan []byte        `json:"-"`
}

const (
	AV_QUIT int = iota + 1
	AV_HOST
)

func NewAvHost(address string, port string, remotes []string, interval int) (host *AvHost) {
	host = &AvHost{
		Streams:  make([]*AvStream, 0),
		Remotes:  remotes,
		Interval: interval,
		mux:      &http.ServeMux{},
		cmd:      make(chan int),
		buf:      make(chan []byte),
	}
	if len(port) == 0 {
		port = "8080"
	}

	if len(address) == 0 {
		address = GetOutboundIP()
	}

	host.Url = address + ":" + port
	host.Server = &http.Server{
		Addr:    host.Url,
		Handler: host.mux,
	}
	var err error
	host.tmpl, err = template.New("response").Parse(`<div id="response-div" class="fade-it">{{.}}</div>`)
	if err != nil {
		log.Printf("NewAvHost Parse Template: %v", err)
	}

	host.mux.HandleFunc("/host", func(w http.ResponseWriter, r *http.Request) {
		buf, err := json.Marshal(host)
		if err != nil {
			buf = ([]byte)(err.Error())
			log.Printf("Handle '/host': %v", err)
		}
		_, err = w.Write(buf)
		if err != nil {
			log.Printf("Handle '/host': %v %s", err, string(buf))
		}

	})

	go func() {
		err := host.Server.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to serve host at: %v '%v'", host.Url, err)
		}
	}()

	go host.Monitor()
	return
}

func (host *AvHost) Monitor() {
	var (
		waitPeriod = time.Millisecond * time.Duration(host.Interval)
		nextScan   = time.Now()
	)

	for {
		select {
		case cmd := <-host.cmd:
			switch cmd {
			case AV_QUIT:
				log.Print("AvHost Monitor Done")
				return
			case AV_HOST:
			}
		default:
		}

		if nextScan.Compare(time.Now()) <= 0 {
			host.scanLocal()
			host.scanRemotes()
			nextScan.Add(waitPeriod)
			continue
		}

		time.Sleep(time.Millisecond * 1000)
	}
}

func (host *AvHost) Mux() *http.ServeMux { return host.mux }

func (host *AvHost) Quit() {
	host.cmd <- AV_QUIT
	for _, avStream := range host.Streams {
		if avStream.IsOpened() {
			log.Printf("Stopping '%s'\n", avStream.Source.Path())
			avStream.Server.Quit()
		}
	}
}

func (host *AvHost) scanLocal() {
	devices := v4l.FindDevices()
	for _, info := range devices {
		if !info.Camera {
			continue
		}
		if info.DriverName != UVCVideoDriver {
			continue
		}

		avStream := host.findAvStreamPath(info.Path)
		if avStream != nil {
			if avStream.Source.IsOpened() {
				continue
			}
			log.Printf("found path %v, %v driver %v", avStream.Url, info.Path, info.DriverName)
		} else {
			avStream = host.findAvStreamClosed()
			if avStream != nil {
				log.Printf("found closed %v, %v driver %v", avStream.Url, info.Path, info.DriverName)
			}
		}

		localcam := NewLocalCam(&info)
		config := &VideoConfig{
			Codec:  "MJPG",
			Width:  1920,
			Height: 1080,
			FPS:    30,
		}
		err := localcam.Open(config)
		if err != nil {
			log.Print("ScanLocal ", err)
			return
		}
		// avStream = NewAvStream(len(host.Streams), config, localcam)
		if avStream == nil {
			host.addStream(localcam, &localcam.videoConfig, nil, nil)
		} else {
			host.updateStream(avStream, localcam, &localcam.videoConfig)
		}

	}
	return
}

func (host *AvHost) scanRemotes() {
	// log.Print("REMOTES ", host.Remotes)
	for _, addr := range host.Remotes {
		remote, err := host.fetchRemote(addr)
		if err != nil {
			log.Printf("Fetching remote %s. %s", addr, err)
			continue
		}
		// log.Printf("Fetched remote %s. %v", addr, remote)

		for _, stream := range remote.Streams {
			streamAddr := addr + stream.Url
			avStream := host.findAvStreamPath(streamAddr)
			if avStream != nil {
				if avStream.Source.IsOpened() {
					continue
				}
				log.Printf("found remote %v, %v", avStream.Url, addr)
			} else {
				avStream = host.findAvStreamClosed()
				if avStream != nil {
					log.Printf("found remote closed %v, %v", avStream.Url, addr)
				}
			}

			remotecam := NewRemoteCam(streamAddr)
			err := remotecam.Open(&stream.Config)
			if err != nil {
				log.Print("ScanRemotes ", err)
				return
			}
			if avStream == nil {
				host.addStream(remotecam, &stream.Config, nil, nil)
			} else {
				host.updateStream(avStream, remotecam, &stream.Config)
			}
		}
	}
	return
}

func (host *AvHost) updateStream(avStream *AvStream,
	source VideoSource, config *VideoConfig) {
	avStream.Source = source
	avStream.Config = *config
	if avStream.Server == nil {
		log.Fatal("avStream.Server==nil")
	}

	avStream.Server.Source = source
	go avStream.Server.Serve()
	log.Printf("Updated stream %s -> %s", avStream.Url, avStream.Source.Path())
}

func (host *AvHost) addStream(
	source VideoSource, config *VideoConfig,
	audioSource AudioSource,
	listener StreamListener) (avStream *AvStream) {

	id := len(host.Streams)
	avStream = NewAvStream(id, config, source)
	avStream.Server = NewAvServer(id, source, &avStream.Config, nil, listener)
	host.Streams = append(host.Streams, avStream)
	go avStream.Server.Serve()
	host.createAvStreamHandlers(id, config.Driver)
	log.Printf("Added stream %s -> %s", avStream.Url, avStream.Source.Path())
	return
}

func (host *AvHost) createAvStreamHandlers(id int, driver string) {
	mux := host.mux
	avStream := host.Streams[id]
	host.mux.Handle(avStream.Url, avStream.Server.Stream())
	mux.HandleFunc(avStream.Url+"/",
		func(w http.ResponseWriter, r *http.Request) {
			url, _ := strings.CutPrefix(r.URL.Path, avStream.Url)
			switch avStream.Source.(type) {
			case *LocalCam:
				localcam := avStream.Source.(*LocalCam)
				if url == "/reset" {
					err := localcam.Reset()
					if err != nil {
						log.Println("AvStream Reset Handler: ", err, r.URL.Path)
					}
					return
				}

				ctrl, ok := AvUrlToName[url]
				if !ok {
					log.Println("Unsupported AvStream Request: ", r.URL.Path)
					return
				}

				info, ok := localcam.Controls[ctrl.Name]
				if !ok {
					log.Printf("Unsupported AvStream Control: %s '%s'",
						r.URL.Path, ctrl.Name)
					return
				}

				value, err := localcam.device.GetControl(info.CID)
				if err != nil {
					log.Println("Unsupported AvStream Control Value: ", r.URL.Path, err)
					return
				}

				v4lCtrl := localcam.Controls[ctrl.Name]
				newValue := value + v4lCtrl.Step*ctrl.Control.Multiplier
				if newValue >= v4lCtrl.Min && newValue <= v4lCtrl.Max {
					value = newValue
					err = localcam.device.SetControl(v4lCtrl.CID, value)
					if err != nil {
						log.Println("Set Control AvStream: ", r.URL.Path, err)
						return
					}
				}

			case *RemoteCam:
				resp, err := http.Get(avStream.Source.Path() + url)
				if err != nil {
					log.Println("Set Control AvStream: ", r.URL.Path, err)
					return
				}
				defer resp.Body.Close()
				buf, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Print(err)
					return
				}
				w.Write(buf)

			default:
				log.Println("Unsupported AvStream Requested: ", r.URL.Path)
				return
			}
		})
}

func (host *AvHost) findAvStreamPath(path string) (avStream *AvStream) {
	for _, avStream = range host.Streams {
		if avStream.Source.Path() == path {
			return
		}
	}
	return nil
}
func (host *AvHost) findAvStreamClosed() (avStream *AvStream) {
	for _, avStream = range host.Streams {
		if !avStream.IsOpened() {
			return
		}
	}
	return nil
}

func (host *AvHost) fetchRemote(remoteAddr string) (remote *AvHost, err error) {
	var (
		resp *http.Response
	)

	resp, err = http.Get(remoteAddr + "/host")
	if err != nil {
		log.Print("FetchRemote Get", err)
		return
	}

	remote, err = ReadRemote(resp.Body)
	if err != nil {
		log.Print("FetchRemote ReadRemote", err)
		return
	}
	return
}
