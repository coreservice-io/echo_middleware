package EchoMiddleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/universe-30/ULog"
)

type (
	// RecoverConfig defines the config for Recover middleware.
	RecoverConfig struct {
		Logger  ULog.Logger
		OnPanic func(panic_err interface{})
	}
)

// RecoverWithConfig returns a Recover middleware with config.
// See: `Recover()`.
func RecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					if config.OnPanic != nil {
						config.OnPanic(r)
					}

					if config.Logger != nil {
						msg := fmt.Sprintf("[PANIC RECOVER] %s", r)
						config.Logger.Errorln(msg)
					}

					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}