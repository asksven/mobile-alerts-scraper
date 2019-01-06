// make_http_request.go
// Based on https://www.devdungeon.com/content/web-scraping-go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

type kvPair struct {
	key   string
	value string
}

type Reading struct {
	SensorName           string          `json:"sensor_name"`
	SensorId             string          `json:"sensor_id"`
	SensorLocation       string          `json:"sensor_location"`
	ReadingType          string          `json:"reading_type"`
	ReadingValue         decimal.Decimal `json:"reading_value"`     // can be 0 if the reading is not a number. In this case we use reading_value_str
	ReadingValue_str     string          `json:"reading_value_str"` // we try to avoid using this as long as the readings are decimal values
	ReadingUnit          string          `json:"reading_unit"`
	ReadingTimestamp_str string          `json:"reading_timestamp_str"`
	ReadingTimestamp_s   int64           `json:"reading_timestamp_s"`
}

var parsedElements []kvPair
var readings []Reading
var phoneId string
var location string
var debug bool

// This will get called for each Value-Element found
func parseValue(index int, element *goquery.Selection) {
	// See if the href attribute exists on the element
	value, err := element.Html()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		fmt.Println("parseValue" + " " + value)
	}
	parsedElements = append(parsedElements, kvPair{key: "Value", value: value})

}

// This will get called for each Key-Element found
func parseKey(index int, element *goquery.Selection) {
	// See if the href attribute exists on the element
	value, err := element.Html()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		fmt.Println("parceKey" + " " + value)
	}
	parsedElements = append(parsedElements, kvPair{key: "Key", value: value})
	//log.Println(string)

}

// This will get called for each Key-Element found
func parseName(index int, element *goquery.Selection) {
	// See if the href attribute exists on the element
	value, err := element.Html()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		fmt.Println("parseName" + " " + value)
	}
	parsedElements = append(parsedElements, kvPair{key: "Name", value: value})
	//log.Println(string)

}

// This will get called for each sensor-component element found
func parseSensorComponent(index int, element *goquery.Selection) {
	// string, err := element.Html()
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Println(string)
	// }

	element.Find("h5").Each(parseKey)
	element.Find("h4").Each(parseValue)

}

// This will get called for each sensor-component element found
func parseSensorHeader(index int, element *goquery.Selection) {
	// string, err := element.Html()
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Println(string)
	// }

	element.Find("a").Each(parseName)
}

// This will get called for each sensor-header element found
func parseSensor(index int, element *goquery.Selection) {
	// string, err := element.Html()
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Println(string)
	// }
	element.Find(".sensor-header").Each(parseSensorHeader)
	element.Find(".sensor-component").Each(parseSensorComponent)

}

func parseValUnit(input string) (decimal.Decimal, string) {
	s := strings.Split(input, " ")
	valueStr, unit := s[0], s[1]
	value, err := decimal.NewFromString(valueStr)
	if err != nil {
		return decimal.Zero, ""
	} else {
		return value, unit
	}
}

func parseTimeStamp(input string) (string, int64) {
	// Writing down the way the standard time would look like formatted our way
	layout := "1/2/2006 3:4:5 PM"
	layout2 := "Mon, 02 Jan 2006 15:04:05"

	// fmt.Println("Layout:" + layout)
	// fmt.Println("Input:" + input)

	t, _ := time.Parse(layout, input)

	//fmt.Println("Output str:" + t.Format(layout2))
	//fmt.Println("Output ns: " + strconv.FormatInt(t.Unix(), 10))
	return t.Format(layout2), t.Unix()
}

// Main
func main() {

	start := time.Now()

	// process command-line arguments
	phoneIdPtr := flag.String("phoneid", "", "the phone-id from the app (mandatory)")
	locationPtr := flag.String("location", "", "the location of the sensors (optional)")
	debugPtr := flag.Bool("debug", false, "verbose output for debugging purposes (optional)")
	flag.Parse()

	if *phoneIdPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		phoneId = *phoneIdPtr
	}

	location = *locationPtr
	debug = *debugPtr

	if debug {
		fmt.Println("Arguments passed")
		fmt.Println("phoneIdPtr:", *phoneIdPtr)
		fmt.Println("locationPtr:", *locationPtr)
		fmt.Println("debugPtr:", *debugPtr)
	}

	// request the data
	response, err := http.PostForm(
		"https://measurements.mobile-alerts.eu/",
		url.Values{
			"phoneid": {phoneId},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	document.Find(".sensor").Each(parseSensor)

	if debug {
		fmt.Println("============================")
	}

	var sensorName string
	var sensorId string
	var readingTimestamp_str string
	var sensorLocation = "Berlin"

	for i := range parsedElements {
		kv := parsedElements[i]
		if debug {
			fmt.Println("Processing " + kv.key)
		}
		switch node := kv.key; node {
		case "Name":
			// its a new sensor
			sensorName = kv.value

		case "Key":
			switch value := kv.value; value {
			case "ID":
				// next element's value is the id
				sensorId = parsedElements[i+1].value
			case "Timestamp":
				// next element's value is the time
				readingTimestamp_str = parsedElements[i+1].value

			case "Temperature":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Temperature Inside":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Temperature Outside":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Temperature Probe":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Windspeed":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Gust":
				// next element's value is gust speed
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Wind Direction":
				// next element's value is the wind direction

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				//	we don't have a decimal value
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = "Direction"
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Humidity":
				// next element's value is the humidity in %
				val := strings.Trim(parsedElements[i+1].value, "%")
				unit := "%"
				hum, err := decimal.NewFromString(val)
				if err != nil {
					panic(err)
				}
				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = hum
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Humidity Inside":
				// next element's value is the humidity in %
				val := strings.Trim(parsedElements[i+1].value, "%")
				unit := "%"
				hum, err := decimal.NewFromString(val)
				if err != nil {
					panic(err)
				}
				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = hum
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Humidity Outside":
				// next element's value is the humidity in %
				val := strings.Trim(parsedElements[i+1].value, "%")
				unit := "%"
				hum, err := decimal.NewFromString(val)
				if err != nil {
					panic(err)
				}
				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = hum
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Contact Sensor":
				// next element's value is state (Open/Closed)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				//				current.ReadingType = value
				current.ReadingValue_str = parsedElements[i+1].value
				//				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			case "Rain":
				// next element's value is the temperature and unit
				val, unit := parseValUnit(parsedElements[i+1].value)

				var current Reading
				current.SensorName = sensorName
				current.SensorId = sensorId
				str, s := parseTimeStamp(readingTimestamp_str)
				current.ReadingTimestamp_str = str
				current.ReadingTimestamp_s = s
				current.ReadingType = value
				current.ReadingValue = val
				current.ReadingValue_str = parsedElements[i+1].value
				current.ReadingUnit = unit
				current.SensorLocation = sensorLocation
				readings = append(readings, current)

			}
		}

		if debug {
			fmt.Println(kv.key + ", " + kv.value)
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)

	if debug {
		fmt.Println("======================")
	}

	b, err := json.Marshal(readings)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	if debug {
		fmt.Println("Elapsed ns:", int64(elapsed/time.Nanosecond))
	}

}
