package main

import (
  "io/ioutil"
  "fmt"
  "net/http"
  "encoding/json"

  log "github.com/Sirupsen/logrus"
  "github.com/julienschmidt/httprouter"

  "github.com/robmccoll/samplr/samplr"
)

func JSONRequest(w http.ResponseWriter, r *http.Request, obj interface{}) (isBad bool) {
  if r.Body == nil {
    JSONError(w, http.StatusBadRequest, "A JSON body is required.")
    return true
  }

  body, err := ioutil.ReadAll(r.Body)
  if r.Body == nil {
    JSONError(w, http.StatusBadRequest, "A JSON body is required.")
    return true
  }

  r.Body.Close()

  err = json.Unmarshal(body, obj)
  if err != nil {
    JSONError(w, http.StatusBadRequest, "A JSON body is required.")
    return true
  }

  return false
}

func JSONResponse(w http.ResponseWriter, code int, obj interface{}) {
  rtn,err := json.Marshal(obj)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error":"marshalling json obj failed"}`))
    return
  }

  w.WriteHeader(code)
  w.Write(rtn)
}

func JSONSuccess(w http.ResponseWriter, code int, s string, args ...interface{}) {
  rtn,err := json.Marshal(fmt.Sprintf(s, args...))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error":"marshalling json obj failed"}`))
    return
  }

  w.WriteHeader(code)
  w.Write([]byte(`{"success":"` + string(rtn) +`"}`))
}

func JSONError(w http.ResponseWriter, code int, s string, args ...interface{}) {
  rtn,err := json.Marshal(fmt.Sprintf(s, args...))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error":"marshalling json obj failed"}`))
    return
  }

  w.WriteHeader(code)
  w.Write([]byte(`{"error":"` + string(rtn) +`"}`))
}

type SamplrHTTP struct {
  Samples *samplr.Samplr
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

  _,_ = name,scount
}

func (s *SamplrHTTP) ReadSince(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
  name := params.ByName("name")
  stimestamp := params.ByName("timestamp")

  _,_ = name,stimestamp
}

func (s *SamplrHTTP) ReadRange(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
  name := params.ByName("name")
  stimerange := params.ByName("timerange")

  _,_ = name,stimerange
}

func main() {
  samples := &SamplrHTTP {
    Samples: &samplr.Samplr{ Sets: make(map[string]*samplr.SampleSet) },
  }

  router := httprouter.New()
  router.GET("/samples", samples.SampleList)
  router.GET("/samples/:name/count/:count", samples.ReadN)
  router.GET("/samples/:name/since/:timestamp", samples.ReadSince)
  router.GET("/samples/:name/timerange/:timerange", samples.ReadRange)

  log.Fatal(http.ListenAndServe(":8080", router))
}
