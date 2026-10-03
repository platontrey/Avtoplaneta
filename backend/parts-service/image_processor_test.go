package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptimizeAndSaveImage_Downscaling(t *testing.T) {
	// Create a mock large 3000x2000 image
	largeImg := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	for y := 0; y < 2000; y++ {
		for x := 0; x < 3000; x++ {
			largeImg.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, largeImg, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "img_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	targetFile := filepath.Join(tmpDir, "optimized.jpg")
	err = OptimizeAndSaveImage(&buf, targetFile)
	require.NoError(t, err)

	// Verify the saved file exists
	fileInfo, err := os.Stat(targetFile)
	require.NoError(t, err)
	assert.True(t, fileInfo.Size() > 0)

	// Read and verify dimensions are within 2048
	f, err := os.Open(targetFile)
	require.NoError(t, err)
	defer f.Close()

	cfg, format, err := image.DecodeConfig(f)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	assert.True(t, cfg.Width <= 2048)
	assert.True(t, cfg.Height <= 2048)
	assert.Equal(t, 2048, cfg.Width) // Width was max, should be exactly 2048
}

func TestOptimizeAndSaveImage_TransparentPNG(t *testing.T) {
	// Create a transparent PNG
	pngImg := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	// Draw transparent pixels and a red square in the center
	draw.Draw(pngImg, image.Rect(25, 25, 75, 75), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	var buf bytes.Buffer
	err := png.Encode(&buf, pngImg)
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "img_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	targetFile := filepath.Join(tmpDir, "from_png.jpg")
	err = OptimizeAndSaveImage(&buf, targetFile)
	require.NoError(t, err)

	// Read decoded image and check background corner (0,0) is white, not black
	f, err := os.Open(targetFile)
	require.NoError(t, err)
	defer f.Close()

	decoded, format, err := image.Decode(f)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", format)

	// Corner (0,0) should be close to white (due to JPEG lossy compression, allow small tolerance)
	r, g, b, _ := decoded.At(0, 0).RGBA()
	assert.True(t, r > 0xF000)
	assert.True(t, g > 0xF000)
	assert.True(t, b > 0xF000)
}
