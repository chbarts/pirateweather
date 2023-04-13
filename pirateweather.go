package main

import (
	"encoding/json"
	"net/url"
	"net/http"
	"fmt"
	"flag"
	"time"
	"regexp"
	"io/ioutil"
	"os"
)

type TimeValue struct {
	Time *time.Time
}

func (t TimeValue) String() string {
	if t.Time != nil {
		return t.Time.String()
	}

	return ""
}

func MakeTime(str string) (time.Time, error) {
	const fmt = "2006-01-02T15:04:05"
	reh := regexp.MustCompile(`.+[tT](\d\d)`)
	rem := regexp.MustCompile(`.+[tT](\d\d):(\d\d)`)
	ret := regexp.MustCompile(`.+[tT](\d\d):(\d\d):(\d\d)`)
	rez := regexp.MustCompile(`.+([zZ]|([+\-](\d\d):(\d\d)))`)
	tnow := time.Now()
	location := tnow.Location()
	strs := ""
	if rez.MatchString(str) {
		if tm, err := time.Parse(time.RFC3339, str); err != nil {
			return tnow, err
		} else {
			return tm, nil
		}

	} else if ret.MatchString(str) {
		strs = str
	} else if rem.MatchString(str) {
		strs = str + ":00"
	} else if reh.MatchString(str) {
		strs = str + ":00:00"
	} else {
		strs = str + "T00:00:00"
	}

	if tm, err := time.ParseInLocation(fmt, strs, location); err != nil {
		return tnow, err
	} else {
		return tm, nil
	}
}

func (t TimeValue) Set(str string) error {
	if tm, err := MakeTime(str); err != nil {
		return err
	} else {
		*t.Time = tm
	}

	return nil
}

type ForecastOld struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Offset    float64 `json:"offset"`
	Currently struct {
		Time                int64   `json:"time"`
		Summary             string  `json:"summary"`
		Icon                string  `json:"icon"`
		PrecipIntensity     float64 `json:"precipIntensity"`
		PrecipType          string  `json:"precipType"`
		Temperature         float64 `json:"temperature"`
		ApparentTemperature float64 `json:"apparentTemperature"`
		DewPoint            float64 `json:"dewPoint"`
		Pressure            float64 `json:"pressure"`
		WindSpeed           float64 `json:"windSpeed"`
		WindBearing         float64 `json:"windBearing"`
		CloudCover          float64 `json:"cloudCover"`
	} `json:"currently"`
	Hourly struct {
		Data []struct {
			Time                int64   `json:"time"`
			Icon                string  `json:"icon"`
			Summary             string  `json:"summary"`
			PrecipAccumulation  float64 `json:"precipAccumulation"`
			PrecipType          string  `json:"precipType"`
			Temperature         float64 `json:"temperature"`
			ApparentTemperature float64 `json:"apparentTemperature"`
			DewPoint            float64 `json:"dewPoint"`
			Pressure            float64 `json:"pressure"`
			WindSpeed           float64 `json:"windSpeed"`
			WindBearing         float64 `json:"windBearing"`
			CloudCover          float64 `json:"cloudCover"`
		} `json:"data"`
	} `json:"hourly"`
	Daily struct {
		Data []struct {
			Time                        int64   `json:"time"`
			Icon                        string  `json:"icon"`
			Summary                     string  `json:"summary"`
			SunriseTime                 int64   `json:"sunriseTime"`
			SunsetTime                  int64   `json:"sunsetTime"`
			MoonPhase                   float64 `json:"moonPhase"`
			PrecipAccumulation          float64 `json:"precipAccumulation"`
			PrecipType                  string  `json:"precipType"`
			TemperatureHigh             float64 `json:"temperatureHigh"`
			TemperatureHighTime         int64   `json:"temperatureHighTime"`
			TemperatureLow              float64 `json:"temperatureLow"`
			TemperatureLowTime          int64   `json:"temperatureLowTime"`
			ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
			ApparentTemperatureHighTime int64   `json:"apparentTemperatureHighTime"`
			ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
			ApparentTemperatureLowTime  int64   `json:"apparentTemperatureLowTime"`
			DewPoint                    float64 `json:"dewPoint"`
			Pressure                    float64 `json:"pressure"`
			WindSpeed                   float64 `json:"windSpeed"`
			WindBearing                 float64 `json:"windBearing"`
			CloudCover                  float64 `json:"cloudCover"`
			TemperatureMin              float64 `json:"temperatureMin"`
			TemperatureMinTime          int64   `json:"temperatureMinTime"`
			TemperatureMax              float64 `json:"temperatureMax"`
			TemperatureMaxTime          int64   `json:"temperatureMaxTime"`
			ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
			ApparentTemperatureMinTime  int64   `json:"apparentTemperatureMinTime"`
			ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
			ApparentTemperatureMaxTime  int64   `json:"apparentTemperatureMaxTime"`
		} `json:"data"`
	} `json:"daily"`
}

