package avcam

import (
	"github.com/mattn/go-mjpeg"
)

type StreamHook struct {
	Stream *mjpeg.Stream
}

func NewStreamHook() *StreamHook {
	sh := &StreamHook{}
	sh.Stream = mjpeg.NewStream()
	return sh
}

func (sh *StreamHook) Update(img []byte) {
	// var dst = make([]byte, img.Len())
	// n, err := img.Read(dst)
	// if err != nil {
	// 	log.Println("StreamHook Update", err)
	// 	return
	// }
	// log.Println("StreamHook Update", n, "bytes read")

	sh.Stream.Update(img)
}

func (sh *StreamHook) Close(int) {}
