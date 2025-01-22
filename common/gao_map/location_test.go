package gao_map

import (
	"fmt"
	"testing"
)

const key = ""
const secret = ""

func TestDistricts(t *testing.T) {
	client := NewAmapClient(key, secret)
	resp, err := client.Location.ChinaDistricts()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp)
}

func TestIp(t *testing.T) {
	client := NewAmapClient(key, secret)
	resp, err := client.Location.IpLocation("121.224.33.99")

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(*resp)
}

func TestGeo(t *testing.T) {
	client := NewAmapClient(key, secret)
	resp, err := client.Location.Geo("乐东公园一品墅", "")

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(*resp)
}

func TestReGeo(t *testing.T) {
	client := NewAmapClient(key, secret)
	request := ReGeoRequest{Location: "118.567824,29.306175"}
	resp, err := client.Location.ReGeo(&request)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(*resp)
}
