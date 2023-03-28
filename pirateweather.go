package main

import (
	"encoding/json"
	"net/url"
	"net/http"
	"fmt"
	"flag"
	"time"
	"regexp"
//	"strconv"
	"io/ioutil"
	"os"

	gp "github.com/shawntoffel/go-pirateweather"
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

func getForecast(key string, latitude float64, longitude float64, time string) (gp.ForecastResponse, error) {
	var res gp.ForecastResponse
	surl := ""
	var err error
	if len(time) > 0 {
		surl, err = makeForecastURL(piratetime, key, fmt.Sprintf("%g,%g", latitude, longitude), time)
	} else {
		surl, err = makeForecastURL(piratebase, key, fmt.Sprintf("%g,%g", latitude, longitude), time)
	}

	if err != nil {
		return res, err
	}

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

	fmt.Printf("Weather at %s (%g, %g) is %s %g %s\n", match, lat, long, weather.Currently.Summary, weather.Currently.ApparentTemperature, weather.Flags.Units)
}
