package samplr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/robmccoll/samplr/influxdb"
)

type Sample struct {
	Time time.Time
	Data []byte
}

type SampleSet struct {
	Lock sync.RWMutex
	Name string

	Method  string
	URL     string
	Body    []byte
	Headers http.Header

	Period      time.Duration
	SampleRange time.Duration
	Ticker      *time.Ticker
	Stopper     chan bool

	Samples []*Sample

	CallbackFunc func(Sample)
}

type ExpectedTimeAndCount struct {
	Count struct {
		Count int64 `json:"count"`
	} `json:"GLOBAL.snapshot.count"`
	Time struct {
		Value int64 `json:"value"`
	} `json:"GLOBAL.snapshot.time"`
}

func (set *SampleSet) Collect() {
	set.Ticker = time.NewTicker(set.Period)

	for {
		select {
		case _ = <-set.Ticker.C:
			{

				req, err := http.NewRequest(set.Method, set.URL, bytes.NewBuffer(set.Body))
				if err != nil {
					log.WithField("error", err).Error("Collect() creating request")
					continue
				}

				req.Header = set.Headers

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.WithField("error", err).Error("Collect() HTTP request")
					continue
				}

				if resp.StatusCode >= 300 {
					log.WithField("status", resp.Status).Error("Collect() HTTP request - non 2XX")
					continue
				}

				if resp.Body == nil {
					log.WithField("status", resp.Status).Error("Collect() HTTP request - no body")
					continue
				}

				body, err := ioutil.ReadAll(resp.Body)
				if resp.Body == nil {
					log.WithField("status", resp.Status).Error("Collect() HTTP request - read body")
					continue
				}

				resp.Body.Close()

				var now time.Time

				var timecount ExpectedTimeAndCount
				err = json.Unmarshal(body, &timecount)

				if err != nil {
					log.Warn("Can't parse snapshot time and count.")
					now = time.Now()
				} else {
					now = time.Unix(timecount.Time.Value, 0)
				}

				log.WithField("name", set.Name).WithField("time", now).Info("Adding entry")

				set.Lock.Lock()

				sample := &Sample{Time: now, Data: body}

				i := 0
				for ; i < len(set.Samples) && now.Sub(set.Samples[i].Time) > set.SampleRange; i++ {
					// pass
				}

				for j := 0; j < len(set.Samples)-i; j++ {
					set.Samples[j] = set.Samples[j+i]
				}

				if len(set.Samples) < 1 || now.After(set.Samples[len(set.Samples)-1].Time) {
					set.Samples = append(set.Samples[:len(set.Samples)-i], sample)
				} else {
					log.WithField("name", set.Name).WithField("time", now).Info("Skipping stale data")
				}

				set.Lock.Unlock()
				if set.CallbackFunc != nil {
					set.CallbackFunc(*sample)
				}
			}
		case _ = <-set.Stopper:
			close(set.Stopper)
			return
		}
	}
}

type Samplr struct {
	Lock      sync.RWMutex
	Sets      map[string]*SampleSet
	InfluxURL string
}

func (s *Samplr) AddSampleSet(name, method, url string, body []byte, headers http.Header, period time.Duration, sampleRange time.Duration) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if _, exists := s.Sets[name]; exists {
		return fmt.Errorf("SampleSet %v already exists.", name)
	}

	set := &SampleSet{
		Name:        name,
		Method:      method,
		URL:         url,
		Body:        body,
		Headers:     headers,
		Period:      period,
		SampleRange: sampleRange,
		Stopper:     make(chan bool, 1),
	}

	if s.InfluxURL != "" {
		set.CallbackFunc = PostToInfluxCallback(s.InfluxURL, name)
	}

	s.Sets[name] = set

	go set.Collect()

	return nil
}

func PostToInfluxCallback(influxURL string, name string) func(Sample) {
	return func(s Sample) {
		go influxdb.PostToInflux(influxURL, name, s.Data)
	}
}

func (s *Samplr) RemoveSampleSet(name string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	set, exists := s.Sets[name]

	if !exists {
		return fmt.Errorf("SampleSet %v does not exist.", name)
	}

	delete(s.Sets, name)

	set.Stopper <- true

	return nil
}

func (s *Samplr) ReadNSamples(name string, count int) ([]*Sample, error) {
	s.Lock.Lock()

	set, exists := s.Sets[name]

	s.Lock.Unlock()

	if !exists {
		return nil, fmt.Errorf("SampleSet %v does not exist.", name)
	}

	set.Lock.RLock()

	if count < 1 || count > len(set.Samples) {
		count = len(set.Samples)
	}

	rtn := make([]*Sample, 0, count)

	for i := 0; i < count; i++ {
		rtn = append(rtn, set.Samples[len(set.Samples)-i-1])
	}

	set.Lock.RUnlock()

	return rtn, nil
}

func (s *Samplr) ReadSamplesSince(name string, timestamp time.Time) ([]*Sample, error) {
	s.Lock.Lock()

	set, exists := s.Sets[name]

	s.Lock.Unlock()

	if !exists {
		return nil, fmt.Errorf("SampleSet %v does not exist.", name)
	}

	set.Lock.RLock()

	rtn := make([]*Sample, 0, 200)

	for i := 0; i < len(set.Samples) && set.Samples[len(set.Samples)-1].Time.After(timestamp); i++ {
		rtn = append(rtn, set.Samples[len(set.Samples)-i])
	}

	set.Lock.RUnlock()

	return rtn, nil
}

func (s *Samplr) ReadSamplesRange(name string, timerange time.Duration) ([]*Sample, error) {
	return s.ReadSamplesSince(name, time.Now().Add(-timerange))
}

func (s *Samplr) SampleSetNames() ([]string, error) {
	s.Lock.Lock()

	rtn := make([]string, 0, len(s.Sets))

	for k, _ := range s.Sets {
		rtn = append(rtn, k)
	}

	s.Lock.Unlock()

	return rtn, nil
}
