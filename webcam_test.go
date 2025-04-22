package avcam

import "testing"

func TestFind(t *testing.T) {
	webcams := FindWebcams()
	for i, webcam := range webcams {
		t.Log(i, webcam.Path())
	}
}
