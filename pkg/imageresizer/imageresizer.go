package imageresizer

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/patrickmn/go-cache"

	"ipfsimageresizer/pkg/ipfs"
)

type ImageHandler struct {
	cache       *cache.Cache
	ipfsHandler *ipfs.IpfsHandler
}

func New(ipfs *ipfs.IpfsHandler, cacheExpiration, cacheCleanupInterval time.Duration) *ImageHandler {
	return &ImageHandler{
		cache:       cache.New(cacheExpiration, cacheCleanupInterval),
		ipfsHandler: ipfs,
	}
}

func (handler *ImageHandler) GetImageData(cid string, widthStr, heightStr string) ([]byte, string, error) {
	width, err := strconv.Atoi(widthStr)
	if err != nil || width <= 0 {
		return nil, "", fmt.Errorf("couldnt parse height %v", err)
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height <= 0 {
		return nil, "", fmt.Errorf("couldnt parse height %v", err)
	}

	cacheKey := fmt.Sprintf("%s_%dx%d", cid, width, height)

	if cachedData, found := handler.getAndResetCachedImage(cacheKey); found {
		return cachedData.Bytes, cachedData.Format, nil
	}

	imageData, err := handler.ipfsHandler.FetchFromIPFPLocal(cid)
	if err != nil {
		log.Printf("Error fetching from IPFS: %v", err)
		return nil, "", err
	}

	resizedData, format, err := resizeImage(imageData, width, height)
	if err != nil {
		log.Printf("Error resizing image: %v", err)
		return nil, "", err
	}

	handler.setCachedImage(cacheKey, resizedData, format)

	return resizedData, format, nil
}

type cachedImage struct {
	Bytes  []byte
	Format string
}

func (handler *ImageHandler) getAndResetCachedImage(key string) (*cachedImage, bool) {
	data, found := handler.cache.Get(key)
	if !found {
		return nil, false
	}
	imageData, ok := data.(map[string]interface{})
	if !ok {
		return nil, false
	}
	bytes, ok1 := imageData["bytes"].([]byte)
	format, ok2 := imageData["format"].(string)
	if !ok1 || !ok2 {
		return nil, false
	}

	// Reset cache timeout
	handler.cache.Set(key, imageData, cache.DefaultExpiration)

	return &cachedImage{
		Bytes:  bytes,
		Format: format,
	}, true
}

func (handler *ImageHandler) setCachedImage(key string, data []byte, format string) {
	handler.cache.Set(key, map[string]interface{}{
		"bytes":  data,
		"format": format,
	}, cache.DefaultExpiration)
}

func resizeImage(data []byte, width, height int) ([]byte, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("image decode error: %w", err)
	}

	resizedImg := imaging.Fit(img, width, height, imaging.Lanczos)

	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resizedImg)
	case "gif":
		err = gif.Encode(&buf, resizedImg, nil)
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return nil, "", fmt.Errorf("image encode error: %w", err)
	}

	return buf.Bytes(), format, nil
}
