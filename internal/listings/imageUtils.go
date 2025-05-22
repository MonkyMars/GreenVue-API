package listings

import (
	"bytes"
	"image"
	_ "image/jpeg" // register JPEG format
	_ "image/png"  // register PNG format
	"io"
	"log"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

// convertToWebP converts any image to WebP format
func convertToWebP(reader io.Reader) (*bytes.Buffer, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		log.Println("Error decoding image:", err)
		log.Println("Image format:", format)
		return nil, err
	}

	// Resize while maintaining aspect ratio (max 640px)
	img = resize.Resize(0, 640, img, resize.Lanczos3)

	// Encode to WebP
	webpBuffer := new(bytes.Buffer)

	webpOptions := &webp.Options{Quality: 80}
	err = webp.Encode(webpBuffer, img, webpOptions)

	if err != nil {
		log.Println("Error encoding WebP:", err)
		return nil, err
	}

	// Return the WebP buffer
	return webpBuffer, nil
}
