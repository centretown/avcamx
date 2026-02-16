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
	Url            string
	Streamers      []*AvStream
	Remotes        []string
	Interval       int
	Server         *http.Server       `json:"-"`
	streamListener StreamListener     `json:"-"`
	tmpl           *template.Template `json:"-"`
	mux            *http.ServeMux     `json:"-"`
	cmdChan        chan int           `json:"-"`
	streamsChan    chan []*AvStream   `json:"-"`
	urlChan        chan string        `json:"-"`
	streamChan     chan *AvStream     `json:"-"`
}

const (
	AV_QUIT int = iota + 1
	AV_STREAMS
	AV_URL
)

func NewAvHost(address string, port string, remotes []string, interval int, streamListener StreamListener) (host *AvHost) {
	host = &AvHost{
		Streamers:      make([]*AvStream, 0),
		Remotes:        remotes,
		Interval:       interval,
		streamListener: streamListener,
		mux:            &http.ServeMux{},
		cmdChan:        make(chan int),
		streamsChan:    make(chan []*AvStream),
		urlChan:        make(chan string),
		streamChan:     make(chan *AvStream),
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
		copy := &AvHost{
			Url:       host.Url,
			Streamers: host.Streams(),
			Remotes:   host.Remotes,
			Interval:  host.Interval,
		}
		buf, err := json.Marshal(copy)
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
			log.Printf("Host at: %v '%v'", host.Url, err)
		}
	}()

	go host.Monitor()
	return
}

func (host *AvHost) Stream(url string) (stream *AvStream) {
	host.urlChan <- url
	stream = <-host.streamChan
	return
}

func (host *AvHost) Streams() (streams []*AvStream) {
	host.cmdChan <- AV_STREAMS
	streams = <-host.streamsChan
	return
}

func (host *AvHost) findStream(url string) *AvStream {
	for _, s := range host.Streamers {
		if s.Url == url {
			return &AvStream{
				ID:     s.ID,
				Url:    s.Url,
				Config: s.Config,
				Source: s.Source,
				Server: s.Server,
			}
		}
	}
	return nil
}

func (host *AvHost) copyStreams() (streams []*AvStream) {
	streams = make([]*AvStream, len(host.Streamers))
	for i, stream := range host.Streamers {
		copy := &AvStream{
			ID:     stream.ID,
			Url:    stream.Url,
			Config: stream.Config,
			Source: stream.Source,
			Server: stream.Server,
		}
		streams[i] = copy
	}
	return
}

func (host *AvHost) Monitor() {
	var (
		waitPeriod = time.Millisecond * time.Duration(host.Interval)
		nextScan   = time.Now()
		now        time.Time
	)

	for {
		now = time.Now()
		if nextScan.Compare(now) <= 0 {
			host.scanLocal()
			host.scanRemotes()
			nextScan = now.Add(waitPeriod)
			// continue
		}

		select {
		case cmd := <-host.cmdChan:
			switch cmd {
			case AV_QUIT:
				log.Print("AvHost Monitor Done")
				return
			case AV_STREAMS:
				host.streamsChan <- host.copyStreams()
			}
		case url := <-host.urlChan:
			host.streamChan <- host.findStream(url)
		default:
			time.Sleep(time.Millisecond * 10)
		}

	}
}

func (host *AvHost) Mux() *http.ServeMux { return host.mux }

func (host *AvHost) Quit() {
	for _, avStream := range host.Streamers {
		if avStream.IsOpened() {
			log.Printf("Stopping '%s'\n", avStream.Source.Path())
			avStream.Server.Quit()
		}
	}
	host.cmdChan <- AV_QUIT
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
			host.addStream(localcam, &localcam.videoConfig, nil, host.streamListener)
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

		for _, stream := range remote.Streamers {
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
				host.addStream(remotecam, &stream.Config, nil, host.streamListener)
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

	id := len(host.Streamers)
	avStream = NewAvStream(id, config, source)
	avStream.Server = NewAvServer(id, source, &avStream.Config, nil, listener)
	host.Streamers = append(host.Streamers, avStream)
	go avStream.Server.Serve()
	host.createAvStreamHandlers(id, config.Driver)
	log.Printf("Added stream %s -> %s", avStream.Url, avStream.Source.Path())
	return
}

func (host *AvHost) createAvStreamHandlers(id int, driver string) {
	mux := host.mux
	avStream := host.Streamers[id]
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
						host.tmpl.Execute(w, err.Error())
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
					host.tmpl.Execute(w, err.Error())
					return
				}

				v4lCtrl := localcam.Controls[ctrl.Name]
				newValue := value + v4lCtrl.Step*ctrl.Control.Multiplier
				if newValue >= v4lCtrl.Min && newValue <= v4lCtrl.Max {
					value = newValue
					err = localcam.device.SetControl(v4lCtrl.CID, value)
					if err != nil {
						log.Println("Set Control AvStream: ", r.URL.Path, err)
						// w.Write(([]byte)(err.Error()))
						host.tmpl.Execute(w, err.Error())
						return
					}
				}
				host.tmpl.Execute(w, value)

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
	for _, avStream = range host.Streamers {
		if avStream.Source.Path() == path {
			return
		}
	}
	return nil
}
func (host *AvHost) findAvStreamClosed() (avStream *AvStream) {
	for _, avStream = range host.Streamers {
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
