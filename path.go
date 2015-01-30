package main

import (
	"code.google.com/p/plotinum/plotter"

	"github.com/robmccoll/jsonpath"
	"github.com/robmccoll/samplr/gosamplr"
)

func ExtractPath(samples []*samplr.Sample, path string) (plotter.XYs, error) {
	var err error

	rtn := make(plotter.XYs, len(samples))

	for i, sample := range samples {
		rtn[i].X = float64(sample.Time.UnixNano()) / 1000000000
		rtn[i].Y, err = jsonpath.ExtractFloat64(sample.Data, path)

		if err != nil {
			return nil, err
		}
	}

	return rtn, nil
}
