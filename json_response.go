package main

import (
  "io/ioutil"
  "fmt"
  "net/http"
  "encoding/json"
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
  w.Write([]byte(`{"success":` + string(rtn) +`}`))
}

func JSONError(w http.ResponseWriter, code int, s string, args ...interface{}) {
  rtn,err := json.Marshal(fmt.Sprintf(s, args...))
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error":"marshalling json obj failed"}`))
    return
  }

  w.WriteHeader(code)
  w.Write([]byte(`{"error":` + string(rtn) +`}`))
}

