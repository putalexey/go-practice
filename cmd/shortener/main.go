package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/putalexey/go-practicum/cmd/shortener/config"
	"github.com/putalexey/go-practicum/internal/app"
)

var buildVersion = "N/A"
var buildDate = "N/A"
var buildCommit = "N/A"

func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Parse()
	if cfg.ProfileCPUFile != "" {
		fProfileCPU, err := os.Create(cfg.ProfileCPUFile)
		if err != nil {
			panic(err)
		}
		defer fProfileCPU.Close()
		if err := pprof.StartCPUProfile(fProfileCPU); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}

	ctx, cancel := context.WithCancel(context.Background())

	finished := sync.WaitGroup{}
	finished.Add(1)
	go func() {
		defer finished.Done()
		app.Run(ctx, cfg)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
	case <-ctx.Done():
	}

	log.Println("Shutting down server...")
	cancel()

	finished.Wait()
}
