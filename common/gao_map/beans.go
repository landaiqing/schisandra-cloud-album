package gao_map

type WeatherResponse struct {
	Response

	Lives []struct {
		Adcode           string `json:"adcode"`
		City             string `json:"city"`
		Province         string `json:"province"`
		Weather          string `json:"weather"`
		Temperature      string `json:"temperature"`
		WindDirection    string `json:"winddirection"`
		WindPower        string `json:"windpower"`
		Humidity         string `json:"humidity"`
		ReportTime       string `json:"reporttime"`
		TemperatureFloat string `json:"temperature_float"`
		HumidityFloat    string `json:"humidity_float"`
	} `json:"lives"`
	Forecast []struct {
		City       string `json:"city"`
		Adcode     string `json:"adcode"`
		Province   string `json:"province"`
		ReportTime string `json:"reporttime"`
		Casts      []struct {
			Date         string `json:"date"`
			Week         string `json:"week"`
			DayWeather   string `json:"dayweather"`
			NightWeather string `json:"nightweather"`
			DayTemp      string `json:"daytemp"`
			NightTemp    string `json:"nighttemp"`
			DayWind      string `json:"daywind"`
			NightWind    string `json:"nightwind"`
			DayPower     string `json:"daypower"`
			NightPower   string `json:"nightpower"`
		} `json:"casts"`
	} `json:"forecast"`
}

type IpResponse struct {
	Response
	Province  string `json:"province"`
	City      any    `json:"city"`
	Adcode    string `json:"adcode"`
	Rectangle string `json:"rectangle"`
}

type District struct {
	Citycode  any         `json:"citycode"`
	Adcode    string      `json:"adcode"`
	Name      string      `json:"name"`
	Center    string      `json:"center"`
	Level     string      `json:"level"`
	Districts []*District `json:"districts"`
}

type AllDistrictResponse struct {
	Response

	Count     int `json:"count,string"`
	Districts []*District
}

type Geo struct {
	FormattedAddress string `json:"formatted_address"`
	Country          string `json:"country"`
	Province         string `json:"province"`
	Citycode         string `json:"citycode"`
	City             any    `json:"city"`
	Adcode           string `json:"adcode"`
	Location         string `json:"location"`
	Level            string `json:"level"`
}

type GeoResponse struct {
	Response

	Count    string `json:"count"`
	GeoCodes []Geo  `json:"geocodes"`
}

type ReGeoResponse struct {
	Response

	ReGeoCode struct {
		FormattedAddress string `json:"formatted_address"`
		AddressComponent struct {
			City     any    `json:"city"`
			Country  string `json:"country"`
			Province string `json:"province"`
			Citycode string `json:"citycode"`
			District any    `json:"district"`
			Adcode   string `json:"adcode"`
			Township string `json:"township"`
			Towncode string `json:"towncode"`
		} `json:"addressComponent"`
	} `json:"regeocode"`
}

type ReGeoRequest struct {
	Location string
	//PoiType    string
	Radius int
	//Extensions string
	//RoadLevel  int
}
