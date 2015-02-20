package influxdb

import (
	"encoding/json"
	"testing"
)

func TestInfluxFmt(t *testing.T) {
	s := `{"DR.deviceload.2XX.meter":{"15m.rate":0,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0},"DR.deviceload.3XX.meter":{"15m.rate":0,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0},"DR.deviceload.4XX.meter":{"15m.rate":0,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0},"DR.deviceload.5XX.meter":{"15m.rate":0,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0},"DR.deviceload.responsetime.histogram.exp":{"75%":0,"95%":0,"99%":0,"99.9%":0,"count":0,"max":0,"mean":0,"median":0,"min":0,"stddev":0},"DR.deviceload.responsetime.histogram.uni":{"75%":0,"95%":0,"99%":0,"99.9%":0,"count":0,"max":0,"mean":0,"median":0,"min":0,"stddev":0},"DR.devicestore.2XX.meter":{"15m.rate":0,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0},"DR.devicestore.3XX.meter":{"15m.rate":1,"1m.rate":0,"5m.rate":0,"count":0,"mean.rate":0}}`
	var m map[string]map[string]float64
	json.Unmarshal([]byte(s), &m)

	m2 := make(map[string]interface{})
	for key, value := range flatten(m) {
		m2[key] = value
	}
	influx_fmt([]map[string]interface{}{m2}, "em")
	//TODO OH GOD ADD SOMETHING
}
