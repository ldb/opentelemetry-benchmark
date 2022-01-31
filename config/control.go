package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ControlConfig struct {
	Target     string
	Clients    []string
	Monitoring string
}

func NewFrom(reader io.Reader) (ControlConfig, error) {
	s := bufio.NewScanner(reader)
	c := ControlConfig{Clients: []string{}}
	for s.Scan() {
		t := s.Text()
		// Ignore lines that start with `#`. Those can be used for comments.
		if strings.HasPrefix(t, "#") {
			continue
		}
		ss := strings.Split(t, " ")
		if len(ss) != 2 {
			return c, errors.New("malformed config")
		}
		switch ss[0] {
		case "target":
			c.Target = ss[1]
		case "client":
			c.Clients = append(c.Clients, ss[1])
		case "monitoring":
			c.Monitoring = ss[1]
		default:
			return c, fmt.Errorf("unknown token %q", ss[0])
		}
	}
	if err := s.Err(); err != nil {
		return c, fmt.Errorf("error scanning config: %v", err)
	}
	return c, nil
}
