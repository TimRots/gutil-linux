// Utility to display Linux kernel interrupt information.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	PATH_PROC_INTERRUPTS = "/proc/interrupts"
	description          = "Utility to display kernel interrupt information."
	usage                = "Usage: lsirq [OPTIONS]\n\n%s\n\nOptions:\n"
)

var (
	active_cpu_count int
	Interrupts       []Interrupt
	irq              string
	name             []string
	noheadings       = flag.Bool("noheadings", false, "don't print headings")
	pairs            = flag.Bool("pairs", false, "use key=\"value\" output format")
	jsonoutput       = flag.Bool("json", false, "use JSON output format")
)

type Interrupt struct {
	Irq   string
	Total int
	Name  string
}

func init() {
	flag.BoolVar(noheadings, "n", false, "")
	flag.BoolVar(pairs, "p", false, "")
	flag.BoolVar(jsonoutput, "j", false, "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, description)
		flag.PrintDefaults()
	}
}

func irqStat() []Interrupt {
	f, err := os.Open(PATH_PROC_INTERRUPTS)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for ln := 0; scanner.Scan(); ln++ {
		if ln == 0 {
			// First line shows the active cpus
			active_cpu_count = strings.Count(scanner.Text(), "CPU")
			continue
		}

		// Match all non-whitespace character sequences
		i, total, words := 0, 0, regexp.MustCompile(`[\S]+`).FindAllString(scanner.Text(), -1)
		for _, word := range words {
			// As the lines are variadic in length we use active_cpu_count and
			// i count to determine how to parse the value based on the word its position.
			i++
			switch {
			case i == 1:
				irq = strings.ReplaceAll(string(word), ":", "")
			case i <= active_cpu_count+1:
				numint, _ := strconv.Atoi(word)
				total += numint
			case i > active_cpu_count+1:
				name = append(name, string(word))
			}
		}
		Interrupts = append(Interrupts, Interrupt{irq, total, strings.Join(name, " ")})
		irq, name = "", nil

	}

	// Sort slice by number
	sort.Slice(Interrupts, func(y, z int) bool {
		return Interrupts[y].Total > Interrupts[z].Total
	})

	return Interrupts
}

func PrintInterrupts(option string, ops ...bool) {
	irqStat, i := irqStat(), 0
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
