package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	count       = flag.Int("n", 10, "number of iterations")
	durationStr = flag.String("d", "0.5s", "duration of sleep in-between prints")
	message     = flag.String("m", "hello!", "message to print in-between sleeps")
)

func main() {
	flag.Parse()

	duration, err := time.ParseDuration(*durationStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: parsing duration '%s': %v", *durationStr, err)
		os.Exit(1)
	}

	log.Printf(*message)

	for i := 0; i < *count; i++ {
		time.Sleep(duration)
		log.Printf(*message)
	}
}
