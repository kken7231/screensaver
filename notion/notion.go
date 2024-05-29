// Package notion provides functionalities to interact with Notion API and manage calendar data.
package notion

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/kken7231/screensaver/layout"
	"github.com/kken7231/screensaver/util"

	"github.com/gin-gonic/gin"
)

// NotionCalendarCheckQuery checks the query parameters for Notion calendar and validates the size.
func NotionCalendarCheckQuery(c *gin.Context) (layout.WidgetSize, error) {
	var err error

	size_str := c.Query("size")

	if size_str == "" {
		err = fmt.Errorf("please provide size information")
		goto notioncalendar_checkquery_finish

	} else if !layout.SizeCheck(layout.NotionCalendarWidget, (layout.WidgetSize)(size_str)) {
		err = fmt.Errorf("invalid size")
		goto notioncalendar_checkquery_finish
	}
notioncalendar_checkquery_finish:
	return (layout.WidgetSize)(size_str), err
}

// RawQueryResponse represents the root structure of the response from Notion API.
type RawQueryResponse struct {
	Object  string           `json:"object"`
	Results []RawQueryResult `json:"results"`
	Message string           `json:"message,omitempty"`
}

// RawQueryResult represents a single result in the response from Notion API.
type RawQueryResult struct {
	Properties map[string]Property `json:"properties"`
}

// Property represents a property of a Notion database entry.
type Property struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Checkbox    *CheckboxObj    `json:"checkbox,omitempty"`
	Date        *DateObj        `json:"date,omitempty"`
	Email       *EmailObj       `json:"email,omitempty"`
	Formula     *FormulaObj     `json:"formula,omitempty"`
	MultiSelect []SelectObj     `json:"multi_select,omitempty"`
	Number      *NumberObj      `json:"number,omitempty"`
	PhoneNumber *PhoneNumberObj `json:"phone_number,omitempty"`
	RichText    []RichTextObj   `json:"rich_text,omitempty"`
	Select      *SelectObj      `json:"select,omitempty"`
	Status      *StatusObj      `json:"status,omitempty"`
	Title       []RichTextObj   `json:"title,omitempty"`
	URL         *URLObj         `json:"url,omitempty"`
}

// CheckboxObj represents a checkbox property in Notion.
type CheckboxObj bool

// DateObj represents a date property in Notion.
type DateObj struct {
	Start    string      `json:"start"`
	End      string      `json:"end"`
	TimeZone interface{} `json:"time_zone"`
}

// EmailObj represents an email property in Notion.
type EmailObj string

// FormulaObj represents a formula property in Notion.
type FormulaObj struct {
	Type    string     `json:"type"`
	Boolean *bool      `json:"boolean,omitempty"`
	Date    *time.Time `json:"date,omitempty"`
	Number  *int       `json:"number,omitempty"`
	String  *string    `json:"string,omitempty"`
}

// NumberObj represents a number property in Notion.
type NumberObj float64

// PhoneNumberObj represents a phone number property in Notion.
type PhoneNumberObj string

// RichTextObj represents a rich text property in Notion.
type RichTextObj struct {
	Type        string                 `json:"type"`
	Equation    *RichTextEquationObj   `json:"equation,omitempty"`
	Text        *RichTextTextObj       `json:"text,omitempty"`
	Annotations RichTextAnnotationsObj `json:"annotations"`
	PlainText   string                 `json:"plain_text"`
	Href        interface{}            `json:"href"`
}

// RichTextEquationObj represents an equation in a rich text property in Notion.
type RichTextEquationObj struct {
	Expression string `json:"expression"`
}

// RichTextTextObj represents the text content in a rich text property in Notion.
type RichTextTextObj struct {
	Content string      `json:"content"`
	Link    interface{} `json:"link"`
}

// RichTextAnnotationsObj represents the text annotations in a rich text property in Notion.
type RichTextAnnotationsObj struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

// SelectObj represents a select property in Notion.
type SelectObj struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// StatusObj represents a status property in Notion.
type StatusObj struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// URLObj represents a URL property in Notion.
type URLObj string

// RawEvent represents an event with raw data.
type RawEvent struct {
	Name     string
	Start    time.Time
	End      time.Time
	IsAllDay bool
	Level    int
}

// Event represents a structured event.
type Event struct {
	Name      string `json:"name"`
	TimeDesc  string `json:"time_desc"`
	StartMins int    `json:"start_mins"`
	EndMins   int    `json:"end_mins"`
	Level     int    `json:"level"`
	Color     string `json:"color"`
	IsAllDay  bool   `json:"is_all_day"`
}

