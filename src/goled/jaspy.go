package main

import (
	"encoding/json"
	"facette.io/natsort"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"
)

var ports prometheusResponseResults
var portspeeds map[string]float64


type jaspyInterface struct {
	Id int64 `json:"id"`
	index int64
	interfaceType string
	connectedInterface *int64
	DeviceId int64 `json:"deviceId"`
	displayName string
	name string
	alias string
	description string
	pollingEnabled *int64
	speedOverride *int64
	virtualConnection *int64
}


type prometheusResponse struct {
	Status string `json:"status"`
	Data prometheusResponseData `json:"data"`

}

type prometheusResponseData struct {
	ResultType string `json:"resultType"`
	Result []prometheusResponseResult `json:"result"`
}

type prometheusResponseResult struct {
	Metric prometheusResponseResultMetric `json:"metric"`
	Value []interface{} `json:"value"`
}

type prometheusResponseResults []prometheusResponseResult

func (a prometheusResponseResults) Len() int           { return len(a) }
func (a prometheusResponseResults) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a prometheusResponseResults) Less(i, j int) bool {
	if a[i].Metric.FQDN != a[j].Metric.FQDN {
		return natsort.Compare(a[i].Metric.FQDN, a[j].Metric.FQDN)
	}
	return natsort.Compare(a[i].Metric.Name, a[j].Metric.Name)
}



type prometheusResponseResultMetric struct {
	MetricName    string `json:"__name__"`
	FQDN          string `json:"fqdn"`
	Instance      string `json:"instance"`
	InterfaceType string `json:"interface_type"`
	Job           string `json:"job"`
	Name          string `json:"name"`
	Neighbors     string `json:"neighbors"`
}

type jaspyInterfaces []jaspyInterface

func (a jaspyInterfaces) Len() int           { return len(a) }
func (a jaspyInterfaces) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a jaspyInterfaces) Less(i, j int) bool {
	if a[i].DeviceId != a[j].DeviceId {
		return a[i].DeviceId > a[j].DeviceId
	}
	return a[i].Id < a[j].Id
}

func getPortStats() {

	go getPorts()
	go getPortSpeed()
}

func getPorts() {
	for {
		resp, err := http.Get("https://mobydick.netcrew.fi/prometheus/api/v1/query?query=jaspy_interface_up%7Binterface_type%3D%22ethernetCsmacd%22%7D")
		if err != nil {
			fmt.Printf("Error: %v", err)
			<-time.After(30*time.Second)
			continue
		}
		var m prometheusResponse
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %v", err)
			<-time.After(30*time.Second)
			continue
		}
		json.Unmarshal(b, &m)
		//fmt.Printf("data %+v\n", m)

		results := prometheusResponseResults(m.Data.Result)

		sort.Sort(results)

		ports = results
		resp.Body.Close()
		<-time.After(30*time.Second)
	}
}

func getPortSpeed() {
	for {

		resp, err := http.Get("https://mobydick.netcrew.fi/prometheus/api/v1/query?query=sum(rate(jaspy_interface_octets%7Binterface_type%3D%22ethernetCsmacd%22%2Cdirection%3D%22tx%22%7D%5B60s%5D)*8)%20by%20(fqdn%2Cname)%20%2F%20(sum(jaspy_interface_speed)%20by%20(fqdn%2Cname)*1000*1000)")
		if err != nil {
			fmt.Printf("Error: %v", err)
			<-time.After(30*time.Second)
			continue
		}
		var m prometheusResponse
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %v", err)
			<-time.After(30*time.Second)
			continue
		}
		json.Unmarshal(b, &m)

		resp.Body.Close()

		results := prometheusResponseResults(m.Data.Result)

		portspeeds_2 := make(map[string]float64)
		for idx := 0; idx < len(results); idx++ {
			speed_str := results[idx].Value[1].(string)
			speed, err := strconv.ParseFloat(speed_str, 64)
			if err != nil {
				fmt.Printf("Failed to parse %s to float", speed_str)
				continue
			}
			portspeeds_2[fmt.Sprintf("%s,%s", results[idx].Metric.FQDN, results[idx].Metric.Name)] = speed
		}

		portspeeds = portspeeds_2
		<-time.After(30*time.Second)
	}
}

func portStats() {
	var a = 0
	for {
		var x = 2
		var y = 2
		var up = 0
		var down = 0
		var trunkDown = 0
		var trunkUp = 0
		var prevSw = ""
		var col color.Color
		img := image.NewRGBA(image.Rect(0, 0, size_x*16, size_y*16))
		for idx := 0; idx < len(ports); idx++ {
			status, success := ports[idx].Value[1].(string)
			if !success {
				fmt.Printf("Status is not string")
				continue
			}

			//fmt.Printf("%s,%s\n", ports[idx].Metric.FQDN, ports[idx].Metric.Name)

			speed, ok := portspeeds[fmt.Sprintf("%s,%s", ports[idx].Metric.FQDN, ports[idx].Metric.Name)]

			if !ok {
				speed = 0.0
			}

			intensity := 5096+uint16((65535-5096)*speed)

			if status == "1" {

				if ports[idx].Metric.Neighbors == "yes" {
					col = color.RGBA64{0, 0, intensity, 65535}
					trunkUp += 1
				} else {
					col = color.RGBA64{0, intensity, 0, 65535}
					up += 1
				}

			} else {

				if ports[idx].Metric.Neighbors == "yes" {
					col = color.RGBA64{65535, 0, 65535, 65535}
					trunkDown += 1
				} else {
					col = color.RGBA64{intensity, 0, 0, 65535}
					down += 1
				}
			}

			if prevSw != ports[idx].Metric.FQDN {
				prevSw = ports[idx].Metric.FQDN
				x += 1
				y = 2
			}



			img.Set(x, y, col)

			if y >= size_y*16-20 {
				y = 2
				x += 1
			} else {
				y += 1
			}
		}
		addLabel(img, 10, 78, "Liikenne", color.RGBA{64, 64, 64, 255})
		addLabel(img, 10, 90, "porteittain", color.RGBA{64, 64, 64, 255})

		addLabel(img, 123, 90, fmt.Sprintf("%3d Mbps", speed), color.RGBA{64, 64, 64, 255})
		addLabel(img, 123, 10, fmt.Sprintf("Jaa %d", up), color.RGBA{64, 64, 64, 255})
		addLabel(img, 123, 30, fmt.Sprintf("Ei %d", down), color.RGBA{64, 64, 64, 255})
		addLabel(img, 123, 50, fmt.Sprintf("Tyhja %d", trunkUp), color.RGBA{64, 64, 64, 255})
		addLabel(img, 123, 70, fmt.Sprintf("Poissa %d", trunkDown), color.RGBA{64, 64, 64, 255})
		drawImage(img, uint8(a%254))
		a += 1
		<-time.After(30*time.Millisecond)
	}
}


func runPortStats() {
	go getPortStats()
	go getSpeed()
	portStats()
}