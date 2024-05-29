// Package main is the entry point for the application.
package main

import (
	"net/http"

	"go/screensaver/layout"

	"github.com/gin-gonic/gin"
)

// main function sets up the Gin router, registers routes, and starts the server.
func main() {
	// Create a default Gin router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("templates/index.tmpl")

	// Route for the main page
	router.GET("/", func(c *gin.Context) {
		// Get the layout name from query parameters
		layout_name := c.Query("layout")
		// Render the HTML page with the specified layout
		c.HTML(http.StatusOK, "index.tmpl", layout.GetLayout(layout_name))
	})

	// Serve static files
	router.StaticFile("/design", "static/design.html")
	router.StaticFile("/layouts", "static/layouts.html")
	router.StaticFile("/tailwind.css", "static/tailwind.css")
	router.StaticFile("/index.js", "static/js/index.js")
	router.Static("/static/packages", "static/packages")

	// Register API routes
	RegisterApiRoutes(router)

	// Start the server on port 8080
	router.Run(":8080") // Default port for Gin applications
}
