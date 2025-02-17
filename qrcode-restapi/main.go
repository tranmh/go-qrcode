package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	qrcode "github.com/uncopied/go-qrcode"
)

func init() {
	initLog()
}

func initLog() {
	// setup logrus
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Starting time
		startTime := time.Now()

		// Processing request
		// ctx.Next() // bug?

		// End Time
		endTime := time.Now()

		// execution time
		latencyTime := endTime.Sub(startTime)

		// Request method
		reqMethod := ctx.Request.Method

		// Request route
		reqUri := ctx.Request.RequestURI

		// status code
		statusCode := ctx.Writer.Status()

		// Request IP
		clientIP := ctx.ClientIP()

		reqBody, _ := io.ReadAll(ctx.Request.Body)
		ctx.Request.Body = io.NopCloser(bytes.NewReader(reqBody))

		log.WithFields(log.Fields{
			"6_BODY":      string(reqBody),
			"5_METHOD":    reqMethod,
			"2_URI":       reqUri,
			"3_STATUS":    statusCode,
			"4_LATENCY":   latencyTime,
			"1_CLIENT_IP": clientIP,
		}).Info("HTTP REQUEST")

		ctx.Next()
	}
}

// https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 3:
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	return
}

func getQRCode(c *gin.Context) {

	var data string = c.Query("data")
	size, _ := strconv.Atoi(c.Query("size"))
	var ecc string = strings.ToUpper(c.Query("ecc"))
	var recovery_level qrcode.RecoveryLevel = qrcode.Low
	// margin, _ := strconv.Atoi(c.Query("margin")), no margin is supported yet in go-qrcode
	var color string = c.Query("color")
	var bgcolor string = c.Query("bgcolor")
	// qzone, no qzone is supported yet in go-qrcode
	var format string = strings.ToLower(c.Query("format"))

	if (strings.Compare(ecc, "L") == 0) ||
		(strings.Compare(ecc, "M") == 0) ||
		(strings.Compare(ecc, "Q") == 0) ||
		(strings.Compare(ecc, "H") == 0) {

		if strings.Compare(ecc, "L") == 0 {
			recovery_level = qrcode.Low
		} else if strings.Compare(ecc, "M") == 0 {
			recovery_level = qrcode.Medium
		} else if strings.Compare(ecc, "Q") == 0 {
			recovery_level = qrcode.High
		} else if strings.Compare(ecc, "H") == 0 {
			recovery_level = qrcode.Highest
		} else {
			log.Error("ECC aka RecoveryLevel is not supported")
		}
	}

	mycolor, _ := ParseHexColor(color)
	mybgcolor, _ := ParseHexColor(bgcolor)

	var q *qrcode.QRCode
	q, err := qrcode.New(data, recovery_level)
	checkError(err)
	q.ForegroundColor = mycolor
	q.BackgroundColor = mybgcolor

	if strings.Compare(format, "svg") == 0 {
		qrSVG, err := q.SVG()
		checkError(err)

		c.Header("Content-Disposition", "inline; filename=qrcode.svg")
		c.Data(http.StatusOK, "application/octet-stream", []byte(qrSVG))
	} else /* if strings.Compare(format, "eps") == 0 {
		qrEPS, err := q.EPS()
		checkError(err)

		c.Header("Content-Disposition", "inline; filename=qrcode.eps")
		c.Data(http.StatusOK, "application/octet-stream", []byte(qrEPS))
	} else if strings.Compare(format, "pdf") == 0 {
		qrPDF, err := q.PDF()
		checkError(err)

		c.Header("Content-Disposition", "inline; filename=qrcode.pdf")
		c.Data(http.StatusOK, "application/octet-stream", []byte(qrPDF))
	} else */{
		var png []byte
		png, err = q.PNG(size)
		checkError(err)

		c.Header("Content-Disposition", "inline; filename=qrcode.png")
		c.Data(http.StatusOK, "application/octet-stream", png)
	}
}

func main() {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(LoggingMiddleware())

	router.GET("/api/qrcode", getQRCode)

	router.Run(":6868")
}

func checkError(err error) {
	if err != nil {
		log.Error(err)
	}
}
