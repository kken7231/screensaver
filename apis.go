// Package main is the entry point for the application.
package main

import (
	"bytes"
	"cmp"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"time"

	"go/screensaver/layout"
	"go/screensaver/notion"
	"go/screensaver/util"
	"go/screensaver/weather"

	"github.com/gin-gonic/gin"
)

// DrawHorizontalLines generates the HTML content for horizontal time lines in the calendar widget.
func DrawHorizontalLines(minHours, maxHours int, now time.Time) (string, error) {
	var buf bytes.Buffer
	var err error
	var minNow int
	var initialVisibility string
	nRow := maxHours - minHours + 1

	// Parse the template file for the calendar widget.
	tmpl, err := template.New("eventTmpl").ParseFiles("templates/util.tmpl")
	if err != nil {
		err = fmt.Errorf("failed to find a template for notioncalendar Widget: %v", err)
		goto api_drawhorizontallines_finish
	}

	// Calculate the current minute within the specified hour range.
	minNow = (now.Hour()-minHours)*60 + now.Minute()
	initialVisibility = "visible"
	if minNow < minHours*60 || minNow > maxHours*60 {
		initialVisibility = "hidden"
		minNow = minHours * 60
	}

	// Execute the template for the current time line.
	err = tmpl.ExecuteTemplate(&buf, "nowline", gin.H{
		"nRow":              nRow,
		"minNow":            minNow,
		"minHours":          minHours,
		"initialVisibility": initialVisibility,
	})
	if err != nil {
		err = fmt.Errorf("template execution failed for notioncalendar Widget: %v", err)
		goto api_drawhorizontallines_finish
	}

	// Execute the template for each horizontal line at specified intervals.
	for i := minHours; i <= maxHours; i++ {
		if i%3 == 0 || i == maxHours {
			err = tmpl.ExecuteTemplate(&buf, "horline", gin.H{
				"nRow":  nRow,
				"index": i - minHours,
				"text":  i,
			})

			if err != nil {
				err = fmt.Errorf("template execution failed for notioncalendar Widget: %v", err)
				goto api_drawhorizontallines_finish
			}
		}
	}

api_drawhorizontallines_finish:
	return buf.String(), err
}

