package main

import (
	arango "github.com/arangodb/go-driver"
	"github.com/labstack/echo/v4"
)

// DBContext middleware makes db available in request context.
func DBContext(db arango.Database) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	}
}
