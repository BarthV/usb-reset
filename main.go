package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/gousb"
	"github.com/google/gousb/usbid"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(str string) error {
	*i = append(*i, str)
	return nil
}

var (
	usbDebug        int
	vendorStrings   arrayFlags
	deviceStrings   arrayFlags
	waitAfterReset  time.Duration
	continueOnError bool
)

func main() {
	flag.IntVar(&usbDebug, "debug", 0, "libusb debug level (0..3)")
	flag.Var(&vendorStrings, "vendorid", "Reset all devices with corresponding vendorID. Can be specified multiple times.")
	flag.Var(&deviceStrings, "device", "Reset all devices with corresponding vID:pID. Can be specified multiple times.")
	flag.DurationVar(&waitAfterReset, "wait", 4*time.Second, "Sleep time after USB device reset")
	flag.BoolVar(&continueOnError, "continue", false, "If set, the program will not if a single reset fails")
	flag.Parse()

	// Only one context should be needed for an application. It should always be closed at the end.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Debugging can be turned on; this shows some of the inner workings of the libusb package.
	ctx.Debug(usbDebug)

	// OpenDevices is used to find the devices to open.
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		// The usbid package can be used to print out human readable information.
		fmt.Printf("%03d.%03d %s:%s %s\n", desc.Bus, desc.Address, desc.Vendor, desc.Product, usbid.Describe(desc))
		fmt.Printf("Protocol: %s\n", usbid.Classify(desc))

		// If the vendor matches input args, keep the device
		for _, vendorString := range vendorStrings {
			if desc.Vendor.String() == vendorString {
				fmt.Println("VendorID matching, it will be reset.")
				return true
			}
		}

		// If the complete deviceID matches input args, keep the device
		currentDevice := desc.Vendor.String() + ":" + desc.Product.String()
		for _, deviceString := range deviceStrings {
			if currentDevice == deviceString {
				fmt.Println("Device ID matching, it will be reset.")
				return true
			}
		}

		// After inspecting the descriptor, return true or false depending on whether
		// the device is "interesting" or not.  Any descriptor for which true is returned
		// opens a Device which is retuned in a slice (and must be subsequently closed).
		return false
	})

	// All Devices returned from OpenDevices must be closed.
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()

	// OpenDevices can occasionally fail, so be sure to check its return value.
	if err != nil {
		log.Fatalf("list: %s", err)
	}

	for _, dev := range devs {
		// Once the device has been selected from OpenDevices, it is opened
		// and can be interacted with.

		// resetting the current device
		fmt.Printf("Resetting USB for device %s\n", dev.String())
		fmt.Printf("at Address %03d:%03d (speed %s)\n", dev.Desc.Address, dev.Desc.Bus, dev.Desc.Speed.String())
		err := dev.Reset()
		if err != nil {
			if continueOnError {
				log.Printf("reset failed: %s", err)
			} else {
				log.Fatalf("reset failed: %s", err)
			}
		}
		time.Sleep(waitAfterReset)
	}
}
