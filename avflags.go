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
	HostPort   string
	Remotes    []string
	OutputBase string
	Update     bool
	Recorder   bool
}

func NewAvFlags() (avFlags *AvFlags) {
	avFlags = &AvFlags{}
	*avFlags = avDefaultFlags
	return
}

const (
	ConfigName = "avcamx.json"
)

var (
	avDefaultFlags = AvFlags{
		Remotes:    make([]string, 0),
		HostAddr:   GetOutboundIP(),
		HostPort:   "9000",
		OutputBase: "/mnt/molly/output",
		Update:     false,
		Recorder:   false,
	}

	remoteAddrUsage = "remote host ip address (more than one)"
	hostAddrUsage   = "host ip address"
	hostPortUsage   = "host ip port number"
	outputBaseUsage = "recording directory path"
	updateUsage     = "update default values"
)

func (avFlags *AvFlags) Print() {
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%v: %v\n", f.Usage, f.Value)
	})

	// fmt.Printf("Host: %s:%s\n", avFlags.HostAddr, avFlags.HostPort)
	// fmt.Printf("Remotes:\n")
	// for _, adr := range avFlags.Remotes {
	// 	fmt.Printf("- %s\n", adr)
	// }
	// fmt.Printf("MP3 output to: %s\n", avFlags.OutputBase)
	// fmt.Printf("Update default values: %v\n", avFlags.Update)
}

func (avFlags *AvFlags) Parse() {
	flag.Var((*stringArray)(&avFlags.Remotes), "remote", remoteAddrUsage)
	flag.Var((*stringArray)(&avFlags.Remotes), "r", remoteAddrUsage)
	flag.StringVar(&avFlags.HostAddr, "addr", avDefaultFlags.HostAddr, hostAddrUsage)
	flag.StringVar(&avFlags.HostAddr, "a", avDefaultFlags.HostAddr, hostAddrUsage)
	flag.StringVar(&avFlags.HostPort, "port", avDefaultFlags.HostPort, hostPortUsage)
	flag.StringVar(&avFlags.HostPort, "p", avDefaultFlags.HostPort, hostPortUsage)
	flag.StringVar(&avFlags.OutputBase, "output", avDefaultFlags.OutputBase, outputBaseUsage)
	flag.StringVar(&avFlags.OutputBase, "o", avDefaultFlags.OutputBase, outputBaseUsage)
	flag.BoolVar(&avFlags.Update, "update", avDefaultFlags.Update, updateUsage)
	flag.BoolVar(&avFlags.Update, "u", avDefaultFlags.Update, updateUsage)

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
