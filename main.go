package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := setupRouter()
	fmt.Printf("Server starting on port :%s\n", port)
	r.Run(":" + port)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("views/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/:width/:height", generatePlaceholder)
	r.GET("/:width/:height/:color", generatePlaceholder)

	return r
}

func generatePlaceholder(c *gin.Context) {
	width := c.Param("width")
	height := c.Param("height")
	color := c.Param("color")

	w, err := strconv.Atoi(width)
	if err != nil || w <= 0 || w > 3000 {
		c.String(http.StatusBadRequest, "Invalid width")
		return
	}

	h, err := strconv.Atoi(height)
	if err != nil || h <= 0 || h > 3000 {
		c.String(http.StatusBadRequest, "Invalid height")
		return
	}

	if color == "" {
		color = "999999"
	}

	if color[0] == '#' {
		color = color[1:]
	}

	if len(color) != 6 {
		c.String(http.StatusBadRequest, "Invalid color format")
		return
	}

	svg := fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" 
		xmlns="http://www.w3.org/2000/svg">
		<rect width="%d" height="%d" fill="#%s"/>
		<text x="50%%" y="50%%" text-anchor="middle" dominant-baseline="middle" 
			font-family="Arial, Helvetica, sans-serif" font-size="16" fill="%s">
			%dx%d
		</text>
	</svg>`, w, h, w, h, w, h, color,
		getContrastColor(color), w, h)

	c.Header("Content-Type", "image/svg+xml")
	if isDev() {
		c.Header("Cache-Control", "no-cache")
	} else {
		c.Header("Cache-Control", "public, max-age=86400") // cache for 24 hours
	}
	c.String(200, svg)
}

func getContrastColor(hexColor string) string {
	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255

	if luminance > 0.5 {
		return "#000000" // use black text for light backgrounds
	}
	return "#ffffff" // use white text for dark backgrounds
}

func isDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}
