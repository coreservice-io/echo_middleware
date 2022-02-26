# EchoMiddleware

middleware for echo server

### usage
```go
import (
	"github.com/coreservice-io/EchoMiddleware"  // logger and panic handle
	"github.com/coreservice-io/EchoMiddleware/tool"  // use Jsoniter as json parser
)
```

### example

```go
package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/coreservice-io/EchoMiddleware"
	"github.com/coreservice-io/EchoMiddleware/tool"
	"github.com/universe-30/LogrusULog"
	"github.com/coreservice-io/ULog"
)

func main() {
	// need a logger which implements ULog.Logger interface
	// here use LogrusULog as example
	logger, _ := LogrusULog.New("./logs", 2, 20, 30)
	logger.SetLevel(ULog.DebugLevel)

	// new echo instance
	hs := echo.New()

	//use Jsoniter as json parser
	hs.JSONSerializer = tool.NewJsoniter()

	//use logger middleware
	hs.Use(EchoMiddleware.LoggerWithConfig(EchoMiddleware.LoggerConfig{
		Logger: logger, //Logger interface
	}))

	//use recover and panic handler
	hs.Use(EchoMiddleware.RecoverWithConfig(EchoMiddleware.RecoverConfig{
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