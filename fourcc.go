package avcam

import "strings"

func FourCC(fourcc uint32) string {
	code := make([]byte, 4)
	code[0] = byte(fourcc)
	code[1] = byte(fourcc >> 8)
	code[2] = byte(fourcc >> 16)
	code[3] = byte(fourcc >> 24)
	return string(code)
}

func ToFourCC(s string) uint32 {
	if len(s) < 4 {
		return 0
	}

	bs := ([]byte)(strings.ToUpper(s))
	return uint32(bs[0]) |
		(uint32(bs[1]) << 8) |
		(uint32(bs[2]) << 16) |
		(uint32(bs[3]) << 24)
}
