# Technoline Mobile Alerts website scraper

Technoline has such a bad and poorly supported [API](http://www.mobile-alerts.eu/images/public_server_api_documentation.pdf) that I had to implement a scraper to get to my data.

`mobile-alerts-scraper.go` queries the [website](https://measurements.mobile-alerts.eu/) and requires two parameters:
- `--phoneid` your phoneId from the app
- `--location` (optional) if you want to manage multiple locations
- `--debug` (optional) for the ability to run a full trace of the app, for diagnostics and when trying to add more sensors

## Tests

Obviously I do not own every sensor so I have developed base on the ones I own and the sample ones taht can be registered from the website. If you own sensors that are not supported drop me a mail or create an issue with the data from a full-trace run.

Supported sensors: `02, 10, 08, 03, 09, 0B, 07`

## Build

### Locally

1. `go get github.com/PuerkitoBio/goquery github.com/shopspring/decimal`
1. `go run mobile-alerts-scraper.go --phoneid <your-phone-id-goes-here>`

### From Docker

```
docker build -t mobile-alerts-scraper .
docker run --rm mobile-alerts-scraper /go/bin/mobile-alerts-scraper --phoneid <your-phone-id-goes-here>
```
### For raspberry pi

```
docker build -t mobile-alerts-scraper -f $(pwd)/Dockerfile.raspi .
docker run --rm mobile-alerts-scraper /go/bin/mobile-alerts-scraper --phoneid <your-phone-id-goes-here>
```
## Run

### On amd64
```
docker run --rm docker.io/asksven/mobile-alerts-scraper:latest /go/bin/mobile-alerts-scraper --phoneid <your-phone-id-goes-here>
```

### On Raspberry Pi
```
docker run --rm docker.io/asksven/mobile-alerts-scraper:latest-rpi /go/bin/mobile-alerts-scraper --phoneid <your-phone-id-goes-here>
```

## Collect data

### logparser.py

`logparser.py` parses the scraped data, and formats it so that it can be pushed to influxdb:

1. Into two distinct timeseries for temperature and humitity
1. Aside from the value the timeserie has the following tags (for querying): `sensor_id`, `location` (so that you can retrieve data from multiple locations if you want), `reading_type` (as different sensors delivery different types, e.g. Inside out Outside), `sensor_name`

`logparser.py` pushes the data to an influxdb database and requires different parameters for that. These are defined in `setenv_template`. To operate the python programm:

1. rename `setenv_template` to `sentenv`
1. Instanciate the values based on your settings

### Scheduled job

To run the sequence (retrieve data, push to influxdb) as a cronjob:

```
*/30 * * * * sudo docker run --rm asksven/mobile-alerts-scraper:raspi-latest /go/bin/mobile-alerts-scraper --phoneid <your-phoneid> >> /home/pi/git/mobile-alerts-scraper/logs/mobile-alerts_`date "+\%Y-\%m-\%d_\%H\%M"`.log \ 
               && cd /home/pi/git/mobile-alerts-scraper && source setenv && ./loop.sh
```

1. `docker run` collects the data and puts it to `/home/pi/git/mobile-alerts-scraper/logs` as a logfile with the timestamp as name
1. `loop.sh` loops over the unprocessed logfiles and pushes the data to influxdb 
## Implementation

The scraper uses `github.com/PuerkitoBio/goquery` to process the DOM:
1. find and process each `<div class="sensor">`
1. find and process each `<div class="sensor-header">` and read the sensor name from the `<a>`
1.find and process each `<div class="sensor-component">` and extract the key from `<h4>` and value from `<h5>`

The output is a json represtation of this structure:
```go
type Reading struct {
	SensorName           string          `json:"sensor_name"`
	SensorId             string          `json:"sensor_id"`
	SensorLocation       string          `json:"sensor_location"`
	ReadingType          string          `json:"reading_type"`
	ReadingValue         decimal.Decimal `json:"reading_value"`     // can be 0 if the reading is not a number. In this case we use reading_value_str
	ReadingValue_str     string          `json:"reading_value_str"` // we try to avoid using this as long as the readings are decimal values
	ReadingUnit          string          `json:"reading_unit"`
	ReadingTimestamp_str string          `json:"reading_timestamp"`
	ReadingTimestamp_ns  int64           `json:"reading_timestamp_str"`
}
```


