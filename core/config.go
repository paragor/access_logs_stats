package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

type OutputConfig struct {
	Type     string            `json:"type"`
	Settings map[string]string `json:"settings"`
}

type configJson struct {
	InputDsn   string   `json:"input"`
	Regexp     string   `json:"regexp"`
	Period     string   `json:"period"`
	Counts     []string `json:"counts"`
	Aggregates []string `json:"aggregates"`

	Filters []*Filter       `json:"filters"`
	Outputs []*OutputConfig `json:"output"`
}

type Config struct {
	InputDsn string

	Counts     map[string]bool
	Aggregates map[string]bool

	Outputs []*OutputConfig
	Rex     *regexp.Regexp
	Period  time.Duration
	Filters []*Filter
}

type Filter struct {
	Filter string `json:"filter"`
	Prefix string `json:"prefix"`
	Items  []struct {
		Field   string   `json:"field"`
		Metrics []string `json:"metrics"`
	} `json:"items"`

	FilterRex *regexp.Regexp
}

func NewConfig(filepath string) (config Config, err error) {
	configJson := new(configJson)
	config.Aggregates = make(map[string]bool)
	config.Counts = make(map[string]bool)

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &configJson)
	if err != nil {
		return config, err
	}

	config.InputDsn = configJson.InputDsn
	config.Period, err = time.ParseDuration(configJson.Period)
	if err != nil {
		return config, err
	}

	config.Rex, err = regexp.Compile(configJson.Regexp)
	if err != nil {
		return config, err
	}

	config.Outputs = configJson.Outputs

	for _, el := range configJson.Counts {
		config.Counts[el] = true
	}

	for _, el := range configJson.Aggregates {
		config.Aggregates[el] = true
	}

	for _, f := range configJson.Filters {
		f.FilterRex, err = regexp.Compile(f.Filter)

		for _, filterItem := range f.Items {
			for _, metric := range filterItem.Metrics {
				//если метрика содержит cps_, значит она должна быть в counts
				//иначе должна быть aggregates

				if strings.Contains(metric, "cps_") && !config.Counts[filterItem.Field] {
					err = errors.New(
						fmt.Sprintf("field \"%s\" must in in \"counts\" section "+
							"because you want metric \"%s\"",
							filterItem.Field, metric))
				} else if !strings.Contains(metric, "cps_") && !config.Aggregates[filterItem.Field] {
					err = errors.New(
						fmt.Sprintf("field \"%s\" must in in \"aggregates\" section"+
							" because you want metric \"%s\"",
							filterItem.Field, metric))
				}
			}
		}

		check(err)
		config.Filters = append(config.Filters, f)
	}

	if len(config.Filters) == 0 {
		err = ERR_FILTERS_NOT_SET
	}

	if len(config.Outputs) == 0 {
		return config, ERR_OUTPUT_NOT_SET
	}

	return config, err
}