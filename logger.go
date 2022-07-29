package ginlogrus

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TimeFormat represents time format layout.
var TimeFormat = time.RFC3339

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
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if _, ok := skip[path]; ok {
			return
		}

		entry := logger.WithFields(logrus.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency,
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"rawQuery":   rawQuery,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
			"time":       start.Format(TimeFormat),
			"error":      c.Errors.ByType(gin.ErrorTypePrivate).String(),
		})

		entry.Info()
	}
}
