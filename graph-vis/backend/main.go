package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load config
	cfg, err := LoadConfig("./app.env")
	if err != nil {
		panic(err)
	}

	// Init database
	db, err := GetArangoDB(cfg)
	if err != nil {
		panic(err)
	}

	// _, err := http.Get("https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// Echo instance
	router := echo.New()

	// Middleware
	router.Pre(middleware.RemoveTrailingSlash())
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	router.Use(TaxonSvcContext(db))
	// router.Use(auth.ParseJWT)

	// Routes
	api := router.Group("/api/v1")
	{
		taxon := api.Group("/taxon")
		{
			taxon.GET("/:rank/:id/children", TaxonGetChildren)
			taxon.GET("/:rank/:id", TaxonGet)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	router.Logger.Fatal(router.Start(fmt.Sprintf(":%s", port)))
}
