/**
 * Copyright 2020 Centreon Team
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package collector

import (
	"bufio"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter prometheus
type Exporter struct {
	URL     *string
	mutex   sync.RWMutex
	metrics map[string]*prometheus.Desc
	logger  log.Logger
}

// Describe describes all the metrics ever exported by the SensisionExporter exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		ch <- m
	}
}

// Collect fetches the stats from configured SensisionExporter location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	client := &http.Client{}
	request, err := http.NewRequest("GET", *e.URL, nil)
	if err != nil {
		level.Error(e.logger).Log("msg", "Cannot create the request to get sensision metrics", "err", err.Error())
		return
	}
	response, err := client.Do(request)
	if err != nil {
		level.Error(e.logger).Log("msg", "Error getting metrics from sensision", "err", err.Error())
		return
	}
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		parseWarp10(scanner.Text(), ch, e.logger)
	}
}

// NewSensisionExporter create a sensision exporter for prometheus
func NewSensisionExporter(sensisionURL string, logger log.Logger) (*Exporter, error) {
	if _, err := url.Parse(sensisionURL); err != nil {
		return nil, err
	}

	return &Exporter{
		URL:     &sensisionURL,
		metrics: metrics,
		logger:  logger,
	}, nil
}

func labelIndexOf(slice []string, label string) (int, error) {
	for idx, value := range slice {
		if value == label {
			return idx, nil
		}
	}
	return 0, errors.New("Not found")
}

func parseWarp10(lineMetric string, ch chan<- prometheus.Metric, logger log.Logger) {
	matcherWarp10 := regexp.MustCompile("^([0-9]+)/[^/]*/[^/]*[ 	]+([^{]+){([^}]*)}[ 	](.*)$")
	matches := matcherWarp10.FindStringSubmatch(lineMetric)

	if len(matches) == 5 {
		if metric, ok := metrics[strings.ReplaceAll(matches[2], ".", "_")]; ok {
			if value, err := strconv.ParseFloat(matches[4], 64); err == nil {
				// Prepare labels
				var labelValues []string
				// Test if the prometheus metrics has labels
				if labelsList, ok := labels[strings.ReplaceAll(matches[2], ".", "_")]; ok {
					labelValues = make([]string, len(labelsList))
					for _, label := range strings.Split(matches[3], ",") {
						labelInfo := strings.Split(label, "=")
						if len(labelInfo) == 2 {
							labelValue, _ := url.QueryUnescape(labelInfo[1])
							if pos, err := labelIndexOf(labelsList, labelInfo[0]); err == nil {
								labelValues[pos] = labelValue
							} else {
								level.Info(logger).Log("msg", "Missing label", "label", labelInfo[0], "metric", matches[2])
							}
						}
					}
				} else {
					labelValues = make([]string, 0)
				}

				// Write the metrics
				ch <- prometheus.MustNewConstMetric(
					metric,
					prometheus.GaugeValue,
					value,
					labelValues...,
				)
			}
		}
	}
}
