package avcam

type AudioSource interface {
	Record(stop chan int)
	IsEnabled() bool
}
