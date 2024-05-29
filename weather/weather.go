// Package weather provides functionalities for fetching and parsing weather data.
package weather

import (
	"encoding/json"
	"fmt"
	"go/screensaver/layout"
	"go/screensaver/util"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// WeatherCode represents various weather conditions as per the open-meteo API.
type WeatherCode int64

// Constants for different weather conditions from the open-meteo API.
const (
	ClearSky WeatherCode = 0

	MainlyClear  WeatherCode = 1
	PartlyCloudy WeatherCode = 2
	Overcast     WeatherCode = 3

	Fog               WeatherCode = 45
	DepositingRimeFog WeatherCode = 48

	DrizzleLight    WeatherCode = 51
	DrizzleModerate WeatherCode = 53
	DrizzleDense    WeatherCode = 55

	FreezingDrizzleLight WeatherCode = 56
	FreezingDrizzleDense WeatherCode = 57

	RainSlight   WeatherCode = 61
	RainModerate WeatherCode = 63
	RainHeavy    WeatherCode = 65

	FreezingRainLight WeatherCode = 66
	FreezingRainHeavy WeatherCode = 67

	SnowFallSlight   WeatherCode = 71
	SnowFallModerate WeatherCode = 73
	SnowFallHeavy    WeatherCode = 75

	SnowGrains WeatherCode = 77

	RainShowersSlight   WeatherCode = 80
	RainShowersModerate WeatherCode = 81
	RainShowersViolent  WeatherCode = 82

	SnowShowersSlight WeatherCode = 85
	SnowShowersHeavy  WeatherCode = 86

	ThunderstormSlightOrModerate WeatherCode = 95
	ThunderstormWithSlightHail   WeatherCode = 96
	ThunderstormWithHeavyHail    WeatherCode = 99
)

// Mapping of weather codes to icon names.
var weatherIcons = map[WeatherCode]string{
	ClearSky:                     "clear_day",
	MainlyClear:                  "clear_day",
	PartlyCloudy:                 "partly_cloudy_day",
	Overcast:                     "cloud",
	Fog:                          "foggy",
	DepositingRimeFog:            "foggy",
	DrizzleLight:                 "mist",
	DrizzleModerate:              "mist",
	DrizzleDense:                 "mist",
	FreezingDrizzleLight:         "weather_mix",
	FreezingDrizzleDense:         "weather_mix",
	RainSlight:                   "rainy_light",
	RainModerate:                 "rainy_light",
	RainHeavy:                    "rainy_heavy",
	FreezingRainLight:            "rainy_snow",
	FreezingRainHeavy:            "rainy_snow",
	SnowFallSlight:               "weather_snowy",
	SnowFallModerate:             "weather_snowy",
	SnowFallHeavy:                "snowing_heavy",
	SnowGrains:                   "snowing_heavy",
	RainShowersSlight:            "rainy_light",
	RainShowersModerate:          "rainy_light",
	RainShowersViolent:           "rainy_heavy",
	SnowShowersSlight:            "snowing",
	SnowShowersHeavy:             "snowing_heavy",
	ThunderstormSlightOrModerate: "thunderstorm",
	ThunderstormWithSlightHail:   "thunderstorm",
	ThunderstormWithHeavyHail:    "thunderstorm",
}

// Mapping of weather codes to Japanese descriptions.
var weatherDescriptionsJP = map[WeatherCode]string{
	ClearSky:                     "晴天",
	MainlyClear:                  "主に晴れ",
	PartlyCloudy:                 "部分的に曇り",
	Overcast:                     "曇り",
	Fog:                          "霧",
	DepositingRimeFog:            "霧氷",
	DrizzleLight:                 "小雨",
	DrizzleModerate:              "中程度の雨",
	DrizzleDense:                 "濃雨",
	FreezingDrizzleLight:         "弱い着氷性霧雨",
	FreezingDrizzleDense:         "濃い着氷性霧雨",
	RainSlight:                   "小雨",
	RainModerate:                 "中程度の雨",
	RainHeavy:                    "大雨",
	FreezingRainLight:            "弱い着氷性雨",
	FreezingRainHeavy:            "強い着氷性雨",
	SnowFallSlight:               "小雪",
	SnowFallModerate:             "中程度の雪",
	SnowFallHeavy:                "大雪",
	SnowGrains:                   "雪粒",
	RainShowersSlight:            "小雨のにわか雨",
	RainShowersModerate:          "中雨のにわか雨",
	RainShowersViolent:           "激しいにわか雨",
	SnowShowersSlight:            "小雪のにわか雪",
	SnowShowersHeavy:             "大雪のにわか雪",
	ThunderstormSlightOrModerate: "雷雨",
	ThunderstormWithSlightHail:   "弱いひょうを伴う雷雨",
	ThunderstormWithHeavyHail:    "強いひょうを伴う雷雨",
}

// Mapping of weather codes to English descriptions.
var weatherDescriptionsEN = map[WeatherCode]string{
	ClearSky:                     "Clear Sky",
	MainlyClear:                  "Mainly Clear",
	PartlyCloudy:                 "Partly Cloudy",
	Overcast:                     "Overcast",
	Fog:                          "Fog",
	DepositingRimeFog:            "Depositing Rime Fog",
	DrizzleLight:                 "Light Drizzle",
	DrizzleModerate:              "Moderate Drizzle",
	DrizzleDense:                 "Dense Drizzle",
	FreezingDrizzleLight:         "Light Freezing Drizzle",
	FreezingDrizzleDense:         "Dense Freezing Drizzle",
	RainSlight:                   "Light Rain",
	RainModerate:                 "Moderate Rain",
	RainHeavy:                    "Heavy Rain",
	FreezingRainLight:            "Light Freezing Rain",
	FreezingRainHeavy:            "Heavy Freezing Rain",
	SnowFallSlight:               "Light Snowfall",
	SnowFallModerate:             "Moderate Snowfall",
	SnowFallHeavy:                "Heavy Snowfall",
	SnowGrains:                   "Snow Grains",
	RainShowersSlight:            "Light Rain Showers",
	RainShowersModerate:          "Moderate Rain Showers",
	RainShowersViolent:           "Violent Rain Showers",
	SnowShowersSlight:            "Light Snow Showers",
	SnowShowersHeavy:             "Heavy Snow Showers",
	ThunderstormSlightOrModerate: "Slight or Moderate Thunderstorm",
	ThunderstormWithSlightHail:   "Thunderstorm with Light Hail",
	ThunderstormWithHeavyHail:    "Thunderstorm with Heavy Hail",
}

// Default language for weather descriptions.
const DEFAULT_LANG = "EN"

// GetWeatherDescriptions returns the weather description based on the provided code and language.
func GetWeatherDescriptions(code WeatherCode, lang string) string {
	switch lang {
	case "EN":
		return weatherDescriptionsEN[code]
	case "JP":
		return weatherDescriptionsJP[code]
	}
	return weatherDescriptionsEN[code]
}

// WeatherForecastCheckQuery checks and validates query parameters for weather forecast requests.
func WeatherForecastCheckQuery(c *gin.Context) (layout.WidgetSize, string, float64, float64, string, error) {
	var location_name, amedas_code string
	var latitude, longitude float64
	var err error

	size_str := c.Query("size")
	location_name = c.Query("location_name")
	latitude_str := c.Query("location_latitude")
	longitude_str := c.Query("location_longitude")
	amedas_code = c.Query("location_histdata")

	if size_str == "" {
		err = fmt.Errorf("please provide size information")
		goto weatherforecast_checkquery_finish

	} else if !layout.SizeCheck(layout.WeatherForecastWidget, (layout.WidgetSize)(size_str)) {
		err = fmt.Errorf("invalid size")
		goto weatherforecast_checkquery_finish
	}

	if location_name == "" || latitude_str == "" || longitude_str == "" {
		err = fmt.Errorf("please provide location information")
		goto weatherforecast_checkquery_finish
	} else if (layout.WidgetSize)(size_str) == layout.MiddleV && amedas_code == "" {
		err = fmt.Errorf("please provide the amedas location code for historical data (Specific parameter for MiddleV)")
		goto weatherforecast_checkquery_finish
	}
	latitude, err = strconv.ParseFloat(latitude_str, 64)
	if err != nil {
		err = fmt.Errorf("please provide valid latitude information")
		goto weatherforecast_checkquery_finish
	}
	longitude, err = strconv.ParseFloat(longitude_str, 64)
	if err != nil {
		err = fmt.Errorf("please provide valid longitude information")
		goto weatherforecast_checkquery_finish
	}

weatherforecast_checkquery_finish:
	return (layout.WidgetSize)(size_str), location_name, latitude, longitude, amedas_code, err
}

// RawCurrentData represents the current weather data.
type RawCurrentData struct {
	Temperature2M float64 `json:"temperature_2m"`
	WeatherCode   int     `json:"weather_code"`
}

// RawHourlyData represents the hourly weather data.
type RawHourlyData struct {
	Time          []string  `json:"time"`
	Temperature2M []float64 `json:"temperature_2m"`
	WeatherCode   []int     `json:"weather_code"`
}

// RawDailyData represents the daily weather data.
type RawDailyData struct {
	Time             []string  `json:"time"`
	WeatherCode      []int     `json:"weather_code"`
	Temperature2MMax []float64 `json:"temperature_2m_max"`
	Temperature2MMin []float64 `json:"temperature_2m_min"`
}

// RawForecastData represents the complete forecast data.
type RawForecastData struct {
	Timezone string         `json:"timezone"`
	Current  RawCurrentData `json:"current"`
	Hourly   RawHourlyData  `json:"hourly"`
	Daily    RawDailyData   `json:"daily"`
}

// HourIndex represents an hourly index for weather data.
type HourIndex struct {
	Time  int
	Index int
}

// DayIndex represents a daily index for weather data.
type DayIndex struct {
	Time  int
	Index int
}

// CurrentData represents the current weather data for display.
type CurrentData struct {
	Temp        string `json:"temp"`
	WeatherIcon string `json:"weather_icon"`
	WeatherName string `json:"weather_name"`
}

// HourlyData represents the hourly weather data for display.
type HourlyData struct {
	Time        string `json:"time"`
	Temp        string `json:"temp"`
	WeatherIcon string `json:"weather_icon"`
	WeatherName string `json:"weather_name"`
}

// DailyData represents the daily weather data for display.
type DailyData struct {
	Time        string `json:"time"`
	TempMax     string `json:"temp_max"`
	TempMin     string `json:"temp_min"`
	WeatherIcon string `json:"weather_icon"`
	WeatherName string `json:"weather_name"`
}

// ForecastData represents the complete forecast data for display.
type ForecastData struct {
	Current CurrentData  `json:"current"`
	Hourly  []HourlyData `json:"hourly"`
	Daily   []DailyData  `json:"daily"`
}

// FindNextNHours finds the next N hours of weather data from the provided time strings.
func FindNextNHours(datetimes []string, datetimesTimezone string, nHour int) ([]HourIndex, error) {
	var err error
	var now, nextHour time.Time
	var nextHourStr string
	var nextHourIndex int
	var indices []HourIndex

	location, err := time.LoadLocation(datetimesTimezone)
	if err != nil {
		err = fmt.Errorf("invalid time zone \"%s\"", datetimesTimezone)
		goto weather_findnexthours_finish
	}

	now = time.Now().In(location)
	nextHour = now.Add(time.Hour)
	nextHourStr = nextHour.Format("2006-01-02T15:00")

	nextHourIndex = -1

	for i, datetimeStr := range datetimes {
		if datetimeStr == nextHourStr {
			nextHourIndex = i
			break
		}
	}

	if nextHourIndex == -1 {
		return nil, fmt.Errorf("no forthcoming time found")
	}

	indices = make([]HourIndex, nHour)
	for i := 0; i < nHour; i++ {
		indices[i] = HourIndex{
			Time:  nextHour.Add(time.Duration(i) * time.Hour).Hour(),
			Index: nextHourIndex + i,
		}
	}

weather_findnexthours_finish:
	return indices, err
}

// FindNextNDays finds the next N days of weather data from the provided date strings.
func FindNextNDays(dates []string, datesTimezone string, nDay int) ([]DayIndex, error) {
	var err error
	var today, tomorrow time.Time
	var tomorrowStr string
	var tomorrowIndex int
	var indices []DayIndex

	location, err := time.LoadLocation(datesTimezone)
	if err != nil {
		err = fmt.Errorf("invalid time zone \"%s\"", datesTimezone)
		goto weather_findnextdays_finish
	}

	today = time.Now().In(location)
	tomorrow = today.AddDate(0, 0, 1)
	tomorrowStr = tomorrow.Format("2006-01-02")

	tomorrowIndex = -1

	for i, dateStr := range dates {
		if dateStr == tomorrowStr {
			tomorrowIndex = i
			break
		}
	}

	if tomorrowIndex == -1 {
		return nil, fmt.Errorf("no forthcoming day found")
	}

	indices = make([]DayIndex, nDay)
	for i := 0; i < nDay; i++ {
		indices[i] = DayIndex{
			Time:  tomorrow.AddDate(0, 0, i).Day(),
			Index: tomorrowIndex + i,
		}
	}

weather_findnextdays_finish:
	return indices, err
}

// FetchForecastData fetches weather forecast data from the Open Meteo API.
func FetchForecastData(latitude, longitude float64, nDay int) (RawForecastData, error) {
	var result RawForecastData
	var err error
	var resp *http.Response
	var body []byte

	url := fmt.Sprintf("https://api.open-meteo.com/v1/jma?latitude=%f&longitude=%f&current=temperature_2m,weather_code&hourly=temperature_2m,weather_code&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=Asia%%2FTokyo&forecast_days=%d",
		latitude,
		longitude,
		nDay+1,
	)

	resp, err = http.Get(url)
	if err != nil {
		err = fmt.Errorf("failed to fetch weather data (url: %s)", url)
		goto weather_fetchforecastdata_finish
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read weather data body (url: %s)", url)
		goto weather_fetchforecastdata_finish
	}

	if err = json.Unmarshal(body, &result); err != nil {
		err = fmt.Errorf("failed to unmarshal weather data json (url: %s)", url)
		goto weather_fetchforecastdata_finish
	}

weather_fetchforecastdata_finish:
	return result, err
}

// ParseForecastData parses the raw forecast data into structured forecast data for display.
func ParseForecastData(data RawForecastData, nHour, nDay int) (ForecastData, error) {
	var nextHours []HourIndex
	var nextDays []DayIndex
	var err error
	var current CurrentData
	var hourlyDataCol []HourlyData
	var dailyDataCol []DailyData

	// Extract required fields from the JSON response
	timezone := data.Timezone

	current = CurrentData{
		Temp:        fmt.Sprintf(util.TEMP_FORMAT_CUR, data.Current.Temperature2M),
		WeatherIcon: weatherIcons[WeatherCode(data.Current.WeatherCode)],
		WeatherName: GetWeatherDescriptions(WeatherCode(data.Current.WeatherCode), DEFAULT_LANG),
	}

	nextHours, err = FindNextNHours(data.Hourly.Time, timezone, nHour)
	if err != nil {
		err = fmt.Errorf("failed to find the next coming %d hours in the forecast data", nHour)
		goto weather_parseforecastdata_finish
	}

	hourlyDataCol = make([]HourlyData, nHour)
	for i, nextData := range nextHours {
		hourlyDataCol[i] = HourlyData{
			Time:        fmt.Sprintf("%d", nextData.Time),
			Temp:        fmt.Sprintf(util.TEMP_FORMAT_HOUR, data.Hourly.Temperature2M[nextData.Index]),
			WeatherIcon: weatherIcons[WeatherCode(int64(data.Hourly.WeatherCode[nextData.Index]))],
			WeatherName: GetWeatherDescriptions(WeatherCode(int64(data.Hourly.WeatherCode[nextData.Index])), DEFAULT_LANG),
		}
	}

	nextDays, err = FindNextNDays(data.Daily.Time, timezone, nDay)
	if err != nil {
		err = fmt.Errorf("failed to find the next coming %d days in the forecast data", nDay)
		goto weather_parseforecastdata_finish
	}

	dailyDataCol = make([]DailyData, nDay)
	for i, nextData := range nextDays {
		dailyDataCol[i] = DailyData{
			Time:        fmt.Sprintf("%d", nextData.Time),
			TempMax:     fmt.Sprintf(util.TEMP_FORMAT_DAY, data.Daily.Temperature2MMax[nextData.Index]),
			TempMin:     fmt.Sprintf(util.TEMP_FORMAT_DAY, data.Daily.Temperature2MMin[nextData.Index]),
			WeatherIcon: weatherIcons[WeatherCode(int64(data.Daily.WeatherCode[nextData.Index]))],
			WeatherName: GetWeatherDescriptions(WeatherCode(int64(data.Daily.WeatherCode[nextData.Index])), DEFAULT_LANG),
		}
	}

weather_parseforecastdata_finish:
	return ForecastData{current, hourlyDataCol, dailyDataCol}, err
}

// RawHistoricalData represents the historical weather data.
type RawHistoricalData struct {
	Temp             [2]float64 `json:"temp"`
	Humidity         [2]float64 `json:"humidity"`
	Weather          [2]int     `json:"weather"`
	Precipitation10m [2]float64 `json:"precipitation10m"`
	Wind             [2]float64 `json:"wind"`
	WindDirection    [2]int     `json:"windDirection"`
	NormalPressure   [2]float64 `json:"normalPressure"`
}

// RawHistoricalDataMap maps timestamps to RawHistoricalData.
type RawHistoricalDataMap map[string]RawHistoricalData

// HistoricalData represents the structured historical weather data.
type HistoricalData struct {
	Timestamp        int64   `json:"timestamp"`
	Temp             float64 `json:"temp"`
	Humidity         float64 `json:"humidity"`
	Weather          int     `json:"weather"`
	Precipitation10m float64 `json:"precipitation10m"`
	Wind             float64 `json:"wind"`
	WindDirection    int     `json:"windDirection"`
	NormalPressure   float64 `json:"normalPressure"`
}

// FetchHistWeatherData fetches historical weather data from the JMA API.
func FetchHistWeatherData(amedas_code string, ofWhen time.Time, quarterIndex int) (RawHistoricalDataMap, error) {
	var weatherData RawHistoricalDataMap
	var err error
	var resp *http.Response
	var body []byte

	url := fmt.Sprintf("https://www.jma.go.jp/bosai/amedas/data/point/%s/%d%02d%02d_%02d.json",
		amedas_code,
		ofWhen.Year(),
		ofWhen.Month(),
		ofWhen.Day(),
		quarterIndex*3,
	)

	resp, err = http.Get(url)
	if err != nil {
		err = fmt.Errorf("failed to fetch historical weather data (url: %s)", url)
		goto weatherforecast_fetchhistoricaldata_finish
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read historical weather data body (url: %s)", url)
		goto weatherforecast_fetchhistoricaldata_finish
	}

	if err = json.Unmarshal(body, &weatherData); err != nil {
		err = fmt.Errorf("failed to unmarshal historical weather data json (url: %s)", url)
		goto weatherforecast_fetchhistoricaldata_finish
	}

weatherforecast_fetchhistoricaldata_finish:
	return weatherData, err
}

// ParseHistWeatherData parses raw historical weather data into structured historical data.
func ParseHistWeatherData(rawWeatherData RawHistoricalDataMap) ([]HistoricalData, error) {
	// Create a slice to hold the compacted data
	var compactedData []HistoricalData
	var err error

	// Iterate over the parsed data to extract and compact the required fields
	for timestamp, data := range rawWeatherData {
		timestamp_int, err := strconv.ParseInt(timestamp[8:12], 10, 32)
		if err != nil {
			return nil, err
		}

		compactedData = append(compactedData, HistoricalData{
			Timestamp:        timestamp_int,
			Temp:             data.Temp[0],
			Humidity:         data.Humidity[0],
			Weather:          data.Weather[0],
			Precipitation10m: data.Precipitation10m[0],
			Wind:             data.Wind[0],
			WindDirection:    data.WindDirection[0],
			NormalPressure:   data.NormalPressure[0],
		})
	}
	return compactedData, err
}
