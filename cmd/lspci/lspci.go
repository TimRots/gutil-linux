// lspci
// Utility to display pci information.
//
// This applications uses linux-sysfs to gather exposed pci information
// thereafter it will match the values with pci.ids

package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/TimRots/gutil-linux/pci"
)

const (
	PATH_SYS_BUS_PCI_DEVICES = "/sys/bus/pci/devices"
	PATH_SYS_DEVICES_PCI     = "/sys/devices/pci"
	description              = "lspci lists detailed information about all PCI buses and devices in the system"
	usage                    = "Usage: lspci [OPTIONS]\n\n%s\n\nOptions:\n"
)

type PciDevice struct {
	Bus               string
	VendorID          string
	DeviceID          string
	Class             string
	SubsysVendor      string
	SubsysDevice      string
	Irq               string
	Revision          string
	VendorName        string
	DeviceName        string
	DeviceClass       string
	Subsystem         string
	KernelModule      string
	KernelModuleAlias string
	KernelDriver      string
}

var (
	jsonoutput        = flag.Bool("json", false, "use JSON output format")
	numeric           = flag.Bool("numeric", false, "Show numeric ID's")
	numtext           = flag.Bool("numtext", false, "Show both textual and numeric ID's (names & numbers)")
	showdomainnumbers = flag.Bool("domainshow", false, "Always show domain numbers")
	showkerneldrivers = flag.Bool("kerneldrivers", false, "Show kernel drivers handling each device.")
	verbose           = flag.Bool("verbose", false, "Be verbose")
	veryverbose       = flag.Bool("veryverbose", false, "Be very verbose")
	xmloutput         = flag.Bool("xml", false, "use XML output format")

	PciDevices []PciDevice
)

func init() {
	flag.BoolVar(jsonoutput, "j", false, "")
	flag.BoolVar(numeric, "n", false, "")
	flag.BoolVar(numtext, "nn", false, "")
	flag.BoolVar(showdomainnumbers, "D", false, "")
	flag.BoolVar(showkerneldrivers, "k", false, "")
	flag.BoolVar(verbose, "v", false, "")
	flag.BoolVar(veryverbose, "vv", false, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, description)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	switch {
	case *jsonoutput:
		printDevices("json")
	case *xmloutput:
		printDevices("xml")
	default:
		printDevices("")
	}
}

func printDevices(opts string, params ...bool) {
	PciDevices, err := pci.ParsePciDevices()

	switch opts {
	case "json":
		jsonObject, _ := json.MarshalIndent(map[string]interface{}{"pcidevices": PciDevices}, "", "    ")
		fmt.Println(string(jsonObject))
	case "xml":
		xmlObject, _ := xml.MarshalIndent(PciDevices, "", "    ")
		fmt.Println(string(xmlObject))
	default:
		tabWriter(PciDevices)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func tabWriter(PciDevices []pci.PciDevice) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	addBrackets := func(s string) string {
		return "[" + s + "]"
	}

	tw := func(w *tabwriter.Writer, args ...string) {
		fmt.Fprintf(w, strings.Join(args, "\t")+"\n")
	}

	for _, PciDevice := range PciDevices {
		switch {
		case *numtext:
			PciDevice.DeviceName =
				fmt.Sprintf("%v %v: %v %v %v",
					PciDevice.DeviceClass,
					addBrackets(PciDevice.Class),
					PciDevice.VendorName,
					PciDevice.DeviceName,
					addBrackets(PciDevice.VendorID+":"+PciDevice.DeviceID),
				)
		case *numeric:
			PciDevice.DeviceName = fmt.Sprintf("%v: %v:%v",
				PciDevice.Class, PciDevice.VendorID, PciDevice.DeviceID,
			)
		default:
			PciDevice.DeviceName = fmt.Sprintf("%v: %v %v",
				PciDevice.DeviceClass, PciDevice.VendorName, PciDevice.DeviceName,
			)
		}

		if PciDevice.Revision != "00" {
			PciDevice.DeviceName += fmt.Sprintf(" (rev %v)", PciDevice.Revision)
		}
		tw(w, PciDevice.Bus, PciDevice.DeviceName)

		if *verbose || *veryverbose {
			if len(PciDevice.Subsystem) > 0 {
				tw(w, "\tSubsystem:", PciDevice.VendorName+" "+PciDevice.Subsystem)
			}
		}
		if *veryverbose {
			if PciDevice.Irq != "0" {
				tw(w, fmt.Sprintf("\tInterrupt: Pin A routed to IRQ %v", PciDevice.Irq))
			}
		}
		if *showkerneldrivers || *verbose || *veryverbose {
			if len(PciDevice.KernelDriver) > 0 {
				tw(w, "\tKernel driver in use: "+PciDevice.KernelDriver)
			}
		}
	}
}
