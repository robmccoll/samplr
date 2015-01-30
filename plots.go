package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotutil"
	"code.google.com/p/plotinum/vg/vgimg"
)

func (s *SamplrHTTP) LinePlotCount(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	scount := params.ByName("count")
	pathsparam := params.ByName("paths")

	// eating error is fine, count = 0
	count, _ := strconv.Atoi(scount)

	paths := strings.Split(pathsparam, ",")

	p, err := plot.New()
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	p.Title.Text = fmt.Sprintf("LineChart (%v)", time.Now())
	p.X.Label.Text = "Time (Unix)"
	p.Y.Label.Text = "Value"

	linepoints := make([]interface{}, 0, 2*len(paths))

	for _, path := range paths {
		nameAndPath := strings.SplitN(path, "|", 2)
		if len(nameAndPath) != 2 {
			JSONError(w, http.StatusBadRequest, "Name and path not formatted as expected (name|field|sub|field) [%v] %v", path, nameAndPath)
			return
		}

		samples, err := s.Samples.ReadNSamples(nameAndPath[0], count)
		if err != nil {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		samplePaths, err := ExtractPath(samples, nameAndPath[1])
		if err != nil {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		linepoints = append(linepoints, path)
		linepoints = append(linepoints, samplePaths)
	}

	err = plotutil.AddLinePoints(p, linepoints...)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	canvas := vgimg.PngCanvas{vgimg.New(600, 600)}
	p.Draw(plot.MakeDrawArea(canvas))

	w.WriteHeader(http.StatusOK)
	canvas.WriteTo(w)

}
