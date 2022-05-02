# echo_middleware

middleware for echo server

### usage
```go
import (
	"github.com/coreservice-io/echo_middleware"  // logger and panic handle
	"github.com/coreservice-io/echo_middleware/tool"  // use Jsoniter as json parser
)
```

### example

```go
package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/coreservice-io/echo_middleware"
	"github.com/coreservice-io/echo_middleware/tool"
	"github.com/universe-30/logrus_log"
	"github.com/coreservice-io/log"
)

func main() {
	// need a logger which implements log.Logger interface
	// here use logrus_log as example
	logger, _ := logrus_log.New("./logs", 2, 20, 30)
	logger.SetLevel(log.DebugLevel)

	// new echo instance
	hs := echo.New()

	//use Jsoniter as json parser
	hs.JSONSerializer = tool.NewJsoniter()

	//use logger middleware
	hs.Use(echo_middleware.LoggerWithConfig(echo_middleware.LoggerConfig{
		Logger: logger, //Logger interface
	}))

	//use recover and panic handler
	hs.Use(echo_middleware.RecoverWithConfig(echo_middleware.RecoverConfig{
		// callback to handler panic
		OnPanic: func(panic_err interface{}) {
			log.Println(panic_err)
		},
	}))

	///////////////////  JSONP //////////////////////
	hs.GET("/test1", func(c echo.Context) error {
		return c.String(http.StatusOK, "success test1")
	})

	hs.GET("/test2", func(c echo.Context) error {
		//example panic happen
		a := 1
		_ = 1 / (a - 1)

		return c.String(http.StatusOK, "success test2")
	})

	// start server
	hs.Start(":8080")
}
```