package gao_map

import (
	"github.com/duke-git/lancet/v2/netutil"
	"net/url"
)

const weatherInfoUrl = "https://restapi.amap.com/v3/weather/weatherInfo"

type Weather struct {
	client *AmapClient
}

func (w *Weather) Info(adcode, extensions string) (*WeatherResponse, error) {
	val := url.Values{}
	val.Set("city", adcode)

	resp, err := w.client.DoRequest(weatherInfoUrl, "GET", val)
	if err != nil {
		return nil, err
	}

	var data WeatherResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, nil
}
