package collector

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
)

//Filebeat json structure
type Filebeat struct {
	Events struct {
		Active float64 `json:"active"`
		Added  float64 `json:"added"`
		Done   float64 `json:"done"`
	} `json:"events"`

	Harvester struct {
		Closed    float64 `json:"closed"`
		OpenFiles float64 `json:"open_files"`
		Running   float64 `json:"running"`
		Skipped   float64 `json:"skipped"`
		Started   float64 `json:"started"`
	} `json:"harvester"`

	Input struct {
		Log struct {
			Files struct {
				Renamed   float64 `json:"renamed"`
				Truncated float64 `json:"truncated"`
			} `json:"files"`
		} `json:"log"`
	} `json:"input"`
}

type filebeatCollector struct {
	beatInfo *BeatInfo
	stats    *Stats
	metrics  exportedMetrics
}

var lastError = ""
var errorCount = float64(0)

func isHarvesterError(r *regexp.Regexp, l string) string {
	if r.MatchString(l) {
		return strings.Split(l, "\t")[0]
	}

	return ""
}

func getHarvesterErrors(stats *Stats) float64 {
	var err error

	//fmt.Println("getHarvesterErrors CALL")

	cmd := exec.Command("docker",
		"logs",
		"filebeat",
		"--tail",
		"100",
		"--since",
		lastError,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{}

	var o bytes.Buffer
	var e bytes.Buffer

	cmd.Stdout = &o
	cmd.Stderr = &e

	r := regexp.MustCompile("(ERROR).*?(harvester).*?(Read line error)")

	err = cmd.Run()
	if err != nil {
		fmt.Printf("getHarvesterErrors ERROR [%v, %s]\n", err, e.String())
	} else {
		errStr := strings.Split(e.String(), "\n")

		if len(lastError) > 0 {
			errStr = errStr[1:]
		}

		for _, v := range errStr {
			checkResult := isHarvesterError(r, v)

			if len(checkResult) > 0 {
				lastError = checkResult

				errorCount = errorCount + 1
			}
		}
	}

	return errorCount
}

// NewFilebeatCollector constructor
func NewFilebeatCollector(beatInfo *BeatInfo, stats *Stats) prometheus.Collector {
	return &filebeatCollector{
		beatInfo: beatInfo,
		stats:    stats,
		metrics: exportedMetrics{
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "events"),
					"filebeat.events",
					nil, prometheus.Labels{"event": "active"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Events.Active },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "events"),
					"filebeat.events",
					nil, prometheus.Labels{"event": "added"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Events.Added },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "events"),
					"filebeat.events",
					nil, prometheus.Labels{"event": "done"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Events.Done },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "closed"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Harvester.Closed },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "open_files"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Harvester.OpenFiles },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "running"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Harvester.Running },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "skipped"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Harvester.Skipped },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "started"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Harvester.Started },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "harvester"),
					"filebeat.harvester",
					nil, prometheus.Labels{"harvester": "errors"},
				),
				eval:    getHarvesterErrors,
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "input_log"),
					"filebeat.input_log",
					nil, prometheus.Labels{"files": "renamed"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Input.Log.Files.Renamed },
				valType: prometheus.UntypedValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(beatInfo.Beat, "filebeat", "input_log"),
					"filebeat.input_log",
					nil, prometheus.Labels{"files": "truncated"},
				),
				eval:    func(stats *Stats) float64 { return stats.Filebeat.Input.Log.Files.Truncated },
				valType: prometheus.UntypedValue,
			},
		},
	}
}

// Describe returns all descriptions of the collector.
func (c *filebeatCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, metric := range c.metrics {
		ch <- metric.desc
	}

}

// Collect returns the current state of all metrics of the collector.
func (c *filebeatCollector) Collect(ch chan<- prometheus.Metric) {

	for _, i := range c.metrics {
		ch <- prometheus.MustNewConstMetric(i.desc, i.valType, i.eval(c.stats))
	}

}
