package utils

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestImage creates a simple test image file
func createTestImage(t *testing.T, path string, width, height int) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}

	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	require.NoError(t, err)
}

func TestCompressImage_FileDoesNotExist(t *testing.T) {
	result, err := CompressImage("nonexistent.jpg", 1024)
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
}

func TestCompressImage_FileWithinLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	createTestImage(t, testFile, 100, 100)

	fileInfo, _ := os.Stat(testFile)
	maxSize := fileInfo.Size() + 1000

	result, err := CompressImage(testFile, maxSize)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result)

	fileInfoAfter, _ := os.Stat(testFile)
	assert.Equal(t, fileInfo.Size(), fileInfoAfter.Size())
}

func TestCompressImage_JpegCompression(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")

	createTestImage(t, testFile, 1000, 1000)

	maxSize := int64(1024)

	result, err := CompressImage(testFile, maxSize)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), result)

	fileInfoAfter, _ := os.Stat(testFile)
	// File should be compressed (smaller than original)
	assert.LessOrEqual(t, fileInfoAfter.Size(), int64(50000))

	assert.Equal(t, result, fileInfoAfter.Size())
}

func TestCompressImage_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bmp")

	err := os.WriteFile(testFile, []byte("BMP DATA"), 0644)
	require.NoError(t, err)

	result, err := CompressImage(testFile, 1024)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestCompressImage_PngFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	file, err := os.Create(testFile)
	require.NoError(t, err)
	err = png.Encode(file, img)
	require.NoError(t, err)
	file.Close()

	fileInfo, _ := os.Stat(testFile)
	maxSize := fileInfo.Size() + 1000

	result, err := CompressImage(testFile, maxSize)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestCompressImage_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.jpg")

	err := os.WriteFile(testFile, []byte{}, 0644)
	require.NoError(t, err)

	result, err := CompressImage(testFile, 1024)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestCompressImage_VerySmallMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jpg")
	createTestImage(t, testFile, 200, 200)

	maxSize := int64(100)

	result, err := CompressImage(testFile, maxSize)
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), result)

	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	_, _, err = image.Decode(file)
	assert.NoError(t, err)
}
