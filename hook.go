package avcam

type Hook interface {
	Update(img any)
	Close(int)
}

type UiHook interface {
	Hook
	SetUi(ui interface{})
}
