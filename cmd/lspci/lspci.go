// lspci
// Utility to display pci information.
//
// This applications uses linux-sysfs to gather exposed pci information
// thereafter it will match the values with pci.ids

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
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

	PciDevices []PciDevice
	errCnt     = 0
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
	default:
		printDevices("")
	}
}

func printDevices(opts string, params ...bool) {
	PciDevices := ParsePciDevices()

	switch opts {
	case "json":
		jsonObject, _ := json.MarshalIndent(map[string]interface{}{"pcidevices": PciDevices}, "", "    ")
		fmt.Println(string(jsonObject))
	default:
		tabWriter(PciDevices)
	}
}

func tabWriter(PciDevices []PciDevice) {
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

func readFromFile(f string, w int) string {
	var value []string

	file, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if w == 0 {
		w = 1
	}
	scanner, i, w := bufio.NewScanner(file), 0, w
	for scanner.Scan() {
		for _, word := range regexp.MustCompile(`[\S]+`).FindAllString(scanner.Text(), -1) {
			switch i++; {
			case i == w:
				value = append(value, string(word))
			}
			return strings.Join(value, " ")
		}
	}
	return ""
}

func Lookup(searchType, ven, dev, class, subclass string) string {
	var found bool = false

	f, err := os.Open("./pci.ids")
	defer f.Close()
	if err != nil {
		errCnt++
		if errCnt < 2 {
			fmt.Println("Error: Cannot open pci.ids file, defaulting to numeric")
		}
		*numeric = true
		return ""
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		switch searchType {
		case "vendor": // Return first occurence of ven that does not have a \t prefix.
			if strings.Contains(scanner.Text(), ven) && !strings.HasPrefix(scanner.Text(), "\t") {
				return strings.TrimLeft(scanner.Text(), ven+"  ")
			}

		case "device": // Return first occurence of dev after vendor is found
			if strings.Contains(scanner.Text(), ven) && !strings.HasPrefix(scanner.Text(), "\t") {
				found = true
			}
			if strings.HasPrefix(scanner.Text(), "\t"+dev) && found {
				return strings.TrimLeft(scanner.Text(), "\t\t"+dev+"  ")
			}

		case "class": // Split class (eg: 0600), search for "C 06" and return first occurence of "\t\t00" therafter.
			if strings.Contains(scanner.Text(), "C"+" "+class[0:2]) && !strings.HasPrefix(scanner.Text(), "\t") {
				found = true
			}
			if strings.HasPrefix(scanner.Text(), "\t"+class[2:4]) && found {
				return strings.TrimLeft(scanner.Text(), "\t\t"+class+"  ")

			}
		case "subsystem": // Match and return line "\t\t vendor subsystem_device"
			if strings.Contains(scanner.Text(), ven+" "+subclass) {
				return strings.TrimLeft(scanner.Text(), "\t\t"+ven+" "+subclass)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return "Unknown " + searchType
}

func ParsePciDevices() []PciDevice {
	var devices []string

	buildPath := func(items ...string) string {
		var path string
		for _, item := range items {
			path += item + "/"
		}
		return strings.TrimSuffix(path, "/")
	}

	read := func(bus, filename string) string {
		return readFromFile(buildPath(PATH_SYS_BUS_PCI_DEVICES, bus, filename), 1)
	}

	lookupKernelDriver := func(bus string) string {
		read := readFromFile(buildPath(PATH_SYS_DEVICES_PCI+bus[0:7], bus, "uevent"), 1)
		if strings.Contains(read, "DRIVER=") {
			return read[7:]
		}
		return ""
	}

	// Find all devices in /sys/bus/pci/devices/ and append each device to devices[]
	if err := filepath.Walk(
		PATH_SYS_BUS_PCI_DEVICES,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				devices = append(devices, info.Name())
			}
			return nil
		}); err != nil {
		log.Fatal(err)
	}

	// Iterate over each bus and parse & append values to PciDevices[]
	for _, bus := range devices {
		dev := read(bus, "device")[2:6]
		ven := read(bus, "vendor")[2:6]
		class := read(bus, "class")[2:6]
		subDev := read(bus, "subsystem_device")[2:6]
		subVen := read(bus, "subsystem_vendor")[2:6]
		mod := read(bus, "modalias")
		irq := read(bus, "irq")
		rev := read(bus, "revision")[2:4]

		venName := Lookup("vendor", ven, "", "", "")
		devName := Lookup("device", ven, dev, "", "")
		devClass := Lookup("class", "", "", class, "")

		subSys := ""
		if subVen != "0000" {
			subSys = Lookup("subsystem", ven, "", "", subDev)
		}

		kernelDriver := lookupKernelDriver(bus)

		if !*showdomainnumbers && !*jsonoutput {
			bus = strings.TrimPrefix(bus, "0000:")
		}

		PciDevices = append(PciDevices,
			PciDevice{
				bus, ven, dev, class, subVen, subDev, irq, rev,
				venName, devName, devClass, subSys, "", mod, kernelDriver,
			},
		)
	}
	return PciDevices
}
