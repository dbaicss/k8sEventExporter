package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var namespace string = nameSpace

type Exporter struct {
	eventClient     EventClient
	lastScrapeError prometheus.Gauge
	scrapeDuration  prometheus.Gauge
	totalScrapes    prometheus.Counter
}

//定义指标类型
func NewExporter(eventClient EventClient) *Exporter {
	return &Exporter{
		eventClient: eventClient,
		lastScrapeError: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_scrape_error",
			Help:      "Was the last scrape of Kubernetes events successful",
		}),
		scrapeDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scrape_duration",
			Help:      "duration of the scrape",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total Kubernetes event scrapes",
		}),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	metricsCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricsCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricsCh)
	close(metricsCh)
	<-doneCh
	return
}
//event采集
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(ch)
	ch <- e.lastScrapeError
	ch <- e.scrapeDuration
	ch <- e.totalScrapes
	return
}

//聚合统计
func (e *Exporter) scrape(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	var err error
	defer func(startTime time.Time) {
		e.scrapeDuration.Set(time.Since(startTime).Seconds())
		if err == nil {
			e.lastScrapeError.Set(0)
		} else {
			e.lastScrapeError.Set(1)
		}
	}(time.Now())

	if err = e.eventClient.Scrape(ch); err != nil {
		log.Errorf("Error scraping events: %s", err)
	}
}
