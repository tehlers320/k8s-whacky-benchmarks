package client

import (
	"net/http"
	"io"
	"bytes"
	"encoding/json"

	fortio "fortio.org/fortio/rapi"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	MetadataUri = "?jsonPath=.metadata"
)
type FortioRest struct {
	Metadata `json:"metadata"`
}

type Metadata struct {
	URL              string `json:"url,omitempty"`
	Connections      string `json:"c,omitempty"`
	Numcalls         string `json:"n,omitempty"`
	Async            string `json:"async,omitempty"`
	Save             string `json:"save,omitempty"`
	Payload          string `json:"payload,omitempty"`
	Labels           string `json:"labels,omitempty"`
	Resolution       string `json:"r,omitempty"`
	PercList         string `json:"p,omitempty"`
	Qps              string `json:"qps:omitempty"`
	DurStr           string `json:"t,omitempty"`
	Jitter           string `json:"jitter,omitempty"`
	Uniform          string `json:"uniform,omitempty"`
	Nocatchup        string `json:"nocatchup,omitempty"`
	StdClient        string `json:"stdclient,omitempty"`
	H2               string `json:"h2,omitempty"`
	SequentialWarmup string `json:"sequential-warmup,omitempty"`
	HttpsInsecure    string `json:"https-insecure,omitempty"`
	Resolve          string `json:"resolve,omitempty"`
	Timeout          string `json:"timeout,omitempty"`
	Header           string `json:"header,omitempty"`
}

