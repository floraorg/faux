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
	borderRadius := c.DefaultQuery("r", "0")
	dot := c.DefaultQuery("d", "false")
	grad := c.DefaultQuery("g", "false")
	text := c.DefaultQuery("t", "false")

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

	radius, err := strconv.Atoi(borderRadius)
	if err != nil || radius < 0 || radius > min(w/2, h/2) {
		radius = 0
	}

	if color == "" {
		color = "333333"
	}

	if color[0] == '#' {
		color = color[1:]
	}

	if !(len(color) == 6 || len(color) == 3) {
		c.String(http.StatusBadRequest, "Invalid color format")
		return
	}

	if len(color) == 3 {
		c2 := color
		color = ""
		for i := 0; i < len(c2); i++ {
			color += string(c2[i]) + string(c2[i]) // Repeat each character
		}
	}
	lighterColor := color
	slightlyDarkerForDots := color
	dx := 0
	if grad == "true" {
		lighterColor = getMediumLighterShade(color)
	}
	if dot == "true" {
		slightlyDarkerForDots = getSlightlyDarkerShade(color)
		dx = 20
	}

	textColor := "#ffffff00"
	if text == "true" {
		textColor = getContrastColor(color)
	}

	svg := fmt.Sprintf(`<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
		<defs>
			<linearGradient id="mainGrad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
				<stop offset="0%%" style="stop-color:#%s;stop-opacity:1" />
				<stop offset="100%%" style="stop-color:#%s;stop-opacity:1" />
			</linearGradient>
			<pattern id="dots" width="%d" height="%d" patternUnits="userSpaceOnUse">
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
			<mask id="roundedMask">
				<rect width="%d" height="%d" rx="%d" ry="%d" fill="white"/>
			</mask>
		</defs>
		<g mask="url(#roundedMask)">
			<rect width="%d" height="%d" fill="url(#mainGrad)"/>
			<rect width="%d" height="%d" fill="url(#dots)"/>
		</g>
		<text x="50%%" y="50%%" text-anchor="middle" dominant-baseline="middle" 
			font-family="Arial, Helvetica, sans-serif" font-weight="bold" font-size="24" 
			fill="%s" filter="url(#softShadow)">%dx%d</text>
	</svg>`, w, h, w, h, color, lighterColor, dx, dx, slightlyDarkerForDots,
		w, h, radius, radius,
		w, h, w, h,
		textColor, w, h)

	c.Header("Content-Type", "image/svg+xml")
	if isDev() {
		c.Header("Cache-Control", "no-cache")
	} else {
		c.Header("Cache-Control", "public, max-age=86400")
	}
	c.String(200, svg)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// for the main gradient
func getMediumLighterShade(hexColor string) string {
	r, _ := strconv.ParseInt(hexColor[0:2], 16, 64)
	g, _ := strconv.ParseInt(hexColor[2:4], 16, 64)
	b, _ := strconv.ParseInt(hexColor[4:6], 16, 64)

	brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	fmt.Println(brightness)

	var factor float64
	if brightness < 10 {
		factor = 60.0
	} else if brightness < 30 {
		factor = 3.0
	} else if brightness < 60 {
		factor = 2.0
	} else if brightness < 100 {
		factor = 1.75
	} else if brightness < 150 {
		factor = 1.5
	} else {
		factor = 0.8
	}

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

	brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

	var factor float64
	if brightness < 80 {
		factor = 1.8
	} else {
		factor = 0.55
	}
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
		return "#111111"
	}
	return "#eeeeee"
}

func isDev() bool {
	return os.Getenv("ENVIRONMENT") == "DEV"
}
