package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

var (
	imageFiles    []string
	defaultWidth  = 800
	defaultHeight = 600
	minSize       = 10
	maxWidth      = 1920
	maxHeight     = 1080
)

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
	return width >= minSize && height >= minSize && width <= maxWidth && height <= maxHeight
}

func serveImage(c *gin.Context, width, height int) {
	img, format, err := getRandomImage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load image"})
		return
	}

	// Обрезаем и масштабируем до точных размеров (без растягивания)
	resizedImg := imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	// Кодируем в память
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
		err = jpeg.Encode(buf, resizedImg, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JPEG encoding failed"})
			return
		}
		c.Data(http.StatusOK, "image/jpeg", buf.Bytes())
	case "png":
		err = png.Encode(buf, resizedImg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PNG encoding failed"})
			return
		}
		c.Data(http.StatusOK, "image/png", buf.Bytes())
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unsupported image format"})
	}
}

func main() {
	r := gin.Default()

	// Редирект с `/` на `/800/600`
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/%d/%d", defaultWidth, defaultHeight))
	})

	// Обработка `/width/height`
	r.GET("/:width/:height", func(c *gin.Context) {
		width, err := strconv.Atoi(c.Param("width"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid width"})
			return
		}

		height, err := strconv.Atoi(c.Param("height"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height"})
			return
		}

		if !validateDimensions(width, height) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Dimensions must be between %dx%d and %dx%d", minSize, minSize, maxWidth, maxHeight),
			})
			return
		}

		serveImage(c, width, height)
	})

	r.Run(":8080")
}
