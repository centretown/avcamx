package avcamx

import (
	"log"

	"github.com/mattn/go-mjpeg"
)

var _ VideoSource = (*Ipcam)(nil)

type Ipcam struct {
	path     string
	config   *VideoConfig
	decoder  *mjpeg.Decoder
	Buffer   []byte
	isOpened bool
	State    any
}

func NewIpcam(path string) *Ipcam {
	ipc := &Ipcam{
		path: path,
	}
	return ipc
}

func (ipc *Ipcam) Path() string {
	return ipc.path
}

func (ipc *Ipcam) Config() *VideoConfig {
	return ipc.config
}

func (ipc *Ipcam) Close() {
	ipc.isOpened = false
}

func (ipc *Ipcam) IsOpened() bool {
	return ipc.isOpened
}

func (ipc *Ipcam) Open(config *VideoConfig) (err error) {
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

func (ipc *Ipcam) Read() (buf []byte, err error) {
	buf, err = ipc.decoder.DecodeRaw()
	if err != nil {
		log.Println("DecodeRaw", err)
	}

	return
}
