package geo_json

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"os"
	"path/filepath"
	"sync"
)

// RegionType 定义一个类型，用来标识区域是国家、省份还是城市
type RegionType string

const (
	Country  RegionType = "country"
	Province RegionType = "province"
	City     RegionType = "city"
)

// CityRegion 结构体，包含几何数据和城市名称
type CityRegion struct {
	Geometry   orb.Geometry
	Name       string
	Bounds     orb.Bound // 缓存边界
	RegionType RegionType
}

// RegionData 结构体，包含国家、省份和城市的数据
type RegionData struct {
	Countries []CityRegion
	Provinces []CityRegion
	Cities    []CityRegion
}

// LoadGeoJSONFileData 加载三个GeoJSON文件并返回RegionData
func LoadGeoJSONFileData(countryFile, provinceFile, cityFile string) (*RegionData, error) {
	// 创建RegionData实例来存储国家、省份、城市数据
	regionData := &RegionData{}

	// 加载国家数据
	countries, err := loadGeoJSONData(countryFile, Country)
	if err != nil {
		return nil, fmt.Errorf("failed to load countries from %s: %v", countryFile, err)
	}
	regionData.Countries = countries

	// 加载省份数据
	provinces, err := loadGeoJSONData(provinceFile, Province)
	if err != nil {
		return nil, fmt.Errorf("failed to load provinces from %s: %v", provinceFile, err)
	}
	regionData.Provinces = provinces

	// 加载城市数据
	cities, err := loadGeoJSONData(cityFile, City)
	if err != nil {
		return nil, fmt.Errorf("failed to load cities from %s: %v", cityFile, err)
	}
	regionData.Cities = cities

	return regionData, nil
}

// loadGeoJSONData 读取GeoJSON文件并根据给定的区域类型返回CityRegion数据
func loadGeoJSONData(filename string, regionType RegionType) ([]CityRegion, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filename, err)
	}

	var cityRegions []CityRegion
	fc, err := geojson.UnmarshalFeatureCollection(file)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", filename, err)
	}

	for _, feature := range fc.Features {
		if feature.Geometry != nil {
			// 将区域名称和几何数据一起保存，附加区域类型
			cityRegions = append(cityRegions, CityRegion{
				Geometry:   feature.Geometry,
				Name:       feature.Properties["name"].(string),
				RegionType: regionType,
			})
		} else {
			return nil, fmt.Errorf("feature has no geometry %v", feature.ID)
		}
	}

	return cityRegions, nil
}

// GetCityName 根据经纬度和给定的CityRegion判断城市
func getCityName(lat, lon float64, cityRegions []CityRegion) (string, error) {
	point := orb.Point{lon, lat}

	// 遍历城市区域
	for _, region := range cityRegions {
		switch geo := region.Geometry.(type) {
		case orb.Polygon:
			// 如果是Polygon，检查点是否在多边形内
			if planar.PolygonContains(geo, point) {
				return region.Name, nil
			}
		case orb.MultiPolygon:
			// 如果是MultiPolygon，检查点是否在任何一个多边形内
			if planar.MultiPolygonContains(geo, point) {
				return region.Name, nil
			}
		default:
			return "", fmt.Errorf("unsupported geometry type: %v", geo.GeoJSONType())
		}
	}

	return "", fmt.Errorf("point not found in any city region")
}

// GetAddress 根据经纬度识别国家、省份和市
func GetAddress(lat, lon float64, regionData *RegionData) (string, string, string, error) {
	// Step 1: 识别国家
	country, err := getCityName(lat, lon, regionData.Countries)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to identify country: %v", err)
	}
	if country != "中国" {
		return country, "", "", nil
	}
	var wg sync.WaitGroup
	var province, city string
	var provinceErr, cityErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		province, provinceErr = getCityName(lat, lon, regionData.Provinces)
	}()

	go func() {
		defer wg.Done()
		city, cityErr = getCityName(lat, lon, regionData.Cities)
	}()

	wg.Wait()

	if provinceErr != nil {
		return country, "", "", fmt.Errorf("failed to identify province: %v", provinceErr)
	}
	if cityErr != nil {
		return country, province, "", fmt.Errorf("failed to identify city: %v", cityErr)
	}

	return country, province, city, nil
}

func NewGeoJSON() *RegionData {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
		return nil
	}
	countryFile := filepath.Join(dir, "/resources/geo_json/world.zh.json")
	provinceFile := filepath.Join(dir, "/resources/geo_json/china_province.json")
	cityFile := filepath.Join(dir, "/resources/geo_json/china_city.json")
	regionData, err := LoadGeoJSONFileData(countryFile, provinceFile, cityFile)
	if err != nil {
		panic(err)
		return nil
	}
	return regionData
}
