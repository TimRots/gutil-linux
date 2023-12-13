package irq

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const PATH_PROC_INTERRUPTS = "/proc/interrupts"

type Interrupt struct {
	Irq   string
	Total int
	Name  string
}

func IrqStat() (Interrupts []Interrupt, err error) {
	var f *os.File
	if f, err = os.Open(PATH_PROC_INTERRUPTS); err != nil {
		return
	}
	defer f.Close()

	var (
		active_cpu_count int
		irq              string
		name             []string
	)
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

	return
}
