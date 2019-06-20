package main

import (
	"bufio"
	"flag"
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
	cpuUsage       int
	prevIdleTime   uint64
	prevTotalTime  uint64
}

// read cpu usage from /proc/stat
func cpuUsage(percentUsage chan []int) {

	cpus := make([]cpuData, runtime.NumCPU())

	for {
		// loop twice for comparison calculations
		for i := 0; i < 2; i++ {
			file, err := os.Open("/proc/stat")
			if err != nil {
				log.Fatal(err)
			}
			scanner := bufio.NewScanner(file)
			scanner.Scan() //ignore first line, not core specific

			scanner.Text() 

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

					cpus[core].cpuUsage = int((1.0 - float64(cpus[core].deltaIdleTime)/float64(cpus[core].deltaTotalTime)) * 100.0)

					c := make([]int, len(cpus))

					for count := 0; count < len(cpus); count++ {

						c[count] = cpus[count].cpuUsage
					}

					select {
					case _, open := <-percentUsage:
						if open == false {
							return
						}

					default:
						percentUsage <- c
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

	pWorkers := flag.Int("workers", runtime.NumCPU(), "num of worker goroutines")
	//	workers := runtime.NumCPU()
	//	pWorkers := &workers
	pFn := flag.Int("Fn", 50, "num Fibonacci numbers")
	flag.Parse()

	startTime := time.Now()

	jobs := make(chan int, *pFn)
	results := make(chan int, *pFn)

	//	cpus := make([]int, runtime.NumCPU())

	usages := make(chan []int)
	// workers will pull from job channel

	for i := 0; i < *pWorkers; i++ {
		go worker(jobs, results)
	}

	if runtime.GOOS == "linux" {
		go cpuUsage(usages)
	}
	// load up job channel
	for i := 0; i < *pFn; i++ {
		jobs <- i
	}

	close(jobs)
	var sb strings.Builder
	var strCoreUsage string
	for i := 0; i < *pFn; {
		select {
		case x := <-results:
			fmt.Println(x, "\t", strCoreUsage)
			i++
		case usage := <-usages:
			sb.WriteString("CPU USAGE: ")

			for count := 0; count < len(usage); count++ {
				fmt.Fprintf(&sb, "%d ", usage[count])
			}

			strCoreUsage = sb.String()
			sb.Reset()
		}
	}

	close(results)

	elapsedTime := time.Since(startTime)

	fmt.Printf("calculated %d Fn numbers using %d goroutines in %s\n", *pFn, *pWorkers, elapsedTime)
}
