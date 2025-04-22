package avcam

// {"Format":1448695129,"Width":1280,"Height":720,"FPS":{"N":10,"D":1}}
type VideoConfig struct {
	Path   string
	Base   string
	Driver string
	Codec  string
	Width  int
	Height int
	FPS    uint32
}
