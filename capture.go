package avcam

import (
	"fmt"
	"io"
	"log"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func Capture(stop <-chan int, img <-chan []byte,
	width, height int, fps uint32) {

	log.Println("CaptureVideo")
	var (
		reader, writer = io.Pipe()
		err            error
		fpss           = fmt.Sprintf("%d", fps)
		// ts             = fmt.Sprintf("%.3f", duration)
	)

	defer func() {
		if err != nil {
			log.Fatal(err)
		}
	}()
	fname, _ := NextFileName(OutputBase, "mp4")
	done := make(chan error)
	go func() {
		err = ffmpeg.
			Input("pipe:",
				ffmpeg.KwArgs{
					"format":    "jpeg_pipe",
					"pix_fmt":   "yuv420p",
					"framerate": fpss,
					"s":         fmt.Sprintf("%dx%d", width, height),
				}).
			Output(fname,
				ffmpeg.KwArgs{
					"pix_fmt": "yuv420p",
					"vf":      "scale=1280:-1",
					"vsync":   "1",
					// "vf":    "scale=trunc(iw/2)*2:trunc(ih/2)*2",
					// "t":         ts,
				}).
			OverWriteOutput().
			WithInput(reader).
			Run()
		log.Println("ffmpeg process2 done")
		done <- err
		close(done)
	}()

	go write(stop, img, writer)
	log.Println("Starting ffmpeg process2")

}

func write(done <-chan int, imgCh <-chan []byte, writer io.WriteCloser) {

	var (
		count      int
		byteCount  int
		frameCount int
		err        error
		// pixels     []byte = make([]byte, width*height*COLOR_WIDTH)
	)

	defer func() {
		if err != nil {
			log.Fatal(err)
		}
	}()

	writePixels := func(pixels []byte) error {
		count, err = writer.Write(pixels)
		if err != nil {
			return err
		}
		byteCount += count
		frameCount++
		return nil
	}

	var buf []byte

	time.Sleep(time.Second * 1)

	for {
		// time.Sleep(time.Millisecond * 2)
		select {
		case buf = <-imgCh:
			err = writePixels(buf)
			if err != nil {
				log.Println("FFMPEG write", err)
				return
			}
			// log.Println("FFMPEG", len(buf))

		case <-done:
			err = writer.Close()
			log.Println("FFMPEG done", frameCount, byteCount)
			return
		}
	}
}

// func NextFileName(ext string) string {
// 	var namePrefix = "capture"
// 	id := uuid.New()
// 	name := fmt.Sprintf("%s_%s.%s", namePrefix, id.String(), ext)
// 	return name
// }
