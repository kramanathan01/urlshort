package urlshort

import (
	"encoding/json"
	"net/http"
	"strings"
)

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

// Struct members need to be exported for Unmarshalling to work
type pathURL struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the JSON, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
// [
// 	{"path": "/some-path", "url": "https://www.some-url.com/demo"}
// ]
//
// The only errors that can be returned all related to having
// invalid JSON data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(jsn []byte, fallback http.Handler) (http.HandlerFunc, error) {
	pathUrls, err := parseJSON(jsn)
	if err != nil {
		return nil, err
	}
	pathToUrls := buildMap(pathUrls)
	return MapHandler(pathToUrls, fallback), nil
}

func parseJSON(data []byte) ([]pathURL, error) {
	var pathUrls []pathURL
	err := json.Unmarshal(data, &pathUrls)
	if err != nil {
		return nil, err
	}
	return pathUrls, nil
}

func buildMap(pathUrls []pathURL) map[string]string {
	pathToUrls := make(map[string]string)
	for _, pu := range pathUrls {
		pathToUrls[pu.Path] = pu.URL
	}
	return pathToUrls
}
