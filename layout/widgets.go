// Package layout provides functionalities for handling widget layouts and rendering them.
package layout

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"slices"

	"github.com/gin-gonic/gin"
)

// WidgetSize represents the size of a widget.
type WidgetSize string

// Constants for different widget sizes.
const (
	Small   WidgetSize = "small"
	MiddleH WidgetSize = "middleh"
	MiddleV WidgetSize = "middlev"
	Large   WidgetSize = "large"
	LongH   WidgetSize = "longh"
	LongV   WidgetSize = "longv"
)

// Widget represents the structure of a widget with its type, size, position, and data.
type Widget struct {
	Type WidgetType             `json:"type"`
	Size WidgetSize             `json:"size"`
	Row  int64                  `json:"row"`
	Col  int64                  `json:"col"`
	Data map[string]interface{} `json:"data"`
}

// GetId returns a unique identifier for the widget based on its type, row, and column.
func (w Widget) GetId() string {
	return fmt.Sprintf("wg-%s-r%d-c%d", w.Type, w.Row, w.Col)
}

// RenderFromTemplate renders the widget using the specified template name.
func (w Widget) RenderFromTemplate(tmplName string) string {
	if !SizeCheck(w.Type, w.Size) {
		log.Printf("Invalid size %s for %s Widget", w.Size, tmplName)
		return ""
	}
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{"getIndexRange": GetIndexRange}).ParseFiles(fmt.Sprintf("templates/widgets/%s.tmpl", tmplName))
	if err != nil {
		log.Fatalf("Failed to find a template for %s %s Widget: %v", w.Size, tmplName, err)
		return ""
	}
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, string(w.Size), gin.H{
		"widgetId": w.GetId(),
	})
	if err != nil {
		log.Fatalf("Template execution failed for %s %s Widget: %v", w.Size, tmplName, err)
		return ""
	}
	return buf.String()
}

// GetIndexRange returns a slice of integers from 0 to the specified value.
func GetIndexRange(to int) []int {
	retVal := []int{}
	for i := 0; i < to; i++ {
		retVal = append(retVal, i)
	}
	return retVal
}

// WidgetType represents the type of a widget.
type WidgetType string

// Constants for different widget types.
const (
	WeatherForecastWidget WidgetType = "weatherforecast"
	NotionCalendarWidget  WidgetType = "notioncalendar"
	ClockWidget           WidgetType = "clock"
)

// RenderContent renders the content of the widget based on its type.
func (w Widget) RenderContent() string {
	switch w.Type {
	case WeatherForecastWidget:
		return w.RenderFromTemplate("weatherforecast")
	case NotionCalendarWidget:
		return w.RenderFromTemplate("notioncalendar")
	case ClockWidget:
		return w.RenderFromTemplate("clock")
	}
	return "Not Implemented"
}

// DataCheck validates the data of the widget based on its type.
func (w Widget) DataCheck() bool {
	switch w.Type {
	case WeatherForecastWidget:
		check := true
		_, ok := w.Data["location_name"].(string)
		check = check && ok
		_, ok = w.Data["location_latitude"].(float64)
		check = check && ok
		_, ok = w.Data["location_longitude"].(float64)
		check = check && ok
		return check
	case NotionCalendarWidget:
		check := true
		// _, ok := w.Data["database_id"].(string)
		// check = check && ok
		return check
	case ClockWidget:
		return true
	}
	return false
}

// SizeCheck validates if the widget size is supported for the given widget type.
func SizeCheck(wgtype WidgetType, size WidgetSize) bool {
	var supportedSize []WidgetSize
	switch wgtype {
	case WeatherForecastWidget:
		supportedSize = []WidgetSize{Small, MiddleV}
	case NotionCalendarWidget:
		supportedSize = []WidgetSize{MiddleV, LongV, MiddleH}
	case ClockWidget:
		supportedSize = []WidgetSize{MiddleH}
	}
	return slices.Contains(supportedSize, size)
}

// ShowUpdateBtn determines if the update button should be shown for the given widget type.
func ShowUpdateBtn(wgtype WidgetType) bool {
	switch wgtype {
	case WeatherForecastWidget:
		return true
	case NotionCalendarWidget:
		return true
	case ClockWidget:
		return false
	}
	return true
}