type ForecastNew struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Offset    float64 `json:"offset"`
	Elevation int     `json:"elevation"`
	Currently struct {
		Time                 int64   `json:"time"`
		Summary              string  `json:"summary"`
		Icon                 string  `json:"icon"`
		NearestStormDistance int     `json:"nearestStormDistance"`
		NearestStormBearing  int     `json:"nearestStormBearing"`
		PrecipIntensity      float64 `json:"precipIntensity"`
		PrecipProbability    float64 `json:"precipProbability"`
		PrecipIntensityError float64 `json:"precipIntensityError"`
		PrecipType           string  `json:"precipType"`
		Temperature          float64 `json:"temperature"`
		ApparentTemperature  float64 `json:"apparentTemperature"`
		DewPoint             float64 `json:"dewPoint"`
		Humidity             float64 `json:"humidity"`
		Pressure             float64 `json:"pressure"`
		WindSpeed            float64 `json:"windSpeed"`
		WindGust             float64 `json:"windGust"`
		WindBearing          float64 `json:"windBearing"`
		CloudCover           float64 `json:"cloudCover"`
		UvIndex              float64 `json:"uvIndex"`
		Visibility           float64 `json:"visibility"`
		Ozone                float64 `json:"ozone"`
	} `json:"currently"`
	Minutely struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
		Data    []struct {
			Time                 int64   `json:"time"`
			PrecipIntensity      float64 `json:"precipIntensity"`
			PrecipProbability    float64 `json:"precipProbability"`
			PrecipIntensityError float64 `json:"precipIntensityError"`
			PrecipType           string  `json:"precipType"`
		} `json:"data"`
	} `json:"minutely"`
	Hourly struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
		Data    []struct {
			Time                 int64   `json:"time"`
			Icon                 string  `json:"icon"`
			Summary              string  `json:"summary"`
			PrecipIntensity      float64 `json:"precipIntensity"`
			PrecipProbability    float64 `json:"precipProbability"`
			PrecipIntensityError float64 `json:"precipIntensityError"`
			PrecipAccumulation   float64 `json:"precipAccumulation"`
			PrecipType           string  `json:"precipType"`
			Temperature          float64 `json:"temperature"`
			ApparentTemperature  float64 `json:"apparentTemperature"`
			DewPoint             float64 `json:"dewPoint"`
			Humidity             float64 `json:"humidity"`
			Pressure             float64 `json:"pressure"`
			WindSpeed            float64 `json:"windSpeed"`
			WindGust             float64 `json:"windGust"`
			WindBearing          float64 `json:"windBearing"`
			CloudCover           float64 `json:"cloudCover"`
			UvIndex              float64 `json:"uvIndex"`
			Visibility           float64 `json:"visibility"`
			Ozone                float64 `json:"ozone"`
		} `json:"data"`
	} `json:"hourly"`
	Daily struct {
		Summary string `json:"summary"`
		Icon    string `json:"icon"`
		Data    []struct {
			Time                        int64   `json:"time"`
			Icon                        string  `json:"icon"`
			Summary                     string  `json:"summary"`
			SunriseTime                 int64   `json:"sunriseTime"`
			SunsetTime                  int64   `json:"sunsetTime"`
			MoonPhase                   float64 `json:"moonPhase"`
			PrecipIntensity             float64 `json:"precipIntensity"`
			PrecipIntensityMax          float64 `json:"precipIntensityMax"`
			PrecipIntensityMaxTime      int64   `json:"precipIntensityMaxTime"`
			PrecipProbability           float64 `json:"precipProbability"`
			PrecipAccumulation          float64 `json:"precipAccumulation"`
			PrecipType                  string  `json:"precipType"`
			TemperatureHigh             float64 `json:"temperatureHigh"`
			TemperatureHighTime         int64   `json:"temperatureHighTime"`
			TemperatureLow              float64 `json:"temperatureLow"`
			TemperatureLowTime          int64   `json:"temperatureLowTime"`
			ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
			ApparentTemperatureHighTime int64   `json:"apparentTemperatureHighTime"`
			ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
			ApparentTemperatureLowTime  int64   `json:"apparentTemperatureLowTime"`
			DewPoint                    float64 `json:"dewPoint"`
			Humidity                    float64 `json:"humidity"`
			Pressure                    float64 `json:"pressure"`
			WindSpeed                   float64 `json:"windSpeed"`
			WindGust                    float64 `json:"windGust"`
			WindGustTime                int64   `json:"windGustTime"`
			WindBearing                 float64 `json:"windBearing"`
			CloudCover                  float64 `json:"cloudCover"`
			UvIndex                     float64 `json:"uvIndex"`
			UvIndexTime                 int64   `json:"uvIndexTime"`
			Visibility                  float64 `json:"visibility"`
			TemperatureMin              float64 `json:"temperatureMin"`
			TemperatureMinTime          int64   `json:"temperatureMinTime"`
			TemperatureMax              float64 `json:"temperatureMax"`
			TemperatureMaxTime          int64   `json:"temperatureMaxTime"`
			ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
			ApparentTemperatureMinTime  int64   `json:"apparentTemperatureMinTime"`
			ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
			ApparentTemperatureMaxTime  int64   `json:"apparentTemperatureMaxTime"`
		} `json:"data"`
	} `json:"daily"`
	Alerts []any `json:"alerts"`
	Flags  struct {
		Sources     []string `json:"sources"`
		SourceTimes struct {
			Hrrr018  string `json:"hrrr_0-18"`
			HrrrSubh string `json:"hrrr_subh"`
			Hrrr1848 string `json:"hrrr_18-48"`
			Gfs      string `json:"gfs"`
			Gefs     string `json:"gefs"`
		} `json:"sourceTimes"`
		NearestStation int    `json:"nearest-station"`
		Units          string `json:"units"`
		Version        string `json:"version"`
	} `json:"flags"`
}

