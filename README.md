# go-enverbridge

# !!! WORK in progress !!!

This will be the new version of my get_solar.pl script now fully written in GOLANG. No external dependencies like perl modules, wget or curl required. There will be binaries for Windows, Linux and MacOS.

## What is already working?

  - Get StationID from Envertecportal if no ID is provided in the config json
  - collect the data from the portal for the system and the inverters
  - send the data to influxdb and mqtt

## to do's

  - CCU2 implementation
  - Debug messages to be removed
  - implement functions for HTTP calls
  - general optimizations 

# Description

get_solar is a script to collect stats from the Envertec portal and will push these metrics into a InfluxDB. Additionally the script can send the data to a MQTT broker and to a CCU2.

#### Create a configuration file

  - id: keep it empty (the script will collect the stationID from the envertec portal automatically)
  - dbcon: IP of your InfluxDB host
  - database: InfluxDB database name
  - mqttswitch: switch if mqtt should be used (y for yes and n for no)
  - mqttbroker: MQTT broker IP
  - mqttport: MQTT broker port
  - mqttuser: MQTT user
  - mqttpassword: MQTT password
  - ccu2_switch: switch if data should be send to a ccu2
  - ccu2: ccu2 IP
  - influxtag: Influx tag for the metrics
  - username: The email address you are using to login to the Envertec portal
  - password: The password which you use on the Envertec portal

#### Configuration file example

```sh
vi envertech_config.json
```
```
{
    "id" : "",
    "dbcon" : "INFLUXDB:8086",
    "database" : "enverbridge",
    "influxtag" : "enverbridge",
    "mqttswitch" : "n",
    "mqttbroker" : "MQTT-BROKER-IP",
    "mqttuser" : "MQTT-USERNAME",
    "mqttpassword" : "MQTT-PASSWORD",
    "mqttport" : "1883",
    "ccu2_switch" : "n",
    "ccu2" : "CCU2-IP",
    "username" : "EMAIL",
    "password" : "PASSWORD"
}
```

```
go run get_solar.go -config /opt/go-enverbrige/envertech_config.json
```

License
----

MIT

