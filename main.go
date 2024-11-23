package main

import (
	"fmt"
	"log"
	"math"
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

	lighterColor := getMediumLighterShade(color)
	slightlyDarkerForDots := getSlightlyDarkerShade(color)

	svg := fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
		<defs>
			<linearGradient id="mainGrad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
				<stop offset="0%%" style="stop-color:#%s;stop-opacity:1" />
				<stop offset="100%%" style="stop-color:#%s;stop-opacity:1" />
			</linearGradient>
			<pattern id="dots" width="20" height="20" patternUnits="userSpaceOnUse">
				<circle cx="10" cy="10" r="1.2" fill="#%s" opacity="0.4"/>
			</pattern>
			<filter id="softShadow">
				<feGaussianBlur in="SourceAlpha" stdDeviation="2"/>
				<feOffset dx="0" dy="1" result="offsetblur"/>
				<feFlood flood-color="#000000" flood-opacity="0.2"/>
				<feComposite in2="offsetblur" operator="in"/>
				<feMerge>
					<feMergeNode/>
					<feMergeNode in="SourceGraphic"/>
				</feMerge>
			</filter>
		</defs>
		<rect width="%d" height="%d" fill="url(#mainGrad)"/>
		<rect width="%d" height="%d" fill="url(#dots)"/>
		<text x="50%%" y="50%%" text-anchor="middle" dominant-baseline="middle" 
			font-family="Arial, Helvetica, sans-serif" font-weight="bold" font-size="24" 
			fill="%s" filter="url(#softShadow)">%dx%d</text>
	</svg>`, w, h, w, h, color, lighterColor, slightlyDarkerForDots,
		w, h, w, h,
		getContrastColor(color), w, h)

	c.Header("Content-Type", "image/svg+xml")
	if isDev() {
		c.Header("Cache-Control", "no-cache")
	} else {
		c.Header("Cache-Control", "public, max-age=86400")
	}
	c.String(200, svg)
}

// for the main gradient
func getMediumLighterShade(hexColor string) string {
	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	factor := 1.20 // 20% lighter
	newR := int(math.Min(float64(r)*factor, 255))
	newG := int(math.Min(float64(g)*factor, 255))
	newB := int(math.Min(float64(b)*factor, 255))

	return fmt.Sprintf("%02x%02x%02x", newR, newG, newB)
}

// for dots - slightly darker than base color
func getSlightlyDarkerShade(hexColor string) string {
	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	factor := 0.85 // 15% darker
	newR := int(float64(r) * factor)
	newG := int(float64(g) * factor)
	newB := int(float64(b) * factor)

	return fmt.Sprintf("%02x%02x%02x", newR, newG, newB)
}

func getContrastColor(hexColor string) string {
	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255

	if luminance > 0.5 {
		return "#000000"
	}
	return "#ffffff"
}

func isDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}