// RegisterApiRoutes registers the API routes for the application.
func RegisterApiRoutes(r *gin.Engine) {
	// Handler for weather forecast API endpoint.
	r.GET(util.API_ROOT_PATH+"/weatherforecast", func(c *gin.Context) {
		var err error
		var data map[string]interface{}
		var result weather.RawForecastData
		var forecastData weather.ForecastData
		nHour := 5
		nDay := 5

		// Check the query parameters for weather forecast request.
		size, location_name, latitude, longitude, amedas_code, err := weather.WeatherForecastCheckQuery(c)
		if err != nil {
			goto api_weatherforecast_err
		}

		// Fetch the weather forecast data.
		result, err = weather.FetchForecastData(latitude, longitude, nDay)
		if err != nil {
			goto api_weatherforecast_err
		}

		// Parse the fetched weather forecast data.
		forecastData, err = weather.ParseForecastData(result, nHour, nDay)
		if err != nil {
			goto api_weatherforecast_err
		}

		data = map[string]interface{}{
			"location_name": location_name,
		}

		// If the layout size is MiddleV, generate the horizontal lines and historical data.
		if size == layout.MiddleV {
			lines := ""
			now := time.Now()
			data["today"] = now.Format("January 2")
			lines, err = DrawHorizontalLines(0, 24, now)
			if err != nil {
				goto api_weatherforecast_err
			}
			data["lines"] = lines

			graphData := map[string]string{}
			allHistData := []weather.HistoricalData{}
			for i := 0; i <= now.Hour()/3; i++ {
				rawHistData, err := weather.FetchHistWeatherData(amedas_code, now, i)
				if err != nil {
					goto api_weatherforecast_err
				}
				histData, err := weather.ParseHistWeatherData(rawHistData)
				if err != nil {
					goto api_weatherforecast_err
				}
				allHistData = append(allHistData, histData...)
			}
			// Sort the historical data by timestamp.
			slices.SortStableFunc(allHistData, func(a weather.HistoricalData, b weather.HistoricalData) int {
				return cmp.Compare(a.Timestamp, b.Timestamp)
			})
			for _, histData := range allHistData {
				graphData["temp"] += fmt.Sprintf("%f, ", histData.Temp)
			}
		}

		data = util.MergeMaps(data, util.StructToMap(forecastData))

		c.JSON(http.StatusOK, data)
		return

	api_weatherforecast_err:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	})

	// Handler for Notion calendar API endpoint.
	r.GET(util.API_ROOT_PATH+"/notioncalendar", func(c *gin.Context) {
		var err error
		var size layout.WidgetSize
		var queryResponse notion.RawQueryResponse
		var calendarData notion.CalendarData
		var buf bytes.Buffer
		var tmpl *template.Template
		var retData map[string]interface{}

		now := time.Now()

		// Check the query parameters for Notion calendar request.
		size, err = notion.NotionCalendarCheckQuery(c)
		if err != nil {
			goto api_notioncalendar_err
		}

		if size == layout.MiddleV || size == layout.LongV {
			retData = map[string]interface{}{
				"today":  now.Format("January 2"),
				"lines":  "",
				"events": "",
			}

			// Fetch the Notion calendar data.
			queryResponse, err = notion.FetchCalendarData(now)
			if err != nil {
				goto api_notioncalendar_err
			}

			// Parse the fetched calendar data.
			calendarData, err = notion.ParseCalendarData(queryResponse, size == layout.LongV)
			if err != nil {
				goto api_notioncalendar_err
			}

			if len(calendarData.Events) == 0 {
				goto api_notioncalendar_success
			}

			// Generate the horizontal lines for the calendar events.
			retData["lines"], err = DrawHorizontalLines(calendarData.MinHours, calendarData.MaxHours, now)
			if err != nil {
				goto api_notioncalendar_err
			}

			// Parse the template for the calendar events.
			tmpl, err = template.New("eventTmpl").ParseFiles("templates/widgets/notioncalendar.tmpl")
			if err != nil {
				err = fmt.Errorf("failed to find a template for notioncalendar Widget: %v", err)
				goto api_notioncalendar_err
			}
			// Execute the template for each calendar event.

			buf.WriteString(fmt.Sprintf("<div style=\"--nslot: %d; --min-hours: %d; --max-hours: %d;\">\n", calendarData.NSlot, calendarData.MinHours, calendarData.MaxHours))
			for _, event := range calendarData.Events {
				if !event.IsAllDay {
					err = tmpl.ExecuteTemplate(&buf, "event", util.StructToMap(event))
					if err != nil {
						err = fmt.Errorf("template execution failed for notioncalendar Widget: %v", err)
						goto api_notioncalendar_err
					}
				}
			}
			buf.WriteString("</div>")
			retData["events"] = buf.String()
		} else {
			tomorrow := now.AddDate(0, 0, 1)
			dat := now.AddDate(0, 0, 2)

			retData = map[string]interface{}{
				"tomorrow":        tomorrow.Format("January 2"),
				"dat":             dat.Format("January 2"),
				"tomorrow_events": "",
				"dat_events":      "",
			}

			var buf2 bytes.Buffer
			for keyName, date := range map[string]time.Time{"tomorrow_events": tomorrow, "dat_events": dat} {
				queryResponse, err = notion.FetchCalendarData(date)
				if err != nil {
					goto api_notioncalendar_err
				}

				// Parse the fetched calendar data.
				calendarData, err = notion.ParseCalendarData(queryResponse, false)
				if err != nil {
					goto api_notioncalendar_err
				}

				if len(calendarData.Events) == 0 {
					buf.WriteString("<span class=\"w-full\" style=\"font-size: 10%;\" >No events</span>")
				}

				for _, event := range calendarData.Events {
					if event.IsAllDay {
						buf.WriteString(fmt.Sprintf(`<span class=\"w-full\" style=\"font-size: 10%%; background-color: var(--md-sys-color-primary-container);\" >%s</span>`, event.Name))
					} else {
						buf2.WriteString(fmt.Sprintf("<div class=\"wg-hstack\" style=\"font-size: 10%%;\"><span class=\"material-symbols-outlined\" style=\"color: var(--md-sys-color-primary-container);\">circle</span><span>%s %s</span></div>", event.TimeDesc, event.Name))
					}
				}
				retData[keyName] = buf.String() + buf2.String()
				buf.Reset()
				buf2.Reset()
			}
		}

	api_notioncalendar_success:
		c.JSON(http.StatusOK, retData)
		return

	api_notioncalendar_err:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	})
}
