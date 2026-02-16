package avcamx

import (
	"net/http"
	"testing"
)

func TestScan(t *testing.T) {

	host := NewAvHost("", "", []string{}, 1000, nil)
	// time.Sleep(time.Second)

	// for range 60 {
	// 	time.Sleep(time.Second * 2)
	// 	if len(host.streams) > 0 && host.streams[0].IsOpened() {
	// 		testCmd(t, host, host.streams[0])
	// 	}
	// }
	t.Log("Streams...")
	for _, stream := range host.Streams() {
		t.Logf("\tURL=%s, Source=%s, Open=%v", stream.Url, stream.Source.Path(), stream.IsOpened())
	}

	t.Log("find Stream... '/video0'")
	stream := host.Stream("/video0")
	if stream != nil {
		t.Logf("\tURL=%s, Source=%s, Open=%v", stream.Url, stream.Source.Path(), stream.IsOpened())
	}

	host.Quit()
}

var (
	cmds = []string{
		"/reset",
		// "/zoomin", "/zoomin", "/zoomin",
		// "/panleft", "/panleft", "/panleft",
		// "/tiltup", "/tiltup", "/tiltup",
	}
	cmdCount int
)

func testCmd(t *testing.T, host *AvHost, avStream *AvStream) {
	if cmdCount >= len(cmds) {
		cmdCount = 0
	}

	resp, err := http.Get("http://" + host.Url + avStream.Url + cmds[cmdCount])
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.Status, resp.Body)
	cmdCount++
}
