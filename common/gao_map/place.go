package gao_map

import (
	"github.com/duke-git/lancet/v2/condition"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/strutil"
	"net/url"
)

const placeUrl = "https://restapi.amap.com/v5/place/text" //驾车

type Place struct {
	client *AmapClient
}

type PlaceRequest struct {
	Keywords   string
	Types      string
	Region     string
	CityLimit  bool
	PageSize   int
	PageNum    int
	ShowFields string
}

type PlaceResponse struct {
	Response

	Count int `json:"count,string"`
	Pois  []struct {
		Address  string `json:"address"`
		Adcode   string `json:"adcode"`
		Type     string `json:"type"`
		Typecode string `json:"typecode"`
		CityCode string `json:"citycode"`
		Name     string `json:"name"`
		Location string `json:"location"`
		Id       string `json:"id"`
		Business struct {
			Tel           string `json:"tel"`
			Cost          string `json:"cost"`
			Tag           string `json:"tag"`
			OpentimeToday string `json:"opentime_today"`
			OpentimeWeek  string `json:"opentime_week"`
		} `json:"business"`
		Photos []struct {
			Title string `json:"title"`
			Url   string `json:"url"`
		} `json:"photos"`
	} `json:"pois"`
}

func (d *Place) Search(request *PlaceRequest) (*PlaceResponse, error) {
	val := url.Values{}
	val.Set("keywords", request.Keywords)
	val.Set("types", request.Types)
	if !strutil.IsBlank(request.Region) {
		val.Set("region", request.Region)
	}
	val.Set("city_limit", condition.TernaryOperator(request.CityLimit, "1", "0"))
	val.Set("page_size", convertor.ToString(condition.TernaryOperator(request.PageSize > 0, request.PageSize, 25)))
	val.Set("page_num", convertor.ToString(condition.TernaryOperator(request.PageNum > 0, request.PageNum, 1)))
	if !strutil.IsBlank(request.ShowFields) {
		val.Set("show_fields", request.ShowFields)
	}

	resp, err := d.client.DoRequest(placeUrl, "GET", val)
	if err != nil {
		return nil, err
	}

	var data PlaceResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, err
}
