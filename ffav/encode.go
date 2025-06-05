package ffav

// #cgo pkg-config: --define-variable=prefix=. libavcodec
// #cgo pkg-config: --define-variable=prefix=. libavutil
// #include "encode_video.h"
// #include <stdlib.h>
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

var (
	fcodec = C.CString("libx264")
)

type EncodeImpl struct {
	ch       chan int
	Active   bool
	Filename string
	Data     *C.EncodeData
}

func EncodeStep(filename string, ch chan int) {
	fname := C.CString(filename)
	data := C.encode_init(fcodec, fname)
	rc := int(data.ret)
	if rc != 0 {
		log.Fatal("Failed To Initialize Encode", rc)
	}

	defer func() {
		C.free(unsafe.Pointer(fname))
		C.encode_free(data)
		ch <- 1
	}()

	for i := range 200 {
		C.encode_write_data(data, C.int(i))
		rc := int(data.ret)
		if rc != 0 {
			log.Fatal("Failed To Write Data")
		}
	}

}

func EncodeGo() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go EncodeFile("testgoback5.mp4", ch1)
	go EncodeFile("testgoback6.mp4", ch2)
	<-ch1
	<-ch2
	C.free(unsafe.Pointer(fcodec))
}

func EncodeFile(filename string, ch chan int) {
	fname := C.CString(filename)
	status := C.encode_file(fname, fcodec)
	if status == C.SUCCESS {
		log.Println("SUCCESS")
	}
	fmt.Printf("\n\nencode_file=%d\n\n", int(status))
	C.free(unsafe.Pointer(fname))
	ch <- 1
}

func FillData(index int, width int, height int, bufY []byte, bufCb []byte, bufCr []byte) {
	for y := range height {
		for x := range width {
			i := y*width + x
			bufY[i] = byte(x + y + index*3)
			bufCb[i] = byte(128 + y + index*2)
			bufCr[i] = byte(64 + x + index*5)
		}
	}
	return
}

func EncodeDraw(filename string, ch chan int) {
	fname := C.CString(filename)
	data := C.encode_init(fcodec, fname)
	rc := int(data.ret)
	if rc != 0 {
		log.Fatal("Failed To Initialize Encode", rc)
	}

	defer func() {
		C.encode_free(data)
		C.free(unsafe.Pointer(fname))
		ch <- 1
	}()

	var (
		width  = int(data.frame.width)
		height = int(data.frame.height)
		bufY   = make([]byte, width*height)
		bufCr  = make([]byte, width*height)
		bufCb  = make([]byte, width*height)
	)

	for i := range 200 {
		FillData(i, width, height, bufY, bufCb, bufCr)
		cbufY := C.CBytes(bufY)
		cbufCb := C.CBytes(bufCb)
		cbufCr := C.CBytes(bufCr)
		C.encode_buffer(data, C.int(i), (*C.uchar)(cbufY),
			(*C.uchar)(cbufCb), (*C.uchar)(cbufCr))
		C.free(cbufY)
		C.free(cbufCb)
		C.free(cbufCr)
	}

}
