package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"urlshort"
)

var (
	home, mapFile string
)

func main() {

	// mux := defaultMux()

	// pathUrls, err := parseJSON(mapFile)
	// if err != nil {
	// 	log.Panicf("Error getting map JSON: %v", err)
	// }

	// mapHandler := urlshort.MapHandler(pathUrls, mux)

	handler := setHandler()

	// Start Server
	srv := &http.Server{
		Addr:         "localhost:8080",
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
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	home = user.HomeDir
	mapFile = filepath.Join(home, ".map.json")
}

func startServer(srv *http.Server) {
	log.Println("Starting the server on :8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP ListenAndServe: %v", err)
	}
}

func signalWait(srv *http.Server, mfile string) error {
	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)
	for {
		sig := <-sigs
		log.Printf("Received %+v", sig)

		switch sig {
		case syscall.SIGUSR1:
			log.Println("Reloading config")
			// pathUrls, err := parseJSON(mfile)
			// if err != nil {
			// 	log.Panicf("Error getting map JSON: %v", err)
			// }
		case os.Interrupt, syscall.SIGTERM:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			log.Println("Graceful shutdown of server")
			return srv.Shutdown(ctx)
		}
	}

}
func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func getContent(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func parseJSON(file string) (map[string]string, error) {
	m := make(map[string]string)
	jb, err := getContent(file)
	if err != nil {
		log.Panicf("Error getting map JSON: %v", err)
	}
	err = json.Unmarshal(jb, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func setHandler() http.HandlerFunc {
	mux := defaultMux()

	pathUrls, err := parseJSON(mapFile)
	if err != nil {
		log.Panicf("Error getting map JSON: %v", err)
	}

	return urlshort.MapHandler(pathUrls, mux)
}
