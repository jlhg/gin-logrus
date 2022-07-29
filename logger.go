package ginlogrus

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger is the logrus logger handler
func Logger(logger logrus.FieldLogger, notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		start := time.Now()

		// Process request
		c.Next()

		stop := time.Now()
		latency := stop.Sub(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		respBodySize := c.Writer.Size()
		if respBodySize < 0 {
			respBodySize = 0
		}

		reqBody := ""
		if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
			var buf bytes.Buffer

			tee := io.TeeReader(c.Request.Body, &buf)
			bodyBytes, _ := ioutil.ReadAll(tee)
			c.Request.Body = ioutil.NopCloser(&buf)
			reqBody = string(bodyBytes)
		}

		if _, ok := skip[path]; ok {
			return
		}

		entry := logger.WithFields(logrus.Fields{
			"hostname":         hostname,
			"statusCode":       statusCode,
			"latency":          latency,
			"clientIP":         clientIP,
			"method":           c.Request.Method,
			"path":             path,
			"rawQuery":         rawQuery,
			"referer":          referer,
			"requestBody":      reqBody,
			"responseBodySize": respBodySize,
			"userAgent":        clientUserAgent,
			"error":            c.Errors.ByType(gin.ErrorTypePrivate).String(),
		})

		entry.Info()
	}
}
