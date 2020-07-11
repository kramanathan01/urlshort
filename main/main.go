package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"

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

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	// yaml := `
	// - path: /urlshort
	//   url: https://github.com/gophercises/urlshort
	// - path: /urlshort-final
	//   url: https://github.com/gophercises/urlshort/tree/solution
	// `
	// yamlHandler, err := urlshort.YAMLHandler([]byte(yaml), mapHandler)
	// if err != nil {
	// 	panic(err)
	// }
	// _ = yamlHandler

	jsonBlob, err := getContent(mapFile)
	if err != nil {
		log.Panicf("Error getting map JSON: %v", err)
	}

	jsonHandler, err := urlshort.JSONHandler(jsonBlob, mapHandler)
	if err != nil {
		log.Panicf("Error in Handler: %v", err)
	}

	log.Println("Starting the server on :8080")
	http.ListenAndServe("localhost:8080", jsonHandler)
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
