package avcamx

type AudioSource interface {
	Record(stop chan int)
	IsEnabled() bool
}
