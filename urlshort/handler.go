package urlshort

import (
	"net/http"

	"gopkg.in/yaml.v2"
)

type PathURL struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		dest, ok := pathsToUrls[path]
		if ok {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}
		fallback.ServeHTTP(w, r)
	}
}

func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var pathsToURLs []PathURL
	pathMap := make(map[string]string)

	err := yaml.Unmarshal(yml, &pathsToURLs)
	if err != nil {
		return nil, err
	}

	for _, v := range pathsToURLs {
		pathMap[v.Path] = v.URL
	}

	return MapHandler(pathMap, fallback), nil
}
