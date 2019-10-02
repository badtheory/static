// Based on https://github.com/gin-contrib/static

package static

import (
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"fmt"
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
func Serve(urlPrefix string, fs ServeFileSystem) func(next http.Handler) http.Handler {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if fs.Exists(urlPrefix, r.URL.Path) {

				p := strings.TrimPrefix(r.URL.Path, urlPrefix)

				//Open the file
				f, _ := fs.Open(p)
				defer f.Close()

				//Create a buffer to store the header of the file in
				FileHeader := make([]byte, 512)

				//Copy the headers into the FileHeader buffer
				f.Read(FileHeader)

				//Get the file size
				FileStat, _ := f.Stat()                            //Get info from file
				FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

				//We read 512 bytes from the file already, so we reset the offset back to 0
				f.Seek(0, 0)

				w.Header().Set("Content-Type", http.DetectContentType(FileHeader))
				fmt.Println(http.DetectContentType(FileHeader))
				w.Header().Set("Content-Length", FileSize)

				fileserver.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