// CalendarData represents the structured calendar data.
type CalendarData struct {
	MinHours int
	MaxHours int
	NSlot    int
	Events   []Event
}

// HierarchizeEvents organizes events into hierarchical levels based on their start and end times.
func HierarchizeEvents(events *[]RawEvent) {
	if len(*events) == 0 {
		return
	}

	eventsCopy := make([]RawEvent, len(*events))
	copy(eventsCopy, *events)

	// Sort events by start time
	slices.SortStableFunc(eventsCopy, func(a RawEvent, b RawEvent) int {
		first := cmp.Compare(a.Start.Unix(), b.Start.Unix())
		if first != 0 {
			return first
		}
		return cmp.Compare(b.End.Unix(), a.End.Unix())
	})

	curLevel := 0
	var compareTo RawEvent
	var eventsProcessed []RawEvent
	for len(eventsCopy) > 0 {
		eventsCopy[0].Level = curLevel
		eventsProcessed = append(eventsProcessed, eventsCopy[0])
		compareTo = eventsCopy[0]
		eventsCopy = eventsCopy[1:]
		for i, event := range eventsCopy {
			if event.Start.Unix() >= compareTo.End.Unix() {
				event.Level = curLevel
				eventsProcessed = append(eventsProcessed, event)
				compareTo = event
				eventsCopy = append(eventsCopy[:i], eventsCopy[i+1:]...)
			}
		}
		curLevel = curLevel + 1
	}
	*events = eventsProcessed
}

// FetchCalendarData fetches calendar data from the Notion API.
func FetchCalendarData(ofWhen time.Time) (RawQueryResponse, error) {
	var result RawQueryResponse
	var err error
	var resp *http.Response
	var body []byte
	var client *http.Client

	// Calculate yesterday and tomorrow
	tomorrow := ofWhen.AddDate(0, 0, 1)

	// Normalize yesterday(tomorrow) to the start of the day
	startOfToday := time.Date(ofWhen.Year(), ofWhen.Month(), ofWhen.Day(),
		0, 0, 0, 0, ofWhen.Location())
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		0, 0, 0, 0, tomorrow.Location())

	// Format the time to ISO 8601
	todayIso := startOfToday.Format(time.RFC3339) // RFC3339 is a profile of ISO 8601
	tomorrowIso := startOfTomorrow.Format(time.RFC3339)

	endsOnOrAfterTodayFilter := fmt.Sprintf(`{
			"property": "%s",
			"date": {
				"on_or_after": "%s"
			}
		}`, util.DATE_PROPERTYNAME, todayIso)

	startsBeforeTomorrowFilter := fmt.Sprintf(`{
			"property": "%s",
			"date": {
				"before": "%s"
			}
		}`, util.DATE_PROPERTYNAME, tomorrowIso)

	categoryFilter := `{
			"property": "カテゴリ",
			"select": {
				"equals": "EV"
			}
		}`

	// Compile a regex to match tabs and new lines
	regex := regexp.MustCompile(`[\t\n]+`) // Matches one or more tabs or newlines
	data := regex.ReplaceAllString(fmt.Sprintf(`{
			"filter": {
				"and": [
					%s,
					%s,
					%s
				]
			}
		}`, endsOnOrAfterTodayFilter, startsBeforeTomorrowFilter, categoryFilter), "")

	// Create a new request
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%v/query", util.DATABASE_ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		err = fmt.Errorf("failed to create a request for notion calendar data: %v", err)
		goto notion_fetchcalendardata_finish
	}

	// Set Headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", util.API_KEY))
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to fetch notion calendar data: %v", err)
		goto notion_fetchcalendardata_finish
	}
	defer resp.Body.Close()

	// Read the response body
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read notion calendar data body (url: %s)", url)
		goto notion_fetchcalendardata_finish
	}

	// Unmarshal the response body into the result structure
	if err = json.Unmarshal(body, &result); err != nil {
		err = fmt.Errorf("failed to unmarshal notion calendar data json (url: %s)", url)
		goto notion_fetchcalendardata_finish
	}

notion_fetchcalendardata_finish:
	return result, err
}

