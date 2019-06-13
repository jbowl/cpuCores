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

type cpuWrapper struct {
	cpus []cpuData
}

func NewCPUWrapper(count int) *cpuWrapper {
	var c cpuWrapper
	c.cpus = make([]cpuData, count)
	return &c
}

// read cpu usage from /proc/stat
func cpuUsage(c *cpuWrapper) {

	//	cores := runtime.NumCPU()
	//cpus := make([]cpuData, cores)

	for i := 0; i < 2; i++ {
		file, err := os.Open("/proc/stat")
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Scan()

		scanner.Text() // generic

		//for core := 0; core < cores; core++ {
		for core := 0; core < len(c.cpus); core++ {

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

				c.cpus[core].deltaIdleTime = idleTime - c.cpus[core].prevIdleTime
				c.cpus[core].deltaTotalTime = totalTime - c.cpus[core].prevTotalTime

				c.cpus[core].cpuUsage = (1.0 - float64(c.cpus[core].deltaIdleTime)/float64(c.cpus[core].deltaTotalTime)) * 100.0
			}
			c.cpus[core].prevIdleTime = idleTime
			c.cpus[core].prevTotalTime = totalTime

		}
		time.Sleep(time.Second)
	}

}

func main() {

	c := NewCPUWrapper(runtime.NumCPU())

	jobs := make(chan int, 100)

	results := make(chan int, 100)

	go worker(jobs, results)
	go worker(jobs, results)
	//go worker(jobs, results)
	//go worker(jobs, results)

	//	for i := 0; i < cores; i++ {
	//		go worker(jobs, results)
	//		}

	for i := 0; i < 100; i++ {
		jobs <- i
	}

	close(jobs)

	for i := 0; i < 100; i++ {

		select {
		case x := <-results:
			fmt.Println(x)

		default:
			cpuUsage(c)

			fmt.Printf("CPU USAGE: ")

			for count := 0; count < len(c.cpus); count++ {
				fmt.Printf("%2f ", c.cpus[count].cpuUsage)
			}
			fmt.Println("")

		}

	}

	close(results)

}
