package avcamx

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type stringArray []string

var _ flag.Value = (*stringArray)(nil)

// String is an implementation of the flag.Value interface
func (sa *stringArray) String() string {
	return fmt.Sprintf("%v", *sa)
}

// Set is an implementation of the flag.Value interface
func (sa *stringArray) Set(value string) error {
	*sa = append(*sa, value)
	return nil
}

type AvFlags struct {
	HostAddr   string
	Connect    string
	Remotes    []string
	OutputBase string
	Update     bool
	Recorders  int
}

func NewAvFlags() (avFlags *AvFlags) {
	avFlags = &AvFlags{}
	*avFlags = avDefaultFlags
	return
}

const (
	ConfigName       = "avcamx.json"
	CONNECT_NONE     = "none"
	CONNECT_ALL      = "all"
	CONNECT_RESTRICT = "restrict"
)

var (
	avDefaultFlags = AvFlags{
		Connect:    CONNECT_NONE,
		Remotes:    make([]string, 0),
		HostAddr:   GetOutboundIP(),
		OutputBase: "/mnt/molly/output",
		Update:     false,
		Recorders:  0,
	}

	remoteAddrUsage = "remote host ip address (more than one)"
	connectUsage    = "remote connections (none,all,restrict)"
	hostAddrUsage   = "host ip address"
	outputBaseUsage = "recording directory path"
	updateUsage     = "update default values"
)

func (avFlags *AvFlags) Print() {
	fmt.Printf("Host: %s\n", avFlags.HostAddr)
	fmt.Printf("Remote Connections: %s\n", avFlags.Connect)
	fmt.Printf("Remotes:\n")
	for _, adr := range avFlags.Remotes {
		fmt.Printf("- %s\n", adr)
	}
	fmt.Printf("MP3 output to: %s\n", avFlags.OutputBase)
	fmt.Printf("Number of recorders supported:: %d\n", avFlags.Recorders)
	fmt.Printf("Update default values: %v\n", avFlags.Update)
}

func (avFlags *AvFlags) Parse() {
	flag.StringVar(&avFlags.HostAddr, "address", avFlags.HostAddr, hostAddrUsage)
	flag.StringVar(&avFlags.HostAddr, "a", avFlags.HostAddr, hostAddrUsage)
	flag.StringVar(&avFlags.Connect, "connect", avFlags.Connect, connectUsage)
	flag.StringVar(&avFlags.Connect, "c", avFlags.Connect, connectUsage)
	flag.StringVar(&avFlags.OutputBase, "output", avFlags.OutputBase, outputBaseUsage)
	flag.StringVar(&avFlags.OutputBase, "o", avFlags.OutputBase, outputBaseUsage)
	flag.BoolVar(&avFlags.Update, "update", avFlags.Update, updateUsage)
	flag.BoolVar(&avFlags.Update, "u", avFlags.Update, updateUsage)

	flag.Var((*stringArray)(&avFlags.Remotes), "remote", remoteAddrUsage)
	flag.Var((*stringArray)(&avFlags.Remotes), "r", remoteAddrUsage)

	flag.Parse()

	if avFlags.Update {
		exists := avFlags.HasFile()
		err := avFlags.Save()
		if err != nil {
			log.Printf("Error updating configuration file %s. %s", ConfigName, err)
		} else if exists {
			log.Print("Updated configuration file. ", ConfigName)
		} else {
			log.Print("Created configuration file. ", ConfigName)
		}
	}

}

func (avFlags *AvFlags) HasFile() bool {
	_, err := os.Stat(ConfigName)
	return err == nil
}

func (avFlags *AvFlags) Load() (err error) {
	var buf []byte

	buf, err = os.ReadFile(ConfigName)
	if err != nil {
		log.Printf("AvFlags Load ReadFile error: %s", err)
		return
	}

	err = json.Unmarshal(buf, avFlags)
	if err != nil {
		log.Printf("AvFlags Load Unmarshal error: %s", err)
		return
	}

	return
}

func (avFlags *AvFlags) Save() (err error) {
	var buf []byte
	buf, err = json.MarshalIndent(avFlags, "", "  ")
	if err != nil {
		log.Printf("AvFlags Save Marshall error: %s", err)
		return
	}

	err = os.WriteFile(ConfigName, buf, 0644)
	if err != nil {
		log.Printf("AvFlags Save WriteFile error: %s", err)
		return
	}

	return
}

func (avFlags *AvFlags) SetDefault() { *avFlags = avDefaultFlags }
