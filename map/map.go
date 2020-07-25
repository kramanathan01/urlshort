package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"urlshort"
	"urlshort/persist"

	"github.com/fsnotify/fsnotify"
)

var (
	mapFile string
	port    = flag.Int("p", 8080, "listening port")
)

func main() {

	flag.Parse()

	persist.Db.Open()
	defer persist.Db.DB.Close()

	handler := urlshort.SetHandler(mapFile)
	addr := fmt.Sprintf("localhost:%v", *port)

	// Start server
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}
	go startServer(srv)

	//wait for signal
	err := signalWait(srv, mapFile)
	if err != nil {
		log.Fatalf("Failed to shutdown server %v", err)
	}
}

func init() {
	// READ json map from ~/.map/map.json
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	mapFile = filepath.Join(home, ".map.json")
}

func startServer(srv *http.Server) {
	log.Printf("Starting the server on localhost:%+v", *port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP ListenAndServe: %v", err)
	}
}

func closeServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("Graceful shutdown of server")
	return srv.Shutdown(ctx)
}

func signalWait(srv *http.Server, mfile string) error {
	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	//Set up a file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer watcher.Close()

	err = watcher.Add(mfile)
	if err != nil {
		log.Fatal(err)
		return err
	}

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGUSR1:
				log.Println("User 1 signal received. Reloading config...")
				handler := urlshort.SetHandler(mfile)
				srv.Handler = handler
			case os.Interrupt, syscall.SIGTERM:
				return closeServer(srv)
			}
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file changed. Reloading config...")
				handler := urlshort.SetHandler(mfile)
				srv.Handler = handler
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return err
			}
			log.Println("error:", err)
		}
	}
}
