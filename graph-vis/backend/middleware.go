package main

import (
	arango "github.com/arangodb/go-driver"
	"github.com/labstack/echo/v4"
)

// TaxonSvcContext middleware makes TaxonSvc available in request context.
func TaxonSvcContext(db arango.Database) echo.MiddlewareFunc {
	taxonSvc := NewTaxonSvc(db)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("taxonSvc", taxonSvc)
			return next(c)
		}
	}
}
