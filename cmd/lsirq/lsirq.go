// Utility to display Linux kernel interrupt information.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/TimRots/gutil-linux/irq"
)

const (
	description = "Utility to display kernel interrupt information."
	usage       = "Usage: lsirq [OPTIONS]\n\n%s\n\nOptions:\n"
)

var (
	noheadings = flag.Bool("noheadings", false, "don't print headings")
	pairs      = flag.Bool("pairs", false, "use key=\"value\" output format")
	jsonoutput = flag.Bool("json", false, "use JSON output format")
)

func init() {
	flag.BoolVar(noheadings, "n", false, "")
	flag.BoolVar(pairs, "p", false, "")
	flag.BoolVar(jsonoutput, "j", false, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, description)
		flag.PrintDefaults()
	}
}

func PrintInterrupts(option string, ops ...bool) {
	irqStat, err := irq.IrqStat()
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for _, interrupt := range irqStat {
		switch option {
		case "pairs":
			fmt.Printf("IRQ=\"%+v\" TOTAL=\"%+v\" NAME=\"%+v\"\n", interrupt.Irq, interrupt.Total, interrupt.Name)
		case "json":
			jsonObject, _ := json.MarshalIndent(map[string]interface{}{"interrupts": irqStat}, "", "	")
			fmt.Println(string(jsonObject))
			os.Exit(0)
		default:
			if !*noheadings {
				if i++; i == 1 {
					fmt.Printf("%+3v %+8v %+v\n", "IRQ", "TOTAL", "NAME")
				}
			}
			fmt.Printf("%+3v %+8v %+v\n", interrupt.Irq, interrupt.Total, interrupt.Name)
		}
	}
}

func main() {
	flag.Parse()
	switch {
	case *jsonoutput:
		PrintInterrupts("json")
	case *pairs:
		PrintInterrupts("pairs")
	default:
		PrintInterrupts("", *noheadings)
	}
}
