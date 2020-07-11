package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"urlshort"
)

func main() {
	// READ json map from ~/.map/map.json
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	home := user.HomeDir + "/.map"
	mapFile := home + "/map.json"
	logFile := home + "/map.log"

	// Enable Logging
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)

	mux := defaultMux()

	jsonBlob, err := getContent(mapFile)
	if err != nil {
		log.Panicf("Error getting map JSON: %v", err)
	}

	jsonHandler, err := urlshort.JSONHandler(jsonBlob, mux)
	if err != nil {
		log.Panicf("Error in Handler: %v", err)
	}

	srv := &http.Server{
		Addr:         "localhost:8080",
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      jsonHandler,
	}

	//Start server and listen
	go func() {
		log.Println("Starting the server on :8080")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP ListenAndServe: %v", err)
		}
	}()

	// Graceful Shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Failed to shutdown server %v", err)
	}
	log.Println("Graceful shutdown of server")
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
