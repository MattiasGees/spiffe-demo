package customer

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html
var content embed.FS

func (c *CustomerService) webpageHandler(w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(content, "index.html")
	if err != nil {
		http.Error(w, "Could not read requested file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}
