package avcamx

import (
	"encoding/json"
	"testing"
)

func TestAvHost(t *testing.T) {
	host := NewAvHost("", "")
	host.Load()

	buf, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\n", string(buf))
}
