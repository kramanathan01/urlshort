package urlshort

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"urlshort/assets"
)

var pathUrls = make(map[string]string)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlpath := strings.TrimLeft(r.URL.Path, "/")
		if path, ok := pathsToUrls[urlpath]; ok {
			http.Redirect(w, r, path, http.StatusFound)
			return
		}
		// Match paths in the map else
		fallback.ServeHTTP(w, r)
	}
}

func SetHandler(mfile string) http.HandlerFunc {
	var err error
	mux := defaultMux()

	pathUrls, err = parseJSON(mfile)
	if err != nil {
		log.Panicf("Error getting map JSON: %v", err)
	}

	return MapHandler(pathUrls, mux)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", msg)
	mux.HandleFunc("/list", listHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))
	return mux
}

func msg(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Set path:url in $HOME/.map.json")
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	listHTML, err := assets.Asset("templates/list.gohtml")
	if err != nil {
		log.Println(err)
		return
	}
	ll := string(listHTML)
	listTemplate := template.Must(template.New("list").Parse(ll))
	push(w, "../static/style.css")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	render(w, r, listTemplate, "list", pathUrls)
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

func push(w http.ResponseWriter, resource string) {
	pusher, ok := w.(http.Pusher)
	if ok {
		if err := pusher.Push(resource, nil); err == nil {
			fmt.Printf("Pusher error: %v\n", err)
			return
		}
	}
}

func render(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data interface{}) {
	buf := new(bytes.Buffer)
	if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
		fmt.Printf("\nRender Error: %v\n", err)
		return
	}
	w.Write(buf.Bytes())
}
