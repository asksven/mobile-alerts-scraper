import json
from datetime import datetime
import sys
import os
from influxdb import InfluxDBClient
import pytz

arguments = len(sys.argv) - 1
position = 1

while (arguments >= position):
    print ("Parameter %i: %s" % (position, sys.argv[position]))
    position = position + 1

if arguments==0:
    print(sys.argv[0], " requires 1 argument")
    exit(0)

filename=sys.argv[1]

MANDATORY_ENV_VARS = ["USER", "PASSWORD", "DBNAME", "HOST", "PORT"]

try:
    for var in MANDATORY_ENV_VARS:
        if var not in os.environ:
            raise EnvironmentError("Failed because {} is not set.".format(var))
except OSError as err:
    print("Aborting: ", format(err))
    exit(1) 
# [
#   {
#     "sensor_name":"Outside Berlin",
#     "sensor_id":"026DFCDEA8B0",
#     "sensor_location":"Berlin",
#     "reading_type":"Temperature",
#     "reading_value":"1.2",
#     "reading_value_str":"1.2 C",
#     "reading_unit":"C",
#     "reading_timestamp_str":"Mon, 07 Jan 2019 02:26:42",
#     "reading_timestamp_s":1546828002
#   },
#   {
#     "sensor_name":"Wohnzimmer Berlin",
#     "sensor_id":"0727C4ABE591",
#     "sensor_location":"Berlin",
#     "reading_type":"Temperature Inside",
#     "reading_value":"21.5",
#     "reading_value_str":"21.5 C",
#     "reading_unit":"C",
#     "reading_timestamp_str":"Mon, 07 Jan 2019 02:25:15",
#     "reading_timestamp_s":1546827915},
#     ...
# Other reading_types: Humidity Inside (reading_unit=%), Temperature Outside (reading_unit=C),
#     Humidity Outside (reading_unit=%)
# !! Note: times are in local time so we need to convert
try:
    client = InfluxDBClient(
        os.getenv('HOST', ''),
        os.getenv('PORT', '8086'),
        os.getenv('USER', ''),
        os.getenv('PASSWORD', ''),
        os.getenv('DBNAME', ''),
        ssl=True, verify_ssl=True)
except:
    print("An unexpected error occured when connecting to the database. Aborting")
    sys.exc_info()
    exit(1)

try:
  f = open(filename, "r")
except:
    print("File ", filename, " could not be opened")
    exit(1)

for x in f:
    #print(x)
    json_array = json.loads(x)

    for item in json_array:
        print('Time: ', item['reading_timestamp_str'])
        #reading_time = datetime.fromtimestamp(item['reading_timestamp_s'], pytz.timezone('UTC'))
        #print("reading_time in local time: ", reading_time)
        #reading_time=reading_time.astimezone(pytz.timezone('UTC'))

        parsed_date=datetime.strptime(item['reading_timestamp_str'], "%a, %d %b %Y %H:%M:%S")
        print("Parsed: ", parsed_date)
        utc_dt = parsed_date.astimezone(pytz.utc)
        #print("Time for influxdb: ", utc_dt.strftime("%Y-%m-%dT%H:%M:%SZ"))
        reading_time=utc_dt.strftime("%Y-%m-%dT%H:%M:%SZ")
        print("reading_time in converted UTC: ", reading_time)
        print(' as JSON: ', reading_time)
        print('Reading type: ', item['reading_type'])
        print('Value: ', item['reading_value'])
        print('Unit: ', item['reading_unit'])

        if str.startswith(item['reading_type'], "Temperature"):
            json_body = [ \
                {"measurement": "temperature_sensor", \
                 "tags": { \
                     "sensor_id": item['sensor_id'], \
                     "location": item['sensor_location'], \
                     "reading_type": item['reading_type'], \
                     "sensor_name": item['sensor_name']}, \
                 "time": reading_time, \
                 "fields": { \
                     "temperature": float(item['reading_value']) \
            }}]

        if str.startswith(item['reading_type'], "Humidity"):
            json_body = [ \
                {"measurement": "humidity_sensor", \
                 "tags": { \
                     "sensor_id": item['sensor_id'], \
                     "location": item['sensor_location'], \
                     "reading_type": item['reading_type'], \
                     "sensor_name": item['sensor_name']}, \
                 "time": reading_time, \
                 "fields": { \
                     "humidity": float(item['reading_value']) \
            }}]

        try:
            client.write_points(json_body)
            result = client.query('select temperature from plant_sensor;')
            print(">>> Written timestamp ", reading_time)
        except requests.exceptions.ConnectionError as err:
            print("!!! Connection error: ", format(err))
        except:
            print("!!! An unexpected error occured when connecting to the database. Aborting")
            sys.exc_info()
