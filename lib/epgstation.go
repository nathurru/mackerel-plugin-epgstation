package mpepgstation

import (
	"encoding/json"
	"flag"
	"fmt"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type EPGStationPlugin struct {
	Prefix string
	Target string
}

type Stream []struct{}

type Rule struct {
	Total int `json:"total"`
}

type Overlap struct {
	Total int `json:"total"`
}

type Conflict struct {
	Total int `json:"total"`
}

type Skip struct {
	Total int `json:"total"`
}

type Record struct {
	Total int `json:"total"`
}

type Recording struct {
	Total int `json:"total"`
}

type Recorded struct {
	Total int `json:"total"`
}

type Encode struct {
	Queue    []struct{} `json:"queue"`
	Encoding struct {
		ID string `json:"id"`
	} `json:"encoding"`
}

func GetAPI(host string, path string) []byte {
	url := fmt.Sprintf("http://%s/api/%s", host, path)

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Http Error:", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	byteArray, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("response Error:", err)
		os.Exit(1)
	}

	return byteArray
}

func GetStreams(host string) float64 {
	var stream Stream
	byteArray := GetAPI(host, "streams/info")
	json.Unmarshal(byteArray, &stream)

	return float64(len(stream))
}

func GetSkipPrograms(host string) float64 {
	var skip Skip
	byteArray := GetAPI(host, "reserves/skips?limit=1")
	json.Unmarshal(byteArray, &skip)

	return float64(skip.Total)
}

func GetOverlapPrograms(host string) float64 {
	var overlap Overlap
	byteArray := GetAPI(host, "reserves/overlaps?limit=1")
	json.Unmarshal(byteArray, &overlap)

	return float64(overlap.Total)
}

func GetDuplicatePrograms(host string) float64 {
	var conflict Conflict
	byteArray := GetAPI(host, "reserves/conflicts?limit=1")
	json.Unmarshal(byteArray, &conflict)

	return float64(conflict.Total)
}

func GetSchedulePrograms(host string) float64 {
	var record Record
	byteArray := GetAPI(host, "reserves?limit=1")
	json.Unmarshal(byteArray, &record)

	return float64(record.Total)
}

func GetRecordingPrograms(host string) float64 {
	var recording Recording
	byteArray := GetAPI(host, "recorded?limit=1&recording=true")
	json.Unmarshal(byteArray, &recording)

	return float64(recording.Total)
}

func GetRecordedPrograms(host string) float64 {
	var recorded Recorded
	byteArray := GetAPI(host, "recorded?limit=1&recording=false")
	json.Unmarshal(byteArray, &recorded)

	return float64(recorded.Total)
}

func GetRules(host string) float64 {
	var rule Rule
	byteArray := GetAPI(host, "rules?limit=1&offset=0")
	json.Unmarshal(byteArray, &rule)

	return float64(rule.Total)
}

func GetEncodeQueues(host string) float64 {
	var rule Encode
	var total int
	byteArray := GetAPI(host, "encode")
	json.Unmarshal(byteArray, &rule)
	total = len(rule.Queue)

	if rule.Encoding.ID != "" {
		total++
	}

	return float64(total)
}

func (e EPGStationPlugin) MetricKeyPrefix() string {
	if e.Prefix == "" {
		e.Prefix = "EPGStation"
	}
	return e.Prefix
}

func (e EPGStationPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(e.Prefix)
	return map[string]mp.Graphs{
		"stream": mp.Graphs{
			Label: fmt.Sprintf("%s Stream", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "stream", Label: "Streams"},
			},
		},
		"rule": mp.Graphs{
			Label: fmt.Sprintf("%s Rule", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "rule", Label: "Rules"},
			},
		},
		"record": mp.Graphs{
			Label: fmt.Sprintf("%s Recording Reservation", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "schedule", Label: "Schedule Programs"},
				{Name: "skip", Label: "Skip Programs"},
				{Name: "overlap", Label: "Overlap Programs"},
				{Name: "duplicate", Label: "Duplicate Programs"},
			},
		},
		"recording": mp.Graphs{
			Label: fmt.Sprintf("%s Recoding", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "recording", Label: "Recording Programs"},
			},
		},
		"recorded": mp.Graphs{
			Label: fmt.Sprintf("%s Recoded", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "recorded", Label: "Recoded Programs"},
			},
		},
		"encode": mp.Graphs{
			Label: fmt.Sprintf("%s Encode", labelPrefix),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "queue", Label: "Encode Queues"},
			},
		},
	}
}

func (e EPGStationPlugin) FetchMetrics() (map[string]float64, error) {

	return map[string]float64{
		"stream":    GetStreams(e.Target),
		"rule":      GetRules(e.Target),
		"schedule":  GetSchedulePrograms(e.Target),
		"skip":      GetSkipPrograms(e.Target),
		"overlap":   GetOverlapPrograms(e.Target),
		"duplicate": GetDuplicatePrograms(e.Target),
		"recording": GetRecordingPrograms(e.Target),
		"recorded":  GetRecordedPrograms(e.Target),
		"queue":     GetEncodeQueues(e.Target),
	}, nil
}

func Do() {
	optPrefix := flag.String("metric-key-prefix", "EPGStation", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")

	optHost := flag.String("host", "127.0.0.1", "EPGStation hostname")
	optPort := flag.String("port", "8888", "EPGStation port")
	flag.Parse()

	e := EPGStationPlugin{
		Target: fmt.Sprintf("%s:%s", *optHost, *optPort),
		Prefix: *optPrefix,
	}

	helper := mp.NewMackerelPlugin(e)
	helper.Tempfile = *optTempfile
	helper.Run()
}
