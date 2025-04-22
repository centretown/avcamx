package avcam

import (
	"strings"
	"testing"
)

func TestFourcc(t *testing.T) {
	test4cc(t, 1196444237)
	test4cc(t, 1448695129)

	test24cc(t, "mjpg")
	test24cc(t, "yuyv")
}

func test4cc(t *testing.T, fourcc uint32) {
	s := FourCC(fourcc)
	t.Log(fourcc, s)
	f := ToFourCC(s)
	if f != fourcc {
		t.Fatal("f!=fourcc", f, fourcc)
	}
}

func test24cc(t *testing.T, code string) {
	f := ToFourCC(code)
	s := FourCC(f)
	t.Log(code, s, f)
	if strings.ToUpper(code) != s {
		t.Fatal("code!=s", code, s)
	}
}
