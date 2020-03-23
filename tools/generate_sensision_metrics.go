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
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const sensisionMetricsURL = "https://raw.githubusercontent.com/senx/warp10-platform/@version@/warp10/src/main/java/io/warp10/continuum/sensision/SensisionConstants.java"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing version")
		os.Exit(1)
	}
	// Get version to parse
	var version string
	version = os.Args[1]
	urlData := strings.ReplaceAll(sensisionMetricsURL, "@version@", version)

	// Prepare request to get file
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodGet, urlData, nil)
	if err != nil {
		fmt.Println("Error preparing request to download file")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error downloading file")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close()

	// Load template
	sensisionTemplate, err := ioutil.ReadFile("sensision.tpl")
	if err != nil {
		fmt.Println("Cannot open the template file")
		os.Exit(1)
	}
	sensisionTemplateString := string(sensisionTemplate)

	// Open file to write
	os.Remove("sensision.go")
	metricFile, err := os.OpenFile("sensision.go", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Cannot open the file to write")
		os.Exit(1)
	}
	defer metricFile.Close()

	// Static label
	labelsByMetric := map[string]string{
		"SENSISION_CLASS_WARPSCRIPT_RUN_COUNT":                                "path",
		"SENSISION_CLASS_WARPSCRIPT_RUN_FAILURES":                             "path",
		"SENSISION_CLASS_WARPSCRIPT_RUN_TIME_US":                              "path",
		"SENSISION_CLASS_WARPSCRIPT_RUN_ELAPSED":                              "path",
		"SENSISION_CLASS_WARPSCRIPT_RUN_FETCHED":                              "path",
		"SENSISION_CLASS_WARPSCRIPT_RUN_OPS":                                  "path",
		"SENSISION_CLASS_CONTINUUM_FETCH_BYTES_VALUES_PEROWNER":               "consumer,app,owner,consumerapp",
		"SENSISION_CLASS_CONTINUUM_FETCH_BYTES_KEYS_PEROWNER":                 "consumer,app,owner,consumerapp",
		"SENSISION_CLASS_CONTINUUM_FETCH_DATAPOINTS_PEROWNER":                 "consumer,app,owner,consumerapp",
		"SENSISION_CLASS_CONTINUUM_FETCH_BYTES_VALUES":                        "consumer,app,consumerapp",
		"SENSISION_CLASS_CONTINUUM_FETCH_BYTES_KEYS":                          "consumer,app,consumerapp",
		"SENSISION_CLASS_CONTINUUM_FETCH_DATAPOINTS":                          "consumer,app,consumerapp",
		"SENSISION_CLASS_CONTINUUM_STORE_HBASE_DELETE_DATAPOINTS_PEROWNERAPP": "owner,app",
		"SENSISION_CLASS_CONTINUUM_DIRECTORY_GTS_PERAPP":                      "app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_UPDATE_GZIPPED":                    "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_UPDATE_PARSEERRORS":                "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_UPDATE_DATAPOINTS_RAW":             "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_UPDATE_TIME_US":                    "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_META_GZIPPED":                      "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_META_INVALID":                      "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_META_RECORDS":                      "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_DELETE_REQUESTS":                   "producer,app",
		"SENSISION_CLASS_CONTINUUM_INGRESS_DELETE_GTS":                        "producer,app",
		"SENSISION_CLASS_CONTINUUM_THROTTLING_GTS":                            "producer",
		"SENSISION_CLASS_CONTINUUM_GTS_DISTINCT":                              "producer",
		"SENSISION_CLASS_CONTINUUM_THROTTLING_GTS_PER_APP":                    "app",
		"SENSISION_CLASS_CONTINUUM_GTS_DISTINCT_PER_APP":                      "app",
		"SENSISION_CLASS_CONTINUUM_THROTTLING_RATE_PER_APP":                   "app",
		"SENSISION_CLASS_CONTINUUM_THROTTLING_RATE":                           "producer",
		"SENSISION_CLASS_CONTINUUM_ESTIMATOR_RESETS":                          "producer",
		"SENSISION_CLASS_CONTINUUM_ESTIMATOR_RESETS_PER_APP":                  "app",
		"SENSISION_CLASS_QUASAR_FILTER_TOKEN_COUNT":                           "type,error",
		"SENSISION_CLASS_QUASAR_FILTER_TOKEN_TIME_US":                         "type,error",
		"CLASS_WARP_DATALOG_FORWARDER_REQUESTS_FORWARDED":                     "forwarder,id,type",
		"CLASS_WARP_DATALOG_FORWARDER_REQUESTS_FAILED":                        "forwarder,id,type",
		"CLASS_WARP_DATALOG_FORWARDER_REQUESTS_IGNORED":                       "forwarder,id,type",
		"CLASS_WARP_DATALOG_REQUESTS_RECEIVED":                                "id,type",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_UPDATE_DATAPOINTS_RAW":          "producer,cdn",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_UPDATE_PARSEERRORS":             "producer",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_UPDATE_REQUESTS":                "producer,cdn",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_UPDATE_TIME_US":                 "producer,cdn",
		"CLASS_WARP_DATALOG_REQUESTS_LOGGED":                                  "id,type",
		"SENSISION_CLASS_WARP_KAFKA_CONSUMER_OFFSET_FORWARD_LEAPS":            "topic,groupid,partition",
		"SENSISION_CLASS_WARP_KAFKA_CONSUMER_OFFSET_BACKWARD_LEAPS":           "topic,groupid,partition",
		"SENSISION_CLASS_PLASMA_BACKEND_SUBSCRIPTIONS_INVALID_HASHES":         "topic",
		"SENSISION_CLASS_CONTINUUM_FETCH_COUNT":                               "app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_DELETE_DATAPOINTS_PEROWNERAPP":  "owner,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_DELETE_REQUESTS":                "producer,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_DELETE_GTS":                     "producer,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_DELETE_DATAPOINTS":              "producer,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_DELETE_TIME_US":                 "producer,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_STREAM_UPDATE_DATAPOINTS_RAW":   "producer",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_STREAM_UPDATE_MESSAGES":         "producer",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_STREAM_UPDATE_TIME_US":          "producer,app",
		"SENSISION_CLASS_CONTINUUM_STANDALONE_STREAM_UPDATE_REQUESTS":         "producer,app",
		"SENSISION_CLASS_CONTINUUM_STREAM_UPDATE_REQUESTS":                    "producer,app",
		"SENSISION_CLASS_CONTINUUM_STREAM_UPDATE_PARSEERRORS":                 "producer,app",
		"SENSISION_CLASS_CONTINUUM_STREAM_UPDATE_DATAPOINTS_RAW":              "producer,app",
		"SENSISION_CLASS_CONTINUUM_STREAM_UPDATE_MESSAGES":                    "producer,app",
		"SENSISION_CLASS_CONTINUUM_STREAM_UPDATE_TIME_US":                     "producer,app",
		"SENSISION_CLASS_CONTINUUM_FETCH_REQUESTS":                            "type",
		"SENSISION_CLASS_CONTINUUM_SFETCH_WRAPPERS_PERAPP":                    "app",
		"SENSISION_CLASS_CONTINUUM_SFETCH_WRAPPERS_SIZE_PERAPP":               "app",
		"SENSISION_CLASS_CONTINUUM_SFETCH_WRAPPERS_DATAPOINTS_PERAPP":         "app",
		"SENSISION_CLASS_WARPSCRIPT_FUNCTION_COUNT":                           "function",
		"SENSISION_CLASS_WARPSCRIPT_FUNCTION_TIME_US":                         "function",
		"SENSISION_CLASS_WARPSCRIPT_FETCHCOUNT_EXCEEDED":                      "consumer",
	}

	// Parsing file
	scanner := bufio.NewScanner(response.Body)
	var description string
	inDoc := false
	nextMetric := false
	metricsText := ""
	metricMatcher := regexp.MustCompile("^public static final String ([A-Z_]+) = \"([^\"]+)\";$")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "/**" {
			inDoc = true
			nextMetric = false
			description = ""
		} else if line == "*/" {
			inDoc = false
		} else {
			if inDoc {
				if strings.Index(line, "* Number") == 0 || strings.Index(line, "* Total") == 0 || strings.Index(line, "* MADS") == 0 || strings.Index(line, "* Current") == 0 {
					nextMetric = true
					description = strings.ReplaceAll(line, "* ", "")
				}
			} else if nextMetric {
				matches := metricMatcher.FindStringSubmatch(line)
				if len(matches) > 0 {
					appendLabels := "\n"
					metricsText += fmt.Sprintf("\tmetrics[\"%s\"] = prometheus.NewDesc(\n", strings.ReplaceAll(matches[2], ".", "_"))
					metricsText += fmt.Sprintf(
						"\t\tprometheus.BuildFQName(namespace, \"\", \"%s\"),\n",
						strings.ReplaceAll(matches[2], ".", "_"),
					)
					metricsText += fmt.Sprintf("\t\t\"%s\",\n", description)
					if labels, ok := labelsByMetric[matches[1]]; ok {
						appendLabels = fmt.Sprintf("\tlabels[\"%s\"] = []string{\"%s\"}\n\n", strings.ReplaceAll(matches[2], ".", "_"), strings.Join(strings.Split(labels, ","), "\",\""))
						metricsText += fmt.Sprintf("\t\t[]string{\"%s\"},\n", strings.Join(strings.Split(labels, ","), "\",\""))
					} else {
						metricsText += "\t\tmake([]string, 0),\n"
					}
					metricsText += "\t\tprometheus.Labels{},\n"
					metricsText += "\t)\n"
					metricsText += appendLabels
					nextMetric = false
					description = ""
				}
			}
		}
	}

	metricFile.WriteString(fmt.Sprintf(sensisionTemplateString, metricsText))
}
