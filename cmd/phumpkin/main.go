//go:generate go run assets_gen.go

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pzl/phumpkin/pkg/server"
)

func main() {
	opts := parseCLI()
	s := server.New(opts...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, os.Kill, syscall.SIGQUIT)
		<-sigint
		cancel()
	}()

	// start server with context
	err := s.Start(ctx)
	if err != nil {
		panic(err)
	}

	c2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)

	// give <timeout> chance to shutdown gracefully
	defer s.Shutdown(c2)
	defer cancel2()
}
