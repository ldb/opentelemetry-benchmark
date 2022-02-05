package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// Tool `promdl` automatically downloads and saves as CSV some predefined queries from Prometheus.

var (
	flagHost  = flag.String("host", "", "Prometheus host URL")
	startTime = flag.Duration("start", 1*time.Hour, "how much time to go back to start fetching fetch results from")
	endTime   = flag.Duration("end", 30*time.Minute, "how much time to go back to end fetching fetch results from")
)

// A list of queries we want to export from Prometheus
var queries = map[string]string{
	"benchd_error_rate":      "sum(rate(benchd_manager_worker_error_count[1m]))",
	"benchd_receive_rate":    "sum(rate(benchd_manager_traces_received_count[1m]))",
	"benchd_send_rate":       "sum(rate(benchd_manager_traces_sent_count[1m]))",
	"benchd_active_workers":  "benchd_manager_active_workers_count",
	"otel_cpu_busy":          "(((count(count(node_cpu_seconds_total{instance='otel-collector',job='node'}) by (cpu))) - avg(sum by (mode)(rate(node_cpu_seconds_total{mode='idle',instance='otel-collector',job='node'}[1m])))) * 100) / count(count(node_cpu_seconds_total{instance='otel-collector',job='node'}) by (cpu))",
	"otel_memory_used":       "((node_memory_MemTotal_bytes{instance='otel-collector',job='node'} - node_memory_MemFree_bytes{instance='otel-collector',job='node'}) / (node_memory_MemTotal_bytes{instance='otel-collector',job='node'} )) * 100",
	"otel_accepted_spans":    "sum(rate(otelcol_receiver_accepted_spans{}[1m])) by (receiver)",
	"otel_refused_spans":     "sum(rate(otelcol_receiver_refused_spans{}[1m])) by (receiver)",
	"otel_sent_spans":        "sum(rate(otelcol_exporter_sent_spans{}[1m])) by (exporter)",
	"otel_failed_spans":      "sum(rate(otelcol_exporter_send_failed_spans{}[1m])) by (exporter)",
	"otel_cpu_system":        "sum by (instance)(rate(node_cpu_seconds_total{mode='system',instance='otel-collector',job='node'}[1m])) * 100",
	"otel_cpu_user":          "sum by (instance)(rate(node_cpu_seconds_total{mode='user',instance='otel-collector',job='node'}[1m])) * 100",
	"otel_cpu_iowait":        "sum by (instance)(rate(node_cpu_seconds_total{mode='iowait',instance='otel-collector',job='node'}[1m])) * 100",
	"otel_received_bytes":    "rate(node_network_receive_bytes_total{instance='otel-collector',job='node'}[1m])*8",
	"otel_transmitted_bytes": "rate(node_network_transmit_bytes_total{instance='otel-collector',job='node'}[1m])*8",
}

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		flag.PrintDefaults()
		return
	}
	for name, query := range queries {
		log.Printf("running query %s", name)

		wg := sync.WaitGroup{}
		start := time.Now().Add(-*startTime)
		end := time.Now().Add(-*endTime)
		url := generateURL(*flagHost, query, start, end)
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			runQuery(url, name)
			wg.Done()
		}(&wg)

		wg.Wait()
	}
}

func generateURL(baseURL, query string, startTime, endTime time.Time) string {
	u, err := url.Parse(baseURL + "/api/v1/query_range")
	if err != nil {
		log.Printf("error parsing query: %v", err)
		return ""
	}
	q := u.Query()
	q.Add("query", query)
	q.Add("start", startTime.Format(time.RFC3339Nano))
	q.Add("end", endTime.Format(time.RFC3339Nano))
	q.Add("step", (endTime.Sub(startTime) / 120).Round(time.Second).String())
	u.RawQuery = q.Encode()

	return u.String()
}

func runQuery(url, queryName string) {
	if url == "" {
		log.Printf("invalid url %s", url)
	}
	res, err := getResults(url)
	if err != nil {
		log.Fatalf("error getting results: %v", err)
	}
	f, err := os.Create(queryName + ".csv")
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	w := csv.NewWriter(f)
	w.Write([]string{"time", "value"})
	if err := w.WriteAll(res); err != nil {
		log.Fatalf("error writing csv: %v", err)
	}
}

type results [][]string
type response struct {
	Data struct {
		Result []struct {
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func getResults(url string) (results, error) {
	res, err := http.Get(url)
	if err != nil {
		return results{}, fmt.Errorf("erorr performing request: %v", err)
	}
	defer res.Body.Close()
	results := results{}

	d := json.NewDecoder(res.Body)
	resp := response{}
	err = d.Decode(&resp)
	if err != nil {
		log.Fatalf("error decoding response: %v", err)
	}

	for _, values := range resp.Data.Result {
		for _, val := range values.Values {
			t := val[0].(float64)
			v := val[1].(string)
			results = append(results, []string{fmt.Sprintf("%.0f", t), v})
		}
	}
	return results, nil
}
