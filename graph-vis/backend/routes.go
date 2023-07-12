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
	taxSvc := c.Get("taxonSvc").(*TaxonSvc)
	rank := c.Param("rank")
	id := c.Param("id")
	if id == "" {
		return badRequest(c, "Missing taxon ID")
	}
	taxon, err := taxSvc.Get(rank, id)
	if err != nil {
		return badRequest(c, err.Error())
	}
	return c.JSON(http.StatusOK, JSONResp{"data": TaxonResponse(taxon)})
}

// TaxonGetChildren serves JSON response containing a list of taxon children.
func TaxonGetChildren(c echo.Context) (err error) {
	taxSvc := c.Get("taxonSvc").(*TaxonSvc)
	rank := c.Param("rank")
	id := c.Param("id")
	if id == "" {
		return badRequest(c, "Missing taxon ID")
	}
	taxa, err := taxSvc.GetChildren(rank, id)
	if err != nil {
		return badRequest(c, err.Error())
	}
	var taxaResp []TaxonResponse
	for _, taxon := range taxa {
		taxaResp = append(taxaResp, TaxonResponse(taxon))
	}
	return c.JSON(http.StatusOK, JSONResp{"data": taxaResp})
}
