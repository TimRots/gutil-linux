lspci
---
lspci lists detailed information about all PCI buses and devices in the system
pci.ids is sourced from git://git.ucw.cz/pciids.git

## Usage
```
Usage: lspci [OPTIONS]

lspci lists detailed information about all PCI buses and devices in the system

Options:
  -D
  -domainshow
    	Always show domain numbers
  -j
  -json
    	use JSON output format
  -k
  -kerneldrivers
    	Show kernel drivers handling each device.
  -n
  -nn

  -numeric
    	Show numeric ID's
  -numtext
    	Show both textual and numeric ID's (names & numbers)
  -v
  -verbose
    	Be verbose
  -veryverbose
    	Be very verbose
  -vv

  -xml
        use XML output format
```

## Example

```shell
# lspci -nn -v
00:00.0 Host bridge [0600]: Intel Corporation 440FX - 82441FX PMC [Natoma] [8086:1237] (rev 02)
00:01.0 ISA bridge [0601]: Intel Corporation 82371SB PIIX3 ISA [Natoma/Triton II] [8086:7000]
00:01.1 IDE interface [0101]: Intel Corporation 82371AB/EB/MB PIIX4 IDE [8086:7111] (rev 01)
        Kernel driver in use: ata_piix
00:02.0 VGA compatible controller [0300]: InnoTek Systemberatung GmbH VirtualBox Graphics Adapter [80ee:beef]
        Kernel driver in use: vboxvideo
00:03.0 Ethernet controller [0200]: Intel Corporation 82540EM Gigabit Ethernet Controller [8086:100e] (rev 02)
        Subsystem: Intel Corporation PRO/1000 MT Desktop Adapter
        Kernel driver in use: e1000
00:04.0 System peripheral [0880]: InnoTek Systemberatung GmbH VirtualBox Guest Service [80ee:cafe]
        Kernel driver in use: vboxguest
00:07.0 Bridge [0680]: Intel Corporation 82371AB/EB/MB PIIX4 ACPI [8086:7113] (rev 08)
        Kernel driver in use: piix4_smbus
```

json output
```
$ lspci -j
{
    "pcidevices": [
        {
            "Bus": "0000:00:00.0",
            "VendorID": "8086",
            "DeviceID": "1237",
            "Class": "0600",
            "SubsysVendor": "0000",
            "SubsysDevice": "0000",
            "Irq": "0",
            "Revision": "02",
            "VendorName": "Intel Corporation",
            "DeviceName": "440FX - 82441FX PMC [Natoma]",
            "DeviceClass": "Host bridge",
            "Subsystem": "",
            "KernelModule": "",
            "KernelModuleAlias": "pci:v00008086d00001237sv00000000sd00000000bc06sc00i00",
            "KernelDriver": ""
        },
...
```

xml output
```
$ ./lspci -xml
<PciDevice>
    <Bus>00:00.0</Bus>
    <VendorID>8086</VendorID>
    <DeviceID>1237</DeviceID>
    <Class>0600</Class>
    <SubsysVendor>0000</SubsysVendor>
    <SubsysDevice>0000</SubsysDevice>
    <Irq>0</Irq>
    <Revision>02</Revision>
    <VendorName>Intel Corporation</VendorName>
    <DeviceName>440FX - 82441FX PMC [Natoma]</DeviceName>
    <DeviceClass>Host bridge</DeviceClass>
    <Subsystem></Subsystem>
    <KernelModule></KernelModule>
    <KernelModuleAlias>pci:v00008086d00001237sv00000000sd00000000bc06sc00i00</KernelModuleAlias>
    <KernelDriver></KernelDriver>
</PciDevice>
```
