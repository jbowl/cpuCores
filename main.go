package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func fib(n int) int {
	if n <= 1 {
		return n
	}

	return fib(n-1) + fib(n-2)
}

func worker(jobs <-chan int, results chan<- int) {
	for n := range jobs {
		results <- fib(n)
	}
}

type cpuData struct {
	deltaIdleTime  uint64
	deltaTotalTime uint64
	cpuUsage       float64
	prevIdleTime   uint64
	prevTotalTime  uint64
}

// read cpu usage from /proc/stat
func cpuUsage(percentUsage chan string, exit chan int) {

	cpus := make([]cpuData, runtime.NumCPU())
	var sb strings.Builder

	for {
		for i := 0; i < 2; i++ {
			file, err := os.Open("/proc/stat")
			if err != nil {
				log.Fatal(err)
			}
			scanner := bufio.NewScanner(file)
			scanner.Scan()

			scanner.Text() // generic

			//for core := 0; core < cores; core++ {
			for core := 0; core < len(cpus); core++ {

				scanner.Scan()

				//	test := scanner.Text()
				//		fmt.Println(test)

				cpuLine := scanner.Text()[5:] // fix for > 9 cpus
				file.Close()
				if err := scanner.Err(); err != nil {
					log.Fatal(err)
				}
				split := strings.Fields(cpuLine)
				idleTime, _ := strconv.ParseUint(split[3], 10, 64)
				totalTime := uint64(0)
				for _, s := range split {
					u, _ := strconv.ParseUint(s, 10, 64)
					totalTime += u
				}
				if i > 0 {

					cpus[core].deltaIdleTime = idleTime - cpus[core].prevIdleTime
					cpus[core].deltaTotalTime = totalTime - cpus[core].prevTotalTime

					cpus[core].cpuUsage = (1.0 - float64(cpus[core].deltaIdleTime)/float64(cpus[core].deltaTotalTime)) * 100.0

					sb.WriteString("CPU USAGE: ")

					for count := 0; count < len(cpus); count++ {

						fmt.Fprintf(&sb, "%2f ", cpus[count].cpuUsage)
					}

					select {
					case percentUsage <- sb.String():
						sb.Reset()
					case <-exit:
						return
					}

				}
				cpus[core].prevIdleTime = idleTime
				cpus[core].prevTotalTime = totalTime

			}
			time.Sleep(time.Second)
		}
	}
}

func main() {

	var count int

	if len(os.Args) > 1 {
		strCount := os.Args[1]
		c, err := strconv.ParseInt(strCount, 10, 32)

		if err == nil {
			count = int(c)
		}
	}

	//	c := NewCPUWrapper(runtime.NumCPU())

	jobs := make(chan int, count)
	results := make(chan int, 10)
	usages := make(chan string)
	exit := make(chan int)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		go worker(jobs, results)
	}

	go cpuUsage(usages, exit)

	for i := 0; i < int(count); i++ {
		jobs <- i
	}

	close(jobs)

	// i := 0
	for i := 0; i < int(count); i++ {

		select {
		case x := <-results:
			fmt.Println(x)
		case usage := <-usages:
			fmt.Println(usage)
		}

	}

	exit <- 1

	close(results)
}
