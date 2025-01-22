package gao_map

import (
	"errors"
	"github.com/duke-git/lancet/v2/condition"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/validator"
	"net/url"
)

const (
	ipUrl    = "https://restapi.amap.com/v3/ip"
	geoUrl   = "https://restapi.amap.com/v3/geocode/geo"
	regeoUrl = "https://restapi.amap.com/v3/geocode/regeo"
	zoneUrl  = "https://restapi.amap.com/v3/config/district"
)

type Location struct {
	client *AmapClient
}

func (l *Location) ChinaDistricts() ([]*District, error) {
	val := url.Values{}
	val.Set("keywords", "中国")
	val.Set("subdistrict", "3")

	resp, err := l.client.DoRequest(zoneUrl, "GET", val)
	if err != nil {
		return nil, err
	}

	var data AllDistrictResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	} else if data.Count <= 0 {
		return nil, errors.New("no record")
	}

	return data.Districts[0].Districts, err
}

func (l *Location) IpLocation(ip string) (*IpResponse, error) {
	val := url.Values{}
	val.Set("ip", ip)

	resp, err := l.client.DoRequest(ipUrl, "GET", val)
	if err != nil {
		return nil, err
	}

	var data IpResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, nil
}

func (l *Location) ReGeo(request *ReGeoRequest) (*ReGeoResponse, error) {
	if validator.IsEmptyString(request.Location) {
		return nil, errors.New("location can't empty")
	}

	val := url.Values{}
	val.Set("location", request.Location)
	//if request.PoiType != "" {
	//	val.Set("poitype", request.PoiType)
	//}
	//val.Set("extensions", condition.Ternary(request.Extensions == "", "base", request.Extensions))
	val.Set("radius", convertor.ToString(condition.Ternary(request.Radius == 0, 1000, request.Radius)))
	//val.Set("roadlevel", convertor.ToString(request.RoadLevel))

	resp, err := l.client.DoRequest(regeoUrl, "GET", val)
	var data ReGeoResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, nil
}

func (l *Location) Geo(address, city string) (*GeoResponse, error) {
	val := url.Values{}
	val.Set("address", address)
	if city != "" {
		val.Set("city", city)
	}

	resp, err := l.client.DoRequest(geoUrl, "GET", val)
	var data GeoResponse

	if err = netutil.ParseHttpResponse(resp, &data); err != nil {
		return nil, err
	} else if err = checkResponse(data.Response); err != nil {
		return nil, err
	}

	return &data, nil
}
