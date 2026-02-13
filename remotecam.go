package avcamx

import (
	"log"

	"github.com/mattn/go-mjpeg"
)

var _ VideoSource = (*RemoteCam)(nil)

type RemoteCam struct {
	path     string
	config   *VideoConfig
	decoder  *mjpeg.Decoder
	Buffer   []byte
	isOpened bool
	State    any
}

func NewRemoteCam(path string) *RemoteCam {
	ipc := &RemoteCam{
		path: path,
	}
	return ipc
}

func (ipc *RemoteCam) Path() string {
	return ipc.path
}

func (ipc *RemoteCam) Config() *VideoConfig {
	return ipc.config
}

func (ipc *RemoteCam) Close() {
	ipc.isOpened = false
}

func (ipc *RemoteCam) IsOpened() bool {
	return ipc.isOpened
}

func (ipc *RemoteCam) Open(config *VideoConfig) (err error) {
	ipc.config = config
	ipc.decoder, err = mjpeg.NewDecoderFromURL(ipc.path)
	if err != nil {
		log.Println("NewDecoderFromURL", err)
		ipc.isOpened = false
	} else {
		ipc.isOpened = true
	}
	return
}

func (ipc *RemoteCam) Read() (buf []byte, err error) {
	buf, err = ipc.decoder.DecodeRaw()
	if err != nil {
		log.Println("DecodeRaw", err)
	}

	return
}
