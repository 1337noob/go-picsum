package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

var imageFiles []string

func init() {
	rand.Seed(time.Now().UnixNano())
	loadImages()
}

func loadImages() {
	dir := "./images"
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error loading images:", err)
		os.Exit(1)
	}

	if len(imageFiles) == 0 {
		fmt.Println("No images found in ./images!")
		os.Exit(1)
	}
}

func getRandomImage() (image.Image, string, error) {
	randomFile := imageFiles[rand.Intn(len(imageFiles))]
	file, err := os.Open(randomFile)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	return img, format, err
}

func validateDimensions(width, height int) bool {
	return width >= 10 && height >= 10 && width <= 1920 && height <= 1080
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		// Парсим параметры
		width, err := strconv.Atoi(c.Query("width"))
		if err != nil {
			width = 0 // Автоподбор
		}

		height, err := strconv.Atoi(c.Query("height"))
		if err != nil {
			height = 0 // Автоподбор
		}

		// Валидация размеров
		if width != 0 || height != 0 {
			if !validateDimensions(width, height) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Dimensions must be between 10x10 and 1920x1080",
				})
				return
			}
		}

		// Получаем случайное изображение
		img, format, err := getRandomImage()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load image",
			})
			return
		}

		// Масштабируем (если нужно)
		if width > 0 || height > 0 {
			if width == 0 {
				width = img.Bounds().Dx()
			}
			if height == 0 {
				height = img.Bounds().Dy()
			}
			img = imaging.Resize(img, width, height, imaging.Lanczos)
		}

		// Отправляем изображение
		switch format {
		case "jpeg":
			c.DataFromReader(http.StatusOK, -1, "image/jpeg", readerFromImage(img), nil)
		case "png":
			c.DataFromReader(http.StatusOK, -1, "image/png", readerFromImage(img), nil)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Unsupported image format",
			})
		}
	})

	r.Run(":8080")
}

// Вспомогательная функция для преобразования image.Image в io.Reader
func readerFromImage(img image.Image) *io.PipeReader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		switch img.(type) {
		case *image.RGBA:
			jpeg.Encode(pw, img, nil)
		case *image.NRGBA:
			png.Encode(pw, img)
		}
	}()
	return pr
}
