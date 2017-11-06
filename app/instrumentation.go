package main

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"fmt"
)


var requestCount metrics.Counter
var requestLatency metrics.Histogram

func init(){
	//make the counters and metrics
	fieldKeys := []string{"method", "error","release"}

	requestCount = kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Namespace: "runyontr",
		Subsystem: "appinfo_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency = kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Namespace: "runyontr",
		Subsystem: "appinfo_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
}


//instrumentationAppInfo implements the AppInfoService interface.  It provides metrics on calls to the Next service
type instrumentationAppInfo struct{

	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	Next AppInfoService
}



func NewInstrumentationAppInfoService(svc AppInfoService) AppInfoService{

	return &instrumentationAppInfo{
		Next: svc,
		requestCount: requestCount,
		requestLatency: requestLatency,
	}
}


//GetAppInfo returns the app info of the running application
func (s *instrumentationAppInfo) GetAppInfo() (info AppInfo, err error) {
	defer func(startTime time.Time){
		requestCount.With("release",info.Release, "method","GetAppInfo","error",fmt.Sprintf("%v",err)).Add(1)
		requestLatency.With("release",info.Release, "method","GetAppInfo","error",fmt.Sprintf("%v",err)).Observe(float64(time.Since(startTime)))
	}(time.Now())
	info, err = s.Next.GetAppInfo()
	return
}

