// Based on https://github.com/gin-contrib/static

package static

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const INDEX = "index.html"

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type localFileSystem struct {
	http.FileSystem
	root    string
	indexes bool
}

func LocalFile(root string, index bool) *localFileSystem {
	return &localFileSystem{
		FileSystem: http.Dir(root),
		root:       root,
		indexes:    index,
	}
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

// Static returns a middleware handler that serves static files in the given directory.
func Serve(urlPrefix string, location string, index bool) func(next http.Handler) http.Handler {
	fs := LocalFile(location, index)
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if fs.Exists(urlPrefix, r.URL.Path) {

				ext := filepath.Ext(r.URL.Path)
				mime := mime.TypeByExtension(ext)
				w.Header().Set("Content-Type", mime)
				fmt.Println(mime, ext)
				fileserver.ServeHTTP(w, r)

			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
