package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ldb/openetelemtry-benchmark/benchmark"
	"github.com/ldb/openetelemtry-benchmark/config"
	"net/http"
)

var ErrInvalidName = errors.New("invalid Benchmark name")

type Client struct {
	host   string
	client *http.Client
}

func NewClient(host string) Client {
	return Client{
		host:   host,
		client: new(http.Client),
	}
}

// CreateBenchmark creates a new benchmark with name `name`.
func (c *Client) CreateBenchmark(name string) (benchmark.Status, error) {
	if name == "" {
		return benchmark.Status{}, ErrInvalidName
	}
	url := c.host + "/create/" + name
	r, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error creating request: %v", err)
	}
	res, err := c.client.Do(r)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error performing request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return benchmark.Status{}, fmt.Errorf("error creating Benchmark with name %s: %v", name, err)
	}
	status := &benchmark.Status{}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(status); err != nil {
		return benchmark.Status{}, fmt.Errorf("error decoding body: %v", err)
	}
	return *status, nil
}

// ConfigureBenchmark configures benchmark `name` using the config.BenchConfig `config`.
func (c *Client) ConfigureBenchmark(name string, config config.BenchConfig) (benchmark.Status, error) {
	if name == "" {
		return benchmark.Status{}, ErrInvalidName
	}
	url := c.host + "/configure/" + name
	bb, err := json.Marshal(config)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error encoding benchmark configuration: %v", err)
	}
	r := bytes.NewReader(bb)
	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error creating request: %v", err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error performing request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return benchmark.Status{}, fmt.Errorf("error configuring Benchmark with name %s: %v", name, err)
	}
	status := &benchmark.Status{}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(status); err != nil {
		return benchmark.Status{}, fmt.Errorf("error decoding body: %v", err)
	}
	return *status, nil

}

func (c *Client) StartBenchmark(name string) (benchmark.Status, error) {
	if name == "" {
		return benchmark.Status{}, ErrInvalidName
	}
	url := c.host + "/start/" + name
	r, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error creating request: %v", err)
	}
	res, err := c.client.Do(r)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error performing request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return benchmark.Status{}, fmt.Errorf("error starting Benchmark with name %s: %v", name, err)
	}
	status := &benchmark.Status{}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(status); err != nil {
		return benchmark.Status{}, fmt.Errorf("error decoding body: %v", err)
	}
	return *status, nil
}

func (c *Client) StopBenchmark(name string) (benchmark.Status, error) {
	if name == "" {
		return benchmark.Status{}, ErrInvalidName
	}
	url := c.host + "/stop/" + name
	r, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error creating request: %v", err)
	}
	res, err := c.client.Do(r)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error performing request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return benchmark.Status{}, fmt.Errorf("error stopping Benchmark with name %s: %v", name, err)
	}
	status := &benchmark.Status{}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(status); err != nil {
		return benchmark.Status{}, fmt.Errorf("error decoding body: %v", err)
	}
	return *status, nil
}

func (c *Client) Status(name string) (benchmark.Status, error) {
	if name == "" {
		return benchmark.Status{}, ErrInvalidName
	}
	url := c.host + "/status/" + name
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error creating request: %v", err)
	}
	res, err := c.client.Do(r)
	if err != nil {
		return benchmark.Status{}, fmt.Errorf("error performing request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return benchmark.Status{}, fmt.Errorf("error getting status for Benchmark with name %s: %v", name, err)
	}
	status := &benchmark.Status{}
	d := json.NewDecoder(res.Body)
	if err := d.Decode(status); err != nil {
		return benchmark.Status{}, fmt.Errorf("error decoding body: %v", err)
	}
	return *status, nil
}
