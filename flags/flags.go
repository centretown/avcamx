package main

import (
	"github.com/centretown/avcamx"
)

func main() {
	avFlags := avcamx.NewAvFlags()
	avFlags.Parse()
	avFlags.Print()

}
