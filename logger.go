package EchoMiddleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/universe-30/ULog"
	"github.com/valyala/fasttemplate"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		//Skipper Skipper

		// Tags to construct the logger format.
		//
		// - time_unix
		// - time_unix_nano
		// - time_rfc3339
		// - time_rfc3339_nano
		// - time_custom
		// - id (Request ID)
		// - remote_ip
		// - uri
		// - host
		// - method
		// - path
		// - protocol
		// - referer
		// - user_agent
		// - status
		// - error
		// - latency (In nanoseconds)
		// - latency_human (Human readable)
		// - bytes_in (Bytes received)
		// - bytes_out (Bytes sent)
		// - header:<NAME>
		// - query:<NAME>
		// - form:<NAME>
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		Format string `yaml:"format"`

		// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
		CustomTimeFormat string `yaml:"custom_time_format"`

		Logger ULog.Logger

		RecordFailRequest bool

		template *fasttemplate.Template

		pool *sync.Pool
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		//Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
		//	`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		//	`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		//	`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		Format:           "=> ${method} | ${remote_ip} | ${host} | ${uri} | ${bytes_in} B in | ${status} | ${bytes_out} B out | ${latency_human} | ${error}",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		//colorer:          color.New(),
		Logger:            nil,
		RecordFailRequest: false,
	}
)

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc {

	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Logger == nil {
				return next(c)
			}

			if config.Logger.GetLevel() < ULog.DebugLevel && config.RecordFailRequest == false {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			//check status
			n := res.Status

			if n < 400 && config.Logger.GetLevel() < ULog.DebugLevel {
				return
			}

			stop := time.Now()
			buf := config.pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer config.pool.Put(buf)

			if _, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_unix":
					return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
				case "time_unix_nano":
					return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
				case "time_rfc3339":
					return buf.WriteString(time.Now().Format(time.RFC3339))
				case "time_rfc3339_nano":
					return buf.WriteString(time.Now().Format(time.RFC3339Nano))
				case "time_custom":
					return buf.WriteString(time.Now().Format(config.CustomTimeFormat))
				case "id":
					id := req.Header.Get(echo.HeaderXRequestID)
					if id == "" {
						id = res.Header().Get(echo.HeaderXRequestID)
					}
					return buf.WriteString(id)
				case "remote_ip":
					return buf.WriteString(c.RealIP())
				case "host":
					return buf.WriteString(req.Host)
				case "uri":
					return buf.WriteString(req.RequestURI)
				case "method":
					return buf.WriteString(req.Method)
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "protocol":
					return buf.WriteString(req.Proto)
				case "referer":
					return buf.WriteString(req.Referer())
				case "user_agent":
					return buf.WriteString(req.UserAgent())
				case "status":
					return buf.WriteString(strconv.Itoa(n))
				case "error":
					if err != nil {
						// Error may contain invalid JSON e.g. `"`
						b, _ := json.Marshal(err.Error())
						b = b[1 : len(b)-1]
						return buf.Write(b)
					}
				case "latency":
					l := stop.Sub(start)
					return buf.WriteString(strconv.FormatInt(int64(l), 10))
				case "latency_human":
					return buf.WriteString(stop.Sub(start).String())
				case "bytes_in":
					cl := req.Header.Get(echo.HeaderContentLength)
					if cl == "" {
						cl = "0"
					}
					return buf.WriteString(cl)
				case "bytes_out":
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				default:
					switch {
					case strings.HasPrefix(tag, "header:"):
						return buf.Write([]byte(c.Request().Header.Get(tag[7:])))
					case strings.HasPrefix(tag, "query:"):
						return buf.Write([]byte(c.QueryParam(tag[6:])))
					case strings.HasPrefix(tag, "form:"):
						return buf.Write([]byte(c.FormValue(tag[5:])))
					case strings.HasPrefix(tag, "cookie:"):
						cookie, err := c.Cookie(tag[7:])
						if err == nil {
							return buf.Write([]byte(cookie.Value))
						}
					}
				}
				return 0, nil
			}); err != nil {
				return
			}

			if config.RecordFailRequest == true {
				if n >= 400 {
					config.Logger.Errorln(buf.String())
				} else {
					config.Logger.Debugln(buf.String())
				}
			} else {
				config.Logger.Debugln(buf.String())
			}
			return
		}
	}
}
