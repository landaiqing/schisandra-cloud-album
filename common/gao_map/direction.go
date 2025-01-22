package gao_map

import (
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/strutil"
	"net/url"
	"strings"
)

const drivingUrl = "https://restapi.amap.com/v5/direction/driving" //驾车

type Direction struct {
	client *AmapClient
}

type DrivingRequest struct {
	Origin      string
	Destination string
	Strategy    int
	Waypoints   []string
	ShowFields  string
}

type BicyclingRequest struct {
	Origin      string
	Destination string
	ShowFields  string
}

type Route struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Paths       []struct {
		Distance float64 `json:"distance,string"`

		Steps []struct {
			Cost struct {
				Duration      float64 `json:"duration,string"`
				Tolls         float64 `json:"tolls,string"`
				TollsDistance float64 `json:"tolls_distance,string"`
			} `json:"cost"`
			Cities []struct {
				Adcode   string `json:"adcode"`
				Citycode string `json:"citycode"`
				City     string `json:"city"`
			} `json:"cities"`
			Polyline string `json:"polyline"`
		} `json:"steps"`
	} `json:"paths"`
}

type DrivingResponse struct {
	Response

	Count int   `json:"count,string"`
	Route Route `json:"route"`
}

type BicyclingResponse struct {
	Response

	Count int   `json:"count,string"`
	Route Route `json:"route"`
}

func (d *Direction) Bicycling(request *BicyclingRequest) (*BicyclingResponse, error) {
	val := url.Values{}
	val.Set("origin", request.Origin)
	val.Set("destination", request.Destination)
	if !strutil.IsBlank(request.ShowFields) {
		val.Set("show_fields", request.ShowFields)
	}

	resp, err := d.client.DoRequest(drivingUrl, "POST", val)
	if err != nil {
		return nil, err
	}

	var data BicyclingResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, err
}

func (d *Direction) Driving(request *DrivingRequest) (*DrivingResponse, error) {
	val := url.Values{}
	val.Set("origin", request.Origin)
	val.Set("destination", request.Destination)
	val.Set("strategy", convertor.ToString(request.Strategy))
	if len(request.Waypoints) > 0 {
		val.Set("waypoints", strings.Join(request.Waypoints, ";"))
	}
	if !strutil.IsBlank(request.ShowFields) {
		val.Set("show_fields", request.ShowFields)
	}

	resp, err := d.client.DoRequest(drivingUrl, "POST", val)
	if err != nil {
		return nil, err
	}

	var data DrivingResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, err
}
