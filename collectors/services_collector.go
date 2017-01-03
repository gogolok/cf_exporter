package collectors

import (
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ServicesCollector struct {
	namespace                               string
	cfClient                                *cfclient.Client
	serviceInfoMetric                       *prometheus.GaugeVec
	servicesTotalMetric                     prometheus.Gauge
	servicesScrapesTotalMetric              prometheus.Counter
	servicesScrapeErrorsTotalMetric         prometheus.Counter
	lastServicesScrapeErrorMetric           prometheus.Gauge
	lastServicesScrapeTimestampMetric       prometheus.Gauge
	lastServicesScrapeDurationSecondsMetric prometheus.Gauge
}

func NewServicesCollector(namespace string, cfClient *cfclient.Client) *ServicesCollector {
	serviceInfoMetric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "service",
			Name:      "info",
			Help:      "Labeled Cloud Foundry Service information with a constant '1' value.",
		},
		[]string{"service_id", "service_label"},
	)

	servicesTotalMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "services",
			Name:      "total",
			Help:      "Total number of Cloud Foundry Services.",
		},
	)

	servicesScrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "services_scrapes",
			Name:      "total",
			Help:      "Total number of scrapes for Cloud Foundry Services.",
		},
	)

	servicesScrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "services_scrape_errors",
			Name:      "total",
			Help:      "Total number of scrape error of Cloud Foundry Services.",
		},
	)

	lastServicesScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_error",
			Help:      "Whether the last scrape of Services metrics from Cloud Foundry resulted in an error (1 for error, 0 for success).",
		},
	)

	lastServicesScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of Services metrics from Cloud Foundry.",
		},
	)

	lastServicesScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "",
			Name:      "last_services_scrape_duration_seconds",
			Help:      "Duration of the last scrape of Services metrics from Cloud Foundry.",
		},
	)

	return &ServicesCollector{
		namespace:                               namespace,
		cfClient:                                cfClient,
		serviceInfoMetric:                       serviceInfoMetric,
		servicesTotalMetric:                     servicesTotalMetric,
		servicesScrapesTotalMetric:              servicesScrapesTotalMetric,
		servicesScrapeErrorsTotalMetric:         servicesScrapeErrorsTotalMetric,
		lastServicesScrapeErrorMetric:           lastServicesScrapeErrorMetric,
		lastServicesScrapeTimestampMetric:       lastServicesScrapeTimestampMetric,
		lastServicesScrapeDurationSecondsMetric: lastServicesScrapeDurationSecondsMetric,
	}
}

func (c ServicesCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	errorMetric := float64(0)
	if err := c.reportServicesMetrics(ch); err != nil {
		errorMetric = float64(1)
		c.servicesScrapeErrorsTotalMetric.Inc()
	}

	c.servicesScrapesTotalMetric.Inc()
	c.servicesScrapesTotalMetric.Collect(ch)

	c.servicesScrapeErrorsTotalMetric.Collect(ch)

	c.lastServicesScrapeErrorMetric.Set(errorMetric)
	c.lastServicesScrapeErrorMetric.Collect(ch)

	c.lastServicesScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastServicesScrapeTimestampMetric.Collect(ch)

	c.lastServicesScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastServicesScrapeDurationSecondsMetric.Collect(ch)
}

func (c ServicesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.serviceInfoMetric.Describe(ch)
	c.servicesTotalMetric.Describe(ch)
	c.servicesScrapesTotalMetric.Describe(ch)
	c.servicesScrapeErrorsTotalMetric.Describe(ch)
	c.lastServicesScrapeErrorMetric.Describe(ch)
	c.lastServicesScrapeTimestampMetric.Describe(ch)
	c.lastServicesScrapeDurationSecondsMetric.Describe(ch)
}

func (c ServicesCollector) reportServicesMetrics(ch chan<- prometheus.Metric) error {
	c.serviceInfoMetric.Reset()

	services, err := c.cfClient.ListServices()
	if err != nil {
		log.Errorf("Error while listing services: %v", err)
		return err
	}

	for _, service := range services {
		c.serviceInfoMetric.WithLabelValues(
			service.Guid,
			service.Label,
		).Set(float64(1))
	}

	c.serviceInfoMetric.Collect(ch)

	c.servicesTotalMetric.Set(float64(len(services)))
	c.servicesTotalMetric.Collect(ch)

	return nil
}
