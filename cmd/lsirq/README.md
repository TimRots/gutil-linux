lsirq
---
Utility to display Linux kernel interrupt information.


Build:
```shell
$ env GOOS=linux GOARCH=amd64 go build -o lsirq lsirq.go
```

Usage:
```shell
$  ./lsirq -h
Usage: lsirq [OPTIONS]

Utility to display kernel interrupt information.

Options:
  -j
  -json
    	use JSON output format
  -n
  -noheadings
    	don't print headings
  -p
  -pairs
    	use key="value" output format
$  ./lsirq
IRQ    TOTAL NAME
LOC   609122 Local timer interrupts
RES   100788 Rescheduling interrupts
 19    48056 IO-APIC 19-fasteoi eth0
 20    48042 IO-APIC 20-fasteoi vboxguest
CAL    18173 Function call interrupts
 14    14781 IO-APIC 14-edge ata_piix
 ...
$  ./lsirq -j
{
	"interrupts": [
		{
			"Irq": "LOC",
			"Total": 616641,
			"Name": "Local timer interrupts"
		},
		{
			"Irq": "RES",
			"Total": 102152,
$ ./lsirq -p
IRQ="LOC" TOTAL="632728" NAME="Local timer interrupts"
IRQ="RES" TOTAL="104617" NAME="Rescheduling interrupts"
IRQ="20" TOTAL="54285" NAME="IO-APIC 20-fasteoi vboxguest"
IRQ="19" TOTAL="50181" NAME="IO-APIC 19-fasteoi eth0"
IRQ="CAL" TOTAL="18349" NAME="Function call interrupts"
IRQ="14" TOTAL="14904" NAME="IO-APIC 14-edge ata_piix"
IRQ="TLB" TOTAL="985" NAME="TLB shootdowns"
IRQ="MCP" TOTAL="372" NAME="Machine check polls"
IRQ="12" TOTAL="156" NAME="IO-APIC 12-edge i8042"
```
