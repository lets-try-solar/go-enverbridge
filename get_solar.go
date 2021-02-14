// https://github.com/lets-try-solar/enverbridge/blob/master/get_solar.pl

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/andelf/go-curl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// global variables
var (
	id           = ""
	dbcon        = ""
	database     = ""
	influxtag    = ""
	mqttswitch   = ""
	mqttbroker   = ""
	mqttport     = ""
	mqttuser     = ""
	mqttpassword = ""
	ccu2switch   = ""
	ccu2         = ""
	username     = ""
	password     = ""
)

type Config struct {
	ID           string `json:"id"`
	DBCON        string `json:"dbcon"`
	DATABASE     string `json:"database"`
	INFLUXTAG    string `json:"influxtag"`
	MQTTSWITCH   string `json:"mqttswitch"`
	MQTTBROKER   string `json:"mqttbroker"`
	MQTTPORT     string `json:"mqttport"`
	MQTTUSER     string `json:"mqttuser"`
	MQTTPASSWORD string `json:"mqttpassword"`
	CCU2SWITCH   string `json:"ccu2_switch"`
	CCU2         string `json:"ccu2"`
	USERNAME     string `json:"username"`
	PASSWORD     string `json:"password"`
}

// MQTT
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var mqttclient mqtt.Client

func mqttconnect(mqttbroker string, mqttport string, mqttuser string, mqttpassword string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://" + mqttbroker + ":" + mqttport)
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(mqttuser)
	opts.SetPassword(mqttpassword)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	mqttclient = mqtt.NewClient(opts)
	if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func mqttpublish(mqttclient mqtt.Client, topic string, key string, value string) {

	text := fmt.Sprintf(value)
	token := mqttclient.Publish("/enverbridge/"+topic+"/"+key, 0, false, text)
	token.Wait()
	//time.Sleep(time.Second)
}

func mqttdisconnect(mqttclient mqtt.Client) {

	mqttclient.Disconnect(1000)

}

// Main
func main() {

	cflag := flag.String("config", "", "")
	flag.Parse()
	configfilepath := fmt.Sprintf("%v", *cflag)
	if configfilepath == "" {
		fmt.Println("Please provide -config parameter which contains the path and filename of the config file. Example: /opt/enverbridge/envertech_config.json")
		os.Exit(1)
	}
	fmt.Println(configfilepath)
	// Open our configFile
	configFile, err := os.Open(configfilepath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened envertech_config.json")
	fmt.Println(" ")

	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	var config Config

	json.Unmarshal(byteValue, &config)

	id = config.ID
	dbcon = config.DBCON
	database = config.DATABASE
	influxtag = config.INFLUXTAG
	mqttswitch = config.MQTTSWITCH
	mqttbroker = config.MQTTBROKER
	mqttport = config.MQTTPORT
	mqttuser = config.MQTTUSER
	mqttpassword = config.MQTTPASSWORD
	ccu2switch = config.CCU2SWITCH
	ccu2 = config.CCU2
	username = config.USERNAME
	password = config.PASSWORD

	if mqttswitch == "y" {
		fmt.Println("Connect to MQTT Broker tcp://" + mqttbroker + ":" + mqttport)
		mqttconnect(mqttbroker, mqttport, mqttuser, mqttpassword)

	}

	formData := url.Values{
		"userName": {username},
		"pwd":      {password},
	}

	options := cookiejar.Options{
		//PublicSuffixList: publicsuffix.List,
	}
	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{Jar: jar}
	req, err := http.NewRequest("POST", "https://www.envertecportal.com/apiaccount/login", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if id == "" {
		fmt.Println("Config ID empty, get Station ID from portal")
		fmt.Println(" ")
		resp, err = client.Get("https://www.envertecportal.com/terminal/systemoverview")
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		//log.Println(string(data))
		var file = string(data)
		temp := strings.Split(file, "\n")

		for _, item := range temp {
			if strings.Contains(item, "var stationId =") == true {
				in := strings.ReplaceAll(item, " ", "")
				out := strings.Split(in, "'")
				id = out[1]
			}
		}
		fmt.Println("StatioID: " + id)
		defer resp.Body.Close()
	}

	var getStationInfo = "https://www.envertecportal.com/ApiStations/getStationInfo?stationId=" + id
	//fmt.Println(getStationInfo)
	requestbody, err := json.Marshal(map[string]string{})
	resp, err = client.Post(getStationInfo, "application/x-www-form-urlencoded", bytes.NewBuffer(requestbody))
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	//log.Println(string(data))
	var result map[string]interface{}
	json.Unmarshal([]byte(data), &result)
	temp := result["Data"].(map[string]interface{})
	var StationInfo map[string]string
	StationInfo = make(map[string]string)

	StationInfo["monthpower"] = strings.Split(fmt.Sprint(temp["UnitEMonth"]), " ")[0]
	StationInfo["capacity"] = fmt.Sprint(temp["Capacity"])
	StationInfo["etoday"] = fmt.Sprint(temp["Etoday"])
	StationInfo["daypower"] = strings.Split(fmt.Sprint(temp["UnitEToday"]), " ")[0]
	StationInfo["allpower"] = strings.Split(fmt.Sprint(temp["UnitETotal"]), " ")[0]
	StationInfo["power"] = fmt.Sprint(temp["Power"])
	StationInfo["powerstr"] = strings.Split(fmt.Sprint(temp["PowerStr"]), " ")[0]
	StationInfo["yearpower"] = strings.Split(fmt.Sprint(temp["UnitEYear"]), " ")[0]
	StationInfo["strpeakpower"] = strings.Split(fmt.Sprint(temp["StrPeakPower"]), " ")[0]
	StationInfo["invtotal"] = fmt.Sprint(temp["InvTotal"])

	StationInfo["strincome"] = strings.TrimPrefix(fmt.Sprint(temp["StrIncome"]), "€ ")

	for key, value := range StationInfo {
		fmt.Println(key + ": " + value)
		if mqttswitch == "y" {
			mqttpublish(mqttclient, "stationinfo", key, value)
		}
		body := strings.NewReader(key + ",tag=" + influxtag + " value=" + value)
		influxurl := "http://" + dbcon + "/write?db=" + database
		req, err := http.NewRequest("POST", os.ExpandEnv(influxurl), body)
		if err != nil {
			// handle err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// handle err
		}
		defer resp.Body.Close()
	}

	fmt.Println("-----------------------------------")

	POST_DATA := "page=1&perPage=20&orderBy=GATEWAYSN&whereCondition=%7B%22STATIONID%22%3A%22" + id + "%22%7D"

	var sent = false
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, "https://www.envertecportal.com/ApiInverters/QueryTerminalReal")
	easy.Setopt(curl.OPT_POST, true)
	easy.Setopt(curl.OPT_VERBOSE, false)
	easy.Setopt(curl.OPT_COOKIEJAR, "*cookiejar.Jar")

	easy.Setopt(curl.OPT_READFUNCTION,
		func(ptr []byte, userdata interface{}) int {
			// WARNING: never use append()
			if !sent {
				sent = true
				ret := copy(ptr, POST_DATA)
				return ret
			}
			return 0 // sent ok
		})

	// disable HTTP/1.1 Expect 100
	easy.Setopt(curl.OPT_HTTPHEADER, []string{"Expect:"})
	// must set
	easy.Setopt(curl.OPT_POSTFIELDSIZE, len(POST_DATA))

	var output map[string]interface{}
	// make a callback function
	fooTest := func(buf []byte, userdata interface{}) bool {
		//println("DEBUG: size=>", len(buf))
		//println("DEBUG: content=>", string(buf))
		json.Unmarshal([]byte(buf), &output)
		return true
	}

	easy.Setopt(curl.OPT_WRITEFUNCTION, fooTest)

	if err := easy.Perform(); err != nil {
		println("ERROR: ", err.Error())
	}

	outputtemp := output["Data"].(map[string]interface{})
	delete(outputtemp, "Lan")
	delete(outputtemp, "TotalPage")
	delete(outputtemp, "TotalCount")
	delete(outputtemp, "PageNumber")
	delete(outputtemp, "PerPage")

	for _, data := range outputtemp["QueryResults"].([]interface{}) {
		delete(data.(map[string]interface{}), "STATIONID")
		delete(data.(map[string]interface{}), "SNALIAS")
		delete(data.(map[string]interface{}), "SNID")
		delete(data.(map[string]interface{}), "SITETIME")
		for k, v := range data.(map[string]interface{}) {
			value := fmt.Sprintf("%v", v)
			key := strings.ToLower(k)
			fmt.Println(key + ": " + value)
			inverter := fmt.Sprintf("%v", data.(map[string]interface{})["SN"])

			if mqttswitch == "y" {
				mqttpublish(mqttclient, inverter, key, value)
			}

			body := strings.NewReader(key + ",tag=" + influxtag + ",inverter=" + inverter + " value=" + value)

			influxurl := "http://" + dbcon + "/write?db=" + database
			req, err := http.NewRequest("POST", os.ExpandEnv(influxurl), body)
			if err != nil {
				// handle err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// handle err
			}
			defer resp.Body.Close()

		}
		fmt.Println("-----------------------------------")
	}

	if mqttswitch == "y" {
		fmt.Println("Disconnect from MQTT Broker tcp://" + mqttbroker + ":" + mqttport)
		mqttdisconnect(mqttclient)
	}

}
