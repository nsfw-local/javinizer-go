package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	_ "image/png" // Register PNG decoder
)

// CropPosterFromCover crops the right 47.2% of a cover image to create a poster
// This matches the original Javinizer's crop.py implementation which uses:
// left = width / 1.895734597 (which keeps the right 47.2%)
func CropPosterFromCover(coverPath, posterPath string) error {
	// Open and decode the cover image
	coverFile, err := os.Open(coverPath)
	if err != nil {
		return fmt.Errorf("failed to open cover image: %w", err)
	}
	defer coverFile.Close()

	img, format, err := image.Decode(coverFile)
	if err != nil {
		return fmt.Errorf("failed to decode cover image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate crop coordinates
	// left = width / 1.895734597 keeps the right 47.2% of the image
	left := int(float64(width) / 1.895734597)
	top := 0
	right := width
	bottom := height

	// Create cropped image
	croppedBounds := image.Rect(0, 0, right-left, bottom-top)
	cropped := image.NewRGBA(croppedBounds)

	// Copy the cropped portion
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			cropped.Set(x-left, y-top, img.At(x, y))
		}
	}

	// Create output file
	posterFile, err := os.Create(posterPath)
	if err != nil {
		return fmt.Errorf("failed to create poster file: %w", err)
	}
	defer posterFile.Close()

	// Encode as JPEG with high quality
	opts := &jpeg.Options{Quality: 95}
	if err := jpeg.Encode(posterFile, cropped, opts); err != nil {
		return fmt.Errorf("failed to encode poster image: %w", err)
	}

	// Log format for debugging
	_ = format

	return nil
}
