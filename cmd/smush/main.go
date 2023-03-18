package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"go.husin.dev/smush"
)

var (
	configPath  = flag.String("c", "smush.yaml", "path to smush configuration file")
	parallelism = flag.Int64("p", int64(runtime.NumCPU()-1), "maximum parallel commands")

	printVersionOnly = flag.Bool("v", false, "print version and exit")
)

func main() {
	flag.Parse()
	if *printVersionOnly {
		printVersion()
		return
	}

	if err := run(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func loadConfig() *smush.Config {
	f, err := os.Open(*configPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	c, err := smush.ReadConfig(f)
	if err != nil {
		panic(err)
	}
	return c
}

func run() error {
	c := loadConfig()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		<-ctx.Done()
		fmt.Println("terminating")
	}()

	return smush.RunAll(ctx, *parallelism, c.Commands)
}
