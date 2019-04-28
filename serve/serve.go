package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// listenAndServe serves the provided handler on :8080 and runs debug and other handlers
// on localhost:8081. It will by default shut down gracefully on interrupt and term signals,
// giving the caller an opportunity to finish any calls.
func listenAndServe(handler http.Handler) {
	listenAndServeGraceful(2, 30, handler)
}

func listenAndServeGraceful(waitSeconds, gracefulSeconds int, handler http.Handler) {
	if handler == nil {
		handler = http.DefaultServeMux
	}
	addr := os.Getenv("LISTEN_ADDR")
	if len(addr) == 0 {
		addr = ":8080"
	}
	localServer := &http.Server{
		Addr:         "localhost:8081",
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute,
		Handler:      http.DefaultServeMux,
	}
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute,
		Handler:      handler,
	}
	var wg sync.WaitGroup
	ch := make(chan error, 4)
	wg.Add(2)
	go func() { defer wg.Done(); ch <- server.ListenAndServe() }()
	go func() { defer wg.Done(); ch <- localServer.ListenAndServe() }()

	waitDuration := 5 * time.Second
	graceDuration := 30 * time.Second

	interruptCh := make(chan os.Signal, 10)
	signal.Notify(interruptCh, syscall.SIGTERM, syscall.SIGINT)

	log.Printf("Listening on %s (debug on %s)", addr, localServer.Addr)
	select {
	case err := <-ch:
		log.Printf("error: server exited, shutting down gracefully: %v", err)
	case <-interruptCh:
		log.Printf("Starting graceful shutdown")
	}

	ctx, cancel := context.WithTimeout(context.Background(), waitDuration+graceDuration)
	defer cancel()
	go func() {
		select {
		case <-interruptCh:
			log.Printf("Second interrupt, exiting now")
			cancel()
		}
	}()

	select {
	case <-time.After(waitDuration):
	case <-ctx.Done():
	}

	wg.Add(2)
	go func() { defer wg.Done(); ch <- server.Shutdown(ctx) }()
	go func() { defer wg.Done(); ch <- localServer.Shutdown(ctx) }()

	wg.Wait()
	close(ch)
	for err := range ch {
		if err == nil || err == http.ErrServerClosed {
			continue
		}
		log.Printf("error: server exited: %v", err)
	}
	log.Printf("Done")
}
