package main

import (
	http "net/http"

	echo "github.com/labstack/echo/v4"
)

type JSONResp map[string]interface{}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, JSONResp{"error": msg})
}

// TaxonGet serves JSON response containing a single taxon by ID.
func TaxonGet(c echo.Context) (err error) {
	taxSvc := c.Get("TaxonSvc").(TaxonSvc)
	rank := c.Param("rank")
	id := c.Param("id")
	if id == "" {
		return badRequest(c, "Missing taxon ID")
	}
	taxon, err := taxSvc.Get(rank, id)
	if err != nil {
		return badRequest(c, err.Error())
	}
	return c.JSON(http.StatusOK, JSONResp{"data": taxon})
}
