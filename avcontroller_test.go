package avcamx

import "testing"

func TestController(t *testing.T) {
	url := "/zoomin"
	avUrl, ok := AvUrlToName[url]
	if !ok {
		t.Fatal(url, " not found")
	}
	t.Log(avUrl, "found")

	url = "/zoomout"
	avUrl, ok = AvUrlToName[url]
	if !ok {
		t.Fatal(url, " not found")
	}
	t.Log(avUrl, "found")

	for url, avUrl = range AvUrlToName {
		t.Log("url: ", url, " name: ", avUrl.Name)
	}
}