type Minutely struct {
	Time                int64
	PrecipIntenisty     float64
	PrecipProbability   float64
	PrecipType          string
}

type Hourly struct {
	Time                int64
	PrecipAccumulation  float64
	PrecipType          string
	Temperature         float64
	ApparentTemperature float64
	WindSpeed           float64
	WindBearing         float64
	PrecipIntensity     float64
	WindGust            float64
}

type Daily struct {
	Time                int64
	Summary             string
	TemperatureHigh     float64
	TemperatureLow      float64
	PrecipAccumulation  float64
	PrecipType          string
	WindSpeed           float64
	WindGust            float64
	WindBearing         float64
}

type Weather struct {
	Summary             string
	Temperature         float64
	ApparentTemperature float64
	Minutes             []Minutely
	Hours               []Hourly
	Days                []Daily
}

type CensusGeocode struct {
	Result struct {
		Input struct {
			Address struct {
				Address string `json:"address"`
			} `json:"address"`
			Benchmark struct {
				IsDefault            bool   `json:"isDefault"`
				BenchmarkDescription string `json:"benchmarkDescription"`
				ID                   string `json:"id"`
				BenchmarkName        string `json:"benchmarkName"`
			} `json:"benchmark"`
		} `json:"input"`
		AddressMatches []struct {
			TigerLine struct {
				Side        string `json:"side"`
				TigerLineID string `json:"tigerLineId"`
			} `json:"tigerLine"`
			Coordinates struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"coordinates"`
			AddressComponents struct {
				Zip             string `json:"zip"`
				StreetName      string `json:"streetName"`
				PreType         string `json:"preType"`
				City            string `json:"city"`
				PreDirection    string `json:"preDirection"`
				SuffixDirection string `json:"suffixDirection"`
				FromAddress     string `json:"fromAddress"`
				State           string `json:"state"`
				SuffixType      string `json:"suffixType"`
				ToAddress       string `json:"toAddress"`
				SuffixQualifier string `json:"suffixQualifier"`
				PreQualifier    string `json:"preQualifier"`
			} `json:"addressComponents"`
			MatchedAddress string `json:"matchedAddress"`
		} `json:"addressMatches"`
	} `json:"result"`
}

var piratebase = "https://api.pirateweather.net/";
var piratetime = "https://timemachine.pirateweather.net/";

func makeForecastURL(base string, key string, location string, time string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	baseURL.Path += "forecast/"
	baseURL.Path += key
	baseURL.Path += "/"
	baseURL.Path += location
	if len(time) > 0 {
		baseURL.Path += ","
		baseURL.Path += time
	}

	return baseURL.String(), nil
}

func getData(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func makeNewForecast(surl string) (ForecastNew, error) {
	var res ForecastNew
	data, err := getData(surl)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal([]byte(data), &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func makeOldForecast(surl string) (ForecastOld, error) {
	var res ForecastOld
	data, err := getData(surl)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal([]byte(data), &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func getForecast(key string, latitude float64, longitude float64, time string) (Weather, error) {
	var res Weather
	surl := ""
	var err error
	if len(time) > 0 {
		surl, err = makeForecastURL(piratetime, key, fmt.Sprintf("%g,%g", latitude, longitude), time)
		if err != nil {
			return res, err
		}

		odata, err := makeOldForecast(surl)
		if err != nil {
			return res, err
		}

		res.Summary = odata.Currently.Summary
		res.Temperature = odata.Currently.Temperature
		res.ApparentTemperature = odata.Currently.ApparentTemperature
		hourdat := odata.Hourly.Data
		for i:= 0; i < len(hourdat); i++ {
			res.Hours = append(res.Hours, Hourly{hourdat[i].Time, hourdat[i].PrecipAccumulation, hourdat[i].PrecipType, hourdat[i].Temperature, hourdat[i].ApparentTemperature, hourdat[i].WindSpeed, hourdat[i].WindBearing, hourdat[i].PrecipIntensity, -1.0})
		}

		daydat := odata.Daily.Data
		for i := 0; i < len(daydat); i++ {
			res.Days = append(res.Days, Daily{daydat[i].Time, daydat[i].Summary, daydat[i].TemperatureHigh, daydat[i].TemperatureLow, daydat[i].PrecipAccumulation, daydat[i].PrecipType, daydat[i].WindSpeed, -1.0, daydat[i].WindBearing})
		}

	} else {
		surl, err = makeForecastURL(piratebase, key, fmt.Sprintf("%g,%g", latitude, longitude), time)
		if err != nil {
			return res, err
		}

		ndata, err := makeNewForecast(surl)
		if err != nil {
			return res, err
		}

		res.Summary = ndata.Currently.Summary
		res.Temperature = ndata.Currently.Temperature
		res.ApparentTemperature = ndata.Currently.ApparentTemperature
		mindat := ndata.Minutely.Data
		for i := 0; i < len(mindat); i++ {
			res.Minutes = append(res.Minutes, Minutely{mindat[i].Time, mindat[i].PrecipIntensity, mindat[i].PrecipProbability, mindat[i].PrecipType})
		}

		hourdat := ndata.Hourly.Data
		for i:= 0; i < len(hourdat); i++ {
			res.Hours = append(res.Hours, Hourly{hourdat[i].Time, hourdat[i].PrecipAccumulation, hourdat[i].PrecipType, hourdat[i].Temperature, hourdat[i].ApparentTemperature, hourdat[i].WindSpeed, hourdat[i].WindBearing, hourdat[i].PrecipIntensity, hourdat[i].WindGust})
		}

		daydat := ndata.Daily.Data
		for i := 0; i < len(daydat); i++ {
			res.Days = append(res.Days, Daily{daydat[i].Time, daydat[i].Summary, daydat[i].TemperatureHigh, daydat[i].TemperatureLow, daydat[i].PrecipAccumulation, daydat[i].PrecipType, daydat[i].WindSpeed, daydat[i].WindGust, daydat[i].WindBearing})
		}
	}


	return res, nil
}

func getLocation(address string) (CensusGeocode, error) {
	var res CensusGeocode
	baseURL, err := url.Parse("https://geocoding.geo.census.gov/geocoder/locations/onelineaddress")
	if err != nil {
		return res, err
	}

	query := baseURL.Query()
	query.Add("address", address)
	query.Add("benchmark", "2020")
	query.Add("format", "json")
	baseURL.RawQuery = query.Encode()

	data, err := http.Get(baseURL.String())
	if err != nil {
		return res, err
	}

	content, err := ioutil.ReadAll(data.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal([]byte(string(content)), &res)
	if err != nil {
		return res, err
	}

	return res, nil
}

var tstart = &time.Time{}
var loc = flag.String("location", "1600 Pennsylvania Avenue NW, Washington, DC 20500", "location in the United States of America")
var minutely = flag.Bool("minutely", false, "Show minutely forecast on current weather only")
var hourly = flag.Bool("hourly", false, "Show hourly forecast on current or old weather")
var daily = flag.Bool("daily", false, "Show daily forecast on current or old weather")
var tzeug = time.Date(1880, time.November, 10, 23, 0, 0, 0, time.UTC)

func main() {
	tstr := ""
	*tstart = tzeug
	flag.Var(&TimeValue{tstart}, "time", "time to get weather from, RFC 3339 format with optional time and time zone, default to local time (2017-11-01[T00:00:00[-07:00]])")
	flag.Parse()
	key := os.Getenv("PIRATEWEATHER")
	if len(key) == 0 {
		panic("No API key set at PIRATEWEATHER environment variable")
	}

	if *tstart != tzeug {
		tstr = fmt.Sprintf("%d", tstart.Unix())
	}

	geo, err := getLocation(*loc)
	if err != nil {
		panic(err)
	}

	if len(geo.Result.AddressMatches) == 0 {
		panic("No address matched")
	}

	var lat, long float64
	match := ""
	if len(geo.Result.AddressMatches) == 1 {
		lat = geo.Result.AddressMatches[0].Coordinates.Y
		long = geo.Result.AddressMatches[0].Coordinates.X
		match = geo.Result.AddressMatches[0].MatchedAddress
	} else {
		for ind, elt := range geo.Result.AddressMatches {
			fmt.Printf("%d\t%s\n", ind, elt.MatchedAddress)
		}

		fmt.Print("Pick location number: ")
		var i int
		_, err := fmt.Scanf("%d", &i)
		if err != nil {
			panic(err)
		}

		if (i > len(geo.Result.AddressMatches)) || (i < 0) {
			panic("Invalid number")
		}

		lat = geo.Result.AddressMatches[i].Coordinates.Y
		long = geo.Result.AddressMatches[i].Coordinates.X
		match = geo.Result.AddressMatches[i].MatchedAddress
	}

	weather, err := getForecast(key, lat, long, tstr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Weather at %s (%g, %g) is %s %g (feels like %g)\n", match, lat, long, weather.Summary, weather.Temperature, weather.ApparentTemperature)
	if *daily {
		for i := 0; i < len(weather.Days); i++ {
			day := weather.Days[i]
			fmt.Printf("%v\t%s High: %g Low: %g Accum: %g (%s) Wind Speed: %g Bearing %g Gust %g\n", time.Unix(day.Time, 0).Local(), day.Summary, day.TemperatureHigh, day.TemperatureLow, day.PrecipAccumulation, day.PrecipType, day.WindSpeed, day.WindBearing, day.WindGust)
		}
	}
}
