// Package layout provides functionalities for handling layout data structures
// and registering related routes.
package layout

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

// Layout represents the structure of a layout with its name, dimensions, and widgets.
type Layout struct {
	Name    string   `json:"name"`
	Rows    int      `json:"rows"`
	Cols    int      `json:"cols"`
	Widgets []Widget `json:"widgets"`
}

// // RegisterLayoutRoutes registers the routes for saving and loading layouts.
// func RegisterLayoutRoutes(r *gin.Engine) {
// 	// Route for saving a layout
// 	r.POST("/saveLayout", func(c *gin.Context) {
// 		var layout Layout
// 		// Bind JSON payload to layout
// 		if err := c.ShouldBindJSON(&layout); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		// Save layout to file
// 		if err := saveLayout(layout); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.Status(http.StatusOK)
// 	})

// 	// Route for loading a layout
// 	r.GET("/loadLayout", func(c *gin.Context) {
// 		name := c.Query("name")
// 		// Load layout from file
// 		layout, err := loadLayout(name)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusOK, layout)
// 	})
// }

// saveLayout saves the given layout to a JSON file.
// func saveLayout(layout Layout) error {
// 	data, err := json.Marshal(layout)
// 	if err != nil {
// 		return err
// 	}
// 	return os.WriteFile("layouts/"+layout.Name+".json", data, 0644)
// }

// loadLayout loads the layout with the given name from a JSON file.
func loadLayout(name string) (Layout, error) {
	var layout Layout
	data, err := os.ReadFile("layouts/" + name + ".json")
	if err != nil {
		return layout, err
	}
	err = json.Unmarshal(data, &layout)
	return layout, err
}

// GetLayout retrieves and processes the layout with the given name,
// returning a map of its properties for rendering.
func GetLayout(layoutName string) map[string]interface{} {
	var renderedWidgets []map[string]interface{}

	if layoutName == "" {
		layoutName = "default"
	}
	layout, err := loadLayout(layoutName)
	if err != nil {
		log.Fatalf("Unable to load layout %s", layoutName)
		return map[string]interface{}{
			"nrow":    0,
			"ncol":    0,
			"gap":     "16px",
			"margin":  "16px",
			"widgets": renderedWidgets,
		}
	}
	for _, widget := range layout.Widgets {
		if ok := widget.DataCheck(); !ok {
			log.Fatalf("Data Check failed %s", layoutName)
		}
		lrow := 1
		lcol := 1
		// Set widget dimensions based on its size
		switch widget.Size {
		case MiddleV:
			lrow = 2
		case MiddleH:
			lcol = 2
		case Large:
			lrow = 2
			lcol = 2
		case LongH:
			lcol = 4
		case LongV:
			lrow = 4
		}
		// Append widget properties to the rendered widgets slice
		renderedWidgets = append(renderedWidgets, map[string]interface{}{
			"irow":          widget.Row,
			"lrow":          lrow,
			"icol":          widget.Col,
			"lcol":          lcol,
			"padding":       template.CSS("calc(var(--cell-size) * 0.1)"),
			"showUpdateBtn": ShowUpdateBtn(widget.Type),
			"wgcontent":     template.HTML(widget.RenderContent()),
			"widgetId":      widget.GetId(),
			"wgquery":       fmt.Sprintf("size=%s&%s", widget.Size, mapToQueryString(widget.Data)),
			"wgtype":        widget.Type,
		})
	}
	return map[string]interface{}{
		"nrow":    layout.Rows,
		"ncol":    layout.Cols,
		"gap":     "16px",
		"margin":  "16px",
		"widgets": renderedWidgets,
	}
}

// mapToQueryString converts a map to a query string format.
func mapToQueryString(m map[string]interface{}) string {
	var sb strings.Builder
	first := true
	for key, value := range m {
		if !first {
			sb.WriteString("&")
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", value))
		first = false
	}
	return sb.String()
}
