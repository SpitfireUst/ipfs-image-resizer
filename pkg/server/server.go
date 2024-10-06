package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"ipfsimageresizer/pkg/imageresizer"
	"ipfsimageresizer/pkg/ipfs"
)

type CdnResizer struct {
	resizer *imageresizer.ImageHandler
}

func New(cacheExpiration, cacheCleanupInterval time.Duration) *CdnResizer {
	ipfs := ipfs.New()
	imageResizer := imageresizer.New(ipfs, cacheExpiration, cacheCleanupInterval)

	return &CdnResizer{
		resizer: imageResizer,
	}
}

func (cdn *CdnResizer) Start(port int) {
	http.HandleFunc("/image", cdn.imageHandler)

	log.Printf("Starting server on port %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func (cdn *CdnResizer) imageHandler(w http.ResponseWriter, r *http.Request) {
	cid := r.URL.Query().Get("cid")
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")

	if cid == "" || widthStr == "" || heightStr == "" {
		http.Error(w, "Missing required query parameters: cid, width, height", http.StatusBadRequest)
		return
	}

	resizedData, format, err := cdn.resizer.GetImageData(cid, widthStr, heightStr)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	serveImage(w, resizedData, format)
}

func serveImage(w http.ResponseWriter, data []byte, format string) {
	var contentType string
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "gif":
		contentType = "image/gif"
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(data)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