// ParseCalendarData parses the raw query response from Notion API into structured calendar data.
func ParseCalendarData(queryResponse RawQueryResponse, forceAllDay bool) (CalendarData, error) {
	var err error
	var calendarData CalendarData
	var minHours, maxHours, nSlot int
	rawEvents := []RawEvent{}
	events := []Event{}

	if queryResponse.Object == "error" {
		err = fmt.Errorf("error found in the calendar json data: %s", queryResponse.Message)
		goto notion_parsecalendardata_finish
	}
	for _, res := range queryResponse.Results {
		eventNameProp, exists := res.Properties[util.NAME_PROPERTYNAME]
		if !exists {
			err = fmt.Errorf("error found no property corresponding to %s [properties/%s]: %v", util.NAME_PROPERTYNAME, util.NAME_PROPERTYNAME, res)
			goto notion_parsecalendardata_finish
		}

		if eventNameProp.Title == nil || len(eventNameProp.Title) < 1 {
			err = fmt.Errorf("error title is absent [properties/%s/title]: %v", util.NAME_PROPERTYNAME, res)
			goto notion_parsecalendardata_finish
		}

		if eventNameProp.Title[0].Text == nil {
			err = fmt.Errorf("error this is not a text object [properties/%s/title[0]]: %v", util.NAME_PROPERTYNAME, res)
			goto notion_parsecalendardata_finish
		}

		eventName := eventNameProp.Title[0].Text.Content

		// event's date -> properties[util.DATE_PROPERTYNAME]["date"]
		eventDateProp, exists := res.Properties[util.DATE_PROPERTYNAME]
		if !exists {
			err = fmt.Errorf("error found no property corresponding to %s [properties/%s]: %v", util.DATE_PROPERTYNAME, util.DATE_PROPERTYNAME, res)
			goto notion_parsecalendardata_finish
		}

		if eventDateProp.Date == nil {
			err = fmt.Errorf("error title is absent [properties/%s/date]: %v", util.DATE_PROPERTYNAME, res)
			goto notion_parsecalendardata_finish
		}

		eventDate := eventDateProp.Date

		var eventStartDate, eventEndDate time.Time
		layout := "2006-01-02"
		if strings.Contains(eventDate.Start, ":") {
			// time-specified event
			layout += "T15:04:05Z07:00" // The 'Z07:00' format parses timezone offsets
		}
		eventStartDate, err = time.Parse(layout, eventDate.Start)
		if err != nil {
			err = fmt.Errorf("error in converting string date to time [properties/%s/date/start]: %s", util.DATE_PROPERTYNAME, eventDate.Start)
			goto notion_parsecalendardata_finish
		}
		eventEndDate = eventStartDate
		if eventDate.End != "" {
			// end specified
			eventEndDate, err = time.Parse(layout, eventDate.End)
			if err != nil {
				err = fmt.Errorf("error in converting string date to time [properties/%s/date/end]: %s", util.DATE_PROPERTYNAME, eventDate.End)
				goto notion_parsecalendardata_finish
			}
		}
		if !strings.Contains(eventDate.Start, ":") {
			// all-day event
			eventEndDate = eventEndDate.AddDate(0, 0, 1).Add(time.Nanosecond * -1)
		}
		rawEvents = append(rawEvents, RawEvent{
			Name:     eventName,
			Start:    eventStartDate,
			End:      eventEndDate,
			IsAllDay: !strings.Contains(eventDate.Start, ":"),
		})
	}

	HierarchizeEvents(&rawEvents)

	if forceAllDay {
		maxHours = 24
		minHours = 0
	} else {
		minHours = 24
		maxHours = 0
		for _, event := range rawEvents {
			if !event.IsAllDay {
				minHours = min(event.Start.Hour(), minHours)
				maxHours = max(event.End.Add(time.Hour).Hour(), maxHours)
			}
		}
		if maxHours <= minHours {
			fmt.Printf("maxHours is less than or equal to minHours. maybe no event is registered: %v, (%d, %d)\n", rawEvents, minHours, maxHours)
			calendarData.Events = []Event{}
			goto notion_parsecalendardata_finish
		}
	}
	nSlot = maxHours - minHours + 1

	for _, event := range rawEvents {
		events = append(events, Event{
			Name:      event.Name,
			TimeDesc:  fmt.Sprintf("%d:%02d-%d:%02d", event.Start.Hour(), event.Start.Minute(), event.End.Hour(), event.End.Minute()),
			StartMins: event.Start.Hour()*60 + event.Start.Minute(),
			EndMins:   event.End.Hour()*60 + event.End.Minute(),
			Level:     event.Level,
			IsAllDay:  event.IsAllDay,
			Color:     "red",
		})
	}
	calendarData = CalendarData{
		MaxHours: maxHours,
		MinHours: minHours,
		NSlot:    nSlot,
		Events:   events,
	}

notion_parsecalendardata_finish:
	return calendarData, err
}