type RunResponse struct {
	Message           string  `json:"message,omitempty"`
	Count             int     `json:"Count,omitempty"`
	ResultID          string  `json:"ResultID,omitempty"`
	ResultURL         string  `json:"ResultURL,omitempty"`
	RunType           string  `json:"RunType,omitempty"`
	Labels            string  `json:"Labels,omitempty"`
	StartTime         string  `json:"StartTime,omitempty"`
	RequestedQPS      string  `json:"RequestedQPS,omitempty"`
	RequestedDuration string  `json:"RequestedDuration,omitempty"`
	ActualQPS         float64 `json:"ActualQPS,omitempty"`
	ActualDuration    int     `json:"ActualDuration,omitempty"`
	NumThreads        int     `json:"NumThreads,omitempty"`
	Version           string  `json:"Version,omitempty"`
	DurationHistogram struct {
		Count  int     `json:"Count,omitempty"`
		Min    float64 `json:"Min,omitempty"`
		Max    float64 `json:"Max,omitempty"`
		Sum    float64 `json:"Sum,omitempty"`
		Avg    float64 `json:"Avg,omitempty"`
		StdDev float64 `json:"StdDev,omitempty"`
		Data   []struct {
			Start   float64 `json:"Start,omitempty"`
			End     float64 `json:"End,omitempty"`
			Percent int     `json:"Percent,omitempty"`
			Count   int     `json:"Count,omitempty"`
		} `json:"Data,omitempty"`
		Percentiles []struct {
			Percentile int     `json:"Percentile,omitempty"`
			Value      float64 `json:"Value,omitempty"`
		} `json:"Percentiles,omitempty"`
	} `json:"DurationHistogram,omitempty"`
	ErrorsDurationHistogram struct {
		Count  int         `json:"Count,omitempty"`
		Min    int         `json:"Min,omitempty"`
		Max    int         `json:"Max,omitempty"`
		Sum    int         `json:"Sum,omitempty"`
		Avg    int         `json:"Avg,omitempty"`
		StdDev int         `json:"StdDev,omitempty"`
		Data   interface{} `json:"Data,omitempty"`
	} `json:"ErrorsDurationHistogram,omitempty"`
	Exactly          int    `json:"Exactly,omitempty"`
	Jitter           bool   `json:"Jitter,omitempty"`
	Uniform          bool   `json:"Uniform,omitempty"`
	NoCatchUp        bool   `json:"NoCatchUp,omitempty"`
	RunID            int    `json:"RunID,omitempty"`
	AccessLoggerInfo string `json:"AccessLoggerInfo,omitempty"`
	ID               string `json:"ID,omitempty"`
	RetCodes         struct {
		Num200 int `json:"200,omitempty"`
	} `json:"RetCodes,omitempty"`
	IPCountMap struct {
		One270018080 int `json:"127.0.0.1:8080,omitempty"`
	} `json:"IPCountMap,omitempty"`
	Insecure          bool        `json:"Insecure,omitempty"`
	CACert            string      `json:"CACert,omitempty"`
	Cert              string      `json:"Cert,omitempty"`
	Key               string      `json:"Key,omitempty"`
	UnixDomainSocket  string      `json:"UnixDomainSocket,omitempty"`
	URL               string      `json:"URL,omitempty"`
	NumConnections    int         `json:"NumConnections,omitempty"`
	Compression       bool        `json:"Compression,omitempty"`
	DisableFastClient bool        `json:"DisableFastClient,omitempty"`
	HTTP10            bool        `json:"HTTP10,omitempty"`
	H2                bool        `json:"H2,omitempty"`
	DisableKeepAlive  bool        `json:"DisableKeepAlive,omitempty"`
	AllowHalfClose    bool        `json:"AllowHalfClose,omitempty"`
	FollowRedirects   bool        `json:"FollowRedirects,omitempty"`
	Resolve           string      `json:"Resolve,omitempty"`
	HTTPReqTimeOut    int64       `json:"HTTPReqTimeOut,omitempty"`
	UserCredentials   string      `json:"UserCredentials,omitempty"`
	ContentType       string      `json:"ContentType,omitempty"`
	Payload           interface{} `json:"Payload,omitempty"`
	LogErrors         bool        `json:"LogErrors,omitempty"`
	SequentialWarmup  bool        `json:"SequentialWarmup,omitempty"`
	ConnReuseRange    []int       `json:"ConnReuseRange,omitempty"`
	NoResolveEachConn bool        `json:"NoResolveEachConn,omitempty"`
	Offset            int         `json:"Offset,omitempty"`
	Resolution        float64     `json:"Resolution,omitempty"`
	Sizes             struct {
		Count  int `json:"Count,omitempty"`
		Min    int `json:"Min,omitempty"`
		Max    int `json:"Max,omitempty"`
		Sum    int `json:"Sum,omitempty"`
		Avg    int `json:"Avg,omitempty"`
		StdDev int `json:"StdDev,omitempty"`
		Data   []struct {
			Start   int `json:"Start,omitempty"`
			End     int `json:"End,omitempty"`
			Percent int `json:"Percent,omitempty"`
			Count   int `json:"Count,omitempty"`
		} `json:"Data,omitempty"`
	} `json:"Sizes,omitempty"`
	HeaderSizes struct {
		Count  int `json:"Count,omitempty"`
		Min    int `json:"Min,omitempty"`
		Max    int `json:"Max,omitempty"`
		Sum    int `json:"Sum,omitempty"`
		Avg    int `json:"Avg,omitempty"`
		StdDev int `json:"StdDev,omitempty"`
		Data   []struct {
			Start   int `json:"Start,omitempty"`
			End     int `json:"End,omitempty"`
			Percent int `json:"Percent,omitempty"`
			Count   int `json:"Count,omitempty"`
		} `json:"Data,omitempty"`
	} `json:"HeaderSizes,omitempty"`
	Sockets         []int `json:"Sockets,omitempty"`
	SocketCount     int   `json:"SocketCount,omitempty"`
	ConnectionStats struct {
		Count  int     `json:"Count,omitempty"`
		Min    float64 `json:"Min,omitempty"`
		Max    float64 `json:"Max,omitempty"`
		Sum    float64 `json:"Sum,omitempty"`
		Avg    float64 `json:"Avg,omitempty"`
		StdDev int     `json:"StdDev,omitempty"`
		Data   []struct {
			Start   float64 `json:"Start,omitempty"`
			End     float64 `json:"End,omitempty"`
			Percent int     `json:"Percent,omitempty"`
			Count   int     `json:"Count,omitempty"`
		} `json:"Data,omitempty"`
		Percentiles []struct {
			Percentile int     `json:"Percentile,omitempty"`
			Value      float64 `json:"Value,omitempty"`
		} `json:"Percentiles,omitempty"`
	} `json:"ConnectionStats,omitempty"`
	AbortOn int `json:"AbortOn,omitempty"`
}

func StartRun(run *FortioRest) *RunResponse{
	var result RunResponse
	jsonData, err := json.Marshal(run)
	if err != nil {
		log.Errorf("json marshal failed %s", err)
	}
	httpposturl := viper.GetString("fortio.url")  + fortio.RestRunURI + MetadataUri
	log.Debugf("sending run %s %s", httpposturl, jsonData)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("error building request %s", err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		log.Errorf("Failed to post to fortio %s", response)
	}
	defer response.Body.Close()

	log.Infof("response Status:", response.Status)
	body, _ := io.ReadAll(response.Body)


	if err := json.Unmarshal(body, &result); err != nil {   
		log.Errorf("Can not unmarshal JSON %s", err )
	}

	log.Infof("response Body: %s", string(body))
	return &result
}

