package main

import (
	"net/http"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"

	"github.com/robmccoll/samplr/gosamplr"
)

type SamplrHTTP struct {
	Samples *samplr.Samplr
}

type AddSampleRequest struct {
	Name        string
	Method      string
	URL         string
	Body        string
	Headers     http.Header
	Period      string
	SampleRange string
}

func (s *SamplrHTTP) AddSample(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var request = &AddSampleRequest{}

	if JSONRequest(w, r, request) {
		return
	}

	if len(request.Name) < 1 || len(request.URL) < 1 || len(request.Period) < 1 || len(request.SampleRange) < 1 {
		JSONError(w, http.StatusBadRequest, "Missing a required field (Name, URL, Period, SampleRange).")
		return
	}

	if len(request.Method) < 1 {
		request.Method = "GET"
	}

	period, err := time.ParseDuration(request.Period)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Parsing period: "+err.Error())
		return
	}

	sampleRange, err := time.ParseDuration(request.SampleRange)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Parsing SampleRange: "+err.Error())
		return
	}

	err = s.Samples.AddSampleSet(request.Name, request.Method, request.URL,
		[]byte(request.Body), request.Headers, period, sampleRange)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Adding samples failed: "+err.Error())
		return
	}

	JSONSuccess(w, http.StatusOK, request.Name+" added successfully.")
}

func (s *SamplrHTTP) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("name")

	err := s.Samples.RemoveSampleSet(name)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	JSONSuccess(w, http.StatusOK, name+" removed successfully.")
}

func (s *SamplrHTTP) SampleList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rtn, err := s.Samples.SampleSetNames()

	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to retrieve sample names - %v", err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, rtn)
}

func (s *SamplrHTTP) ReadN(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("name")
	scount := params.ByName("count")
	path := params.ByName("path")

	// eating error is fine, count = 0
	count, _ := strconv.Atoi(scount)

	samples, err := s.Samples.ReadNSamples(name, count)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(path) > 0 {
		samplePaths, err := ExtractPath(samples, path)
		if err != nil {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONResponse(w, http.StatusOK, samplePaths)
		return
	}

	JSONResponse(w, http.StatusOK, samples)
}

func (s *SamplrHTTP) ReadSince(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("name")
	stimestamp := params.ByName("timestamp")
	path := params.ByName("path")

	timestamp, err := time.Parse("", stimestamp)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Parsing timestamp failed: "+err.Error())
		return
	}

	samples, err := s.Samples.ReadSamplesSince(name, timestamp)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(path) > 0 {
		samplePaths, err := ExtractPath(samples, path)
		if err != nil {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONResponse(w, http.StatusOK, samplePaths)
		return
	}

	JSONResponse(w, http.StatusOK, samples)
}

func (s *SamplrHTTP) ReadRange(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	name := params.ByName("name")
	stimerange := params.ByName("timerange")
	path := params.ByName("path")

	timerange, err := time.ParseDuration(stimerange)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Parsing timestamp failed: "+err.Error())
		return
	}

	samples, err := s.Samples.ReadSamplesRange(name, timerange)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(path) > 0 {
		samplePaths, err := ExtractPath(samples, path)
		if err != nil {
			JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		JSONResponse(w, http.StatusOK, samplePaths)
		return
	}

	JSONResponse(w, http.StatusOK, samples)
}

func main() {
	samples := &SamplrHTTP{
		Samples: &samplr.Samplr{Sets: make(map[string]*samplr.SampleSet)},
	}

	router := httprouter.New()
	router.POST("/samples", samples.AddSample)

	router.GET("/samples", samples.SampleList)

	router.GET("/samples/:name/count/:count", samples.ReadN)
	router.GET("/samples/:name/since/:timestamp", samples.ReadSince)
	router.GET("/samples/:name/timerange/:timerange", samples.ReadRange)

	router.GET("/samples/:name/count/:count/:path", samples.ReadN)
	router.GET("/samples/:name/since/:timestamp/:path", samples.ReadSince)
	router.GET("/samples/:name/timerange/:timerange/:path", samples.ReadRange)

	router.DELETE("/samples/:name", samples.Delete)

	log.Fatal(http.ListenAndServe(":8080", router))
}
