package main

import (
	"fmt"

	"github.com/centretown/avcamx"
	"github.com/korandiz/v4l"
)

func main() {
	devices := v4l.FindDevices()
	for _, info := range devices {
		if !info.Camera {
			fmt.Printf("Not a camera %s\n", info.Path)
			continue
		}
		if info.DriverName != avcamx.UVCVideoDriver {
			fmt.Printf("Uses '%s' driver camera %s\n",
				info.DriverName,
				info.Path)
			continue
		}
		device, err := v4l.Open(info.Path)
		if err != nil {
			fmt.Printf("Error opening device %s: %v\n", info.Path, err)
			continue
		}

		controls, err := device.ListControls()
		if err != nil {

			fmt.Printf("Error listing controls %s@%s: %v\n", info.DeviceName, info.Path, err)
			continue
		}

		fmt.Printf("%s@%s available controls:\n", info.DeviceName, info.Path)
		for _, ctl := range controls {
			value, err := device.GetControl(ctl.CID)
			if err != nil {
				fmt.Printf("Error GetControl %s@%s: %v", info.DeviceName, info.Path, err)
				continue
			}

			fmt.Printf("%s: value = %d, min = %d, max = %d, default = %d, step = %d\n",
				ctl.Name, value, ctl.Min, ctl.Max, ctl.Default, ctl.Step)
		}
		fmt.Println()

		configs, err := device.ListConfigs()
		if err != nil {
			fmt.Printf("Error listing configs %s@%s: %v\n", info.DeviceName, info.Path, err)
			continue
		}

		fmt.Printf("%s@%s available configurations:\n", info.DeviceName, info.Path)
		for i, c := range configs {
			fmt.Printf("%d: %dx%d @%dfps %s\n", i, c.Width, c.Height, c.FPS.N, avcamx.FourCC(c.Format))
			err := device.SetConfig(c)
			if err != nil {
				fmt.Printf("Error setting config: %v\n", err)
				continue
			}

			err = device.TurnOn()
			if err != nil {
				fmt.Printf("Error turning on config: %v\n", err)
				continue
			}

			fmt.Printf("Successfully turned on device\n")
			device.TurnOff()
		}
		fmt.Print("\n\n")
	}
}
