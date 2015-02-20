package influxdb

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
)

//flatten should flatten a map once, rolling the string up
func flatten(in map[string]map[string]float64) map[string]float64 {
	out := make(map[string]float64)
	for key, value := range in {
		for key2, value := range value {
			out[key+"."+key2] = value
		}
	}
	return out
}

func influx_fmt(maps []map[string]float64, name string) []map[string]interface{} {
	out := make(map[string]interface{})
	out["name"] = name
	cols := make([]string, 0)
	for k, _ := range maps[0] {
		cols = append(cols, k)
	}
	out["columns"] = cols

	points := make([][]interface{}, 0)
	for _, m := range maps {
		row := make([]interface{}, 0)
		for _, point := range m {
			row = append(row, point)
		}
		points = append(points, row)
	}
	out["points"] = points
	return []map[string]interface{}{out}
}

func PostToInflux(influxURL string, name string, b []byte) {
	logrus.Info("Posing to influx")
	m := make(map[string]map[string]float64)
	err := json.Unmarshal(b, &m)
	if err != nil {
		logrus.Error(string(b))
		logrus.WithField("err", err).Error("Failed to unmarshal as expected")
	}
	fmtted := influx_fmt([]map[string]float64{flatten(m)}, name)

	b, err = json.Marshal(fmtted)
	if err != nil {
		logrus.Error(err)
		return
	}
	resp, err := http.Post(influxURL, "application/json", bytes.NewReader(b))
	if err != nil {
		logrus.Error(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		respbody, _ := ioutil.ReadAll(resp.Body)
		logrus.WithField("StatusCode", resp.StatusCode).WithField("resp.body", string(respbody)).WithField("Posted", string(b)).Error("Non 200 from influxdb")
	}

	logrus.Info("Posted to influx")
}
