package geo_json

import (
	"fmt"
	"testing"
)

func TestGen(t *testing.T) {
	// 假设我们要查询的经纬度为乌鲁木齐
	//lat := 43.792818
	//lon := 87.617733

	lat := 28.19409
	lon := 112.982279

	// 初始化时加载GeoJSON数据文件
	cityRegions, err := LoadGeoJSONFileData(
		"E:\\Go_WorkSpace\\schisandra-album-cloud-microservices\\app\\auth\\resources\\geo_json\\world.zh.json",
		"E:\\Go_WorkSpace\\schisandra-album-cloud-microservices\\app\\auth\\resources\\geo_json\\china_province.json",
		"E:\\Go_WorkSpace\\schisandra-album-cloud-microservices\\app\\auth\\resources\\geo_json\\china_city.json",
	)
	if err != nil {
		fmt.Println("Error reading GeoJSON:", err)
		return
	}

	// 获取城市名称
	address, s, s2, err := GetAddress(lat, lon, cityRegions)
	if err != nil {
		fmt.Println("Error finding city:", err)
	}
	fmt.Println("Address:", address)
	fmt.Println("Province:", s)
	fmt.Println("City:", s2)
}
