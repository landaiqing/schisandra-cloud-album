package gao_map

import (
	"errors"
	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/duke-git/lancet/v2/slice"
	"net/http"
	"net/url"
	"strings"
)

type AmapClient struct {
	key       string
	secret    string
	Location  *Location
	Direction *Direction
	Place     *Place
	Weather   *Weather
}

func NewAmapClient(key, secret string) *AmapClient {
	amap := &AmapClient{key: key, secret: secret}
	amap.Location = &Location{client: amap}
	amap.Direction = &Direction{client: amap}
	amap.Place = &Place{client: amap}
	amap.Weather = &Weather{client: amap}

	return amap
}

func (a *AmapClient) DoRequest(url, method string, params url.Values) (resp *http.Response, err error) {
	params.Set("key", a.key)
	if a.secret != "" {
		params.Set("sig", makeSign(a, &params))
	}

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = netutil.HttpGet(url, map[string]string{}, params)
	case "POST":
		resp, err = netutil.HttpPost(url, map[string]string{}, params)
	default:
		err = errors.New("unknown request method")
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("amap request error: " + resp.Status)
	}

	return
}

type Response struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"Infocode"`
}

func makeSign(client *AmapClient, values *url.Values) string {
	keys := maputil.Keys(*values)
	slice.SortBy(keys, func(a, b string) bool {
		return a < b
	})

	data := slice.Map(keys, func(index int, item string) string {
		return item + "=" + values.Get(item)
	})
	str := slice.Join(data, "&") + client.secret
	return cryptor.Md5String(str)
}

func checkResponse(response Response) error {
	if response.Status != "1" || response.Infocode != "10000" {
		return errors.New("amap response error: " + response.Info)
	}

	return nil
}
