package main

import (
	"fmt"
	"net/http"

	"github.com/coreservice-io/echo_middleware"
	"github.com/coreservice-io/echo_middleware/tool"
	"github.com/coreservice-io/log"
	"github.com/coreservice-io/logrus_log"
	"github.com/labstack/echo/v4"
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
			//write your own dealer here
			fmt.Println("todo:not working here")
			hs.Logger.Error(panic_err)
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
