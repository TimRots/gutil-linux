package pci

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PATH_SYS_BUS_PCI_DEVICES = "/sys/bus/pci/devices"
	PATH_SYS_DEVICES_PCI     = "/sys/devices/pci"
)

//go:embed pci.ids
var pciIDs embed.FS

var nonWhitespaceRegex = regexp.MustCompile(`[\S]+`)

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

func readFromFile(f string, w, start, end int) (string, error) {
	if start < 0 || end < 0 || start > end {
		return "", errors.New("invalid start:end")
	}

	var value []string

	file, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if w == 0 {
		w = 1
	}
	scanner, i, w := bufio.NewScanner(file), 0, w
	for scanner.Scan() {
		for _, word := range nonWhitespaceRegex.FindAllString(scanner.Text(), -1) {
			switch i++; {
			case i == w:
				value = append(value, string(word))
			}
			ret := strings.Join(value, " ")
			if start != 0 || end != 0 {
				if start >= len(ret) || end > len(ret) {
					return "", errors.New("invalid start:end")
				}
				return ret[start:end], nil
			}
		}
	}
	return "", nil
}

func ParsePciDevices() (PciDevices []PciDevice, err error) {
	var devices []string
	var path string

	read := func(bus, filename string, start, end int) (string, error) {
		return readFromFile(filepath.Join(PATH_SYS_BUS_PCI_DEVICES, bus, filename), 1, start, end)
	}

	lookupKernelDriver := func(bus string) (string, error) {
		path = filepath.Join(PATH_SYS_BUS_PCI_DEVICES, bus, "uevent")
		read, err := readFromFile(path, 1, 0, 0)
		if err != nil {
			return "", err
		}
		if strings.Contains(read, "DRIVER=") {
			return read[7:], nil
		}
		return "", nil
	}

	// Find all devices in /sys/bus/pci/devices/ and append each device to devices[]
	if err = filepath.Walk(
		PATH_SYS_BUS_PCI_DEVICES,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				devices = append(devices, info.Name())
			}
			return nil
		}); err != nil {

		err = fmt.Errorf("failed to walk %s: %w", PATH_SYS_BUS_PCI_DEVICES, err)
		return
	}

	var errs []error
	// Iterate over each bus and parse & append values to PciDevices[]
	for _, bus := range devices {
		var (
			dev    string
			ven    string
			class  string
			subDev string
			subVen string
			mod    string
			irq    string
			rev    string

			venName  string
			devName  string
			devClass string

			subSys string

			kernelDriver string
		)
		if dev, err = read(bus, "device", 2, 6); err != nil {
			errs = append(errs, fmt.Errorf("failed to read device value from bus %s: %w", bus, err))
			continue
		}
		if ven, err = read(bus, "vendor", 2, 6); err != nil {
			errs = append(errs, fmt.Errorf("failed to read vendor value from bus %s: %w", bus, err))
			continue
		}
		if class, err = read(bus, "class", 2, 6); err != nil {
			errs = append(errs, fmt.Errorf("failed to read class value from bus %s: %w", bus, err))
			continue
		}
		if subDev, err = read(bus, "subsystem_device", 2, 6); err != nil {
			errs = append(errs, fmt.Errorf("failed to read subsystem_device value from bus %s: %w", bus, err))
			continue
		}
		if subVen, err = read(bus, "subsystem_vendor", 2, 6); err != nil {
			errs = append(errs, fmt.Errorf("failed to read subsystem_vendor value from bus %s: %w", bus, err))
			continue
		}
		if mod, err = read(bus, "modalias", 0, 0); err != nil {
			errs = append(errs, fmt.Errorf("failed to read modalias value from bus %s: %w", bus, err))
			continue
		}
		if irq, err = read(bus, "irq", 0, 0); err != nil {
			errs = append(errs, fmt.Errorf("failed to read irq value from bus %s: %w", bus, err))
			continue
		}
		if rev, err = read(bus, "revision", 2, 4); err != nil {
			errs = append(errs, fmt.Errorf("failed to read irq value from bus %s: %w", bus, err))
			continue
		}

		if kernelDriver, err = lookupKernelDriver(bus); err != nil {
			errs = append(errs, fmt.Errorf("failed to look up kernel driver used for bus %s: %w", bus, err))
			continue
		}

		venName, _ = Lookup("vendor", ven, "", "", "")
		devName, _ = Lookup("device", ven, dev, "", "")
		devClass, _ = Lookup("class", "", "", class, "")

		if subVen != "0000" {
			subSys, _ = Lookup("subsystem", ven, "", "", subDev)
		}

		PciDevices = append(PciDevices,
			PciDevice{
				bus, ven, dev, class, subVen, subDev, irq, rev,
				venName, devName, devClass, subSys, "", mod, kernelDriver,
			},
		)
	}

	if len(errs) > 0 {
		err = errors.Join(errs...)
	}

	return
}

func Lookup(searchType, ven, dev, class, subclass string) (string, error) {
	var found bool = false

	// Open pci.ids
	f, _ := pciIDs.Open("pci.ids")
	// close pci.ids file when done
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		switch searchType {
		case "vendor": // Return first occurence of ven that does not have a \t prefix.
			if strings.Contains(scanner.Text(), ven) && !strings.HasPrefix(scanner.Text(), "\t") {
				return strings.TrimLeft(scanner.Text(), ven+"  "), nil
			}

		case "device": // Return first occurence of dev after vendor is found
			if strings.Contains(scanner.Text(), ven) && !strings.HasPrefix(scanner.Text(), "\t") {
				found = true
			}
			if strings.HasPrefix(scanner.Text(), "\t"+dev) && found {
				return strings.TrimLeft(scanner.Text(), "\t\t"+dev+"  "), nil
			}

		case "class": // Split class (eg: 0600), search for "C 06" and return first occurence of "\t\t00" therafter.
			if strings.Contains(scanner.Text(), "C"+" "+class[0:2]) && !strings.HasPrefix(scanner.Text(), "\t") {
				found = true
			}
			if strings.HasPrefix(scanner.Text(), "\t"+class[2:4]) && found {
				return strings.TrimLeft(scanner.Text(), "\t\t"+class+"  "), nil

			}
		case "subsystem": // Match and return line "\t\t vendor subsystem_device"
			if strings.Contains(scanner.Text(), ven+" "+subclass) {
				return strings.TrimLeft(scanner.Text(), "\t\t"+ven+" "+subclass), nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "Unknown " + searchType, nil
}
