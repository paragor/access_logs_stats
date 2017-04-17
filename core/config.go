package core

import (
	"encoding/json"
	"fmt"
	"github.com/blackbass1988/access_logs_stats/core/re"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//Config base struct of parser config
type Config struct {
	InputDsn string

	ExitAfterOneTick bool

	Counts     map[string]bool
	Aggregates map[string]bool

	Outputs []*outputConfig
	Rex     re.RegExp
	Period  time.Duration
	Filters []*Filter
}

type outputConfig struct {
	Type     string            `json:"type" yaml:"type"`
	Settings map[string]string `json:"settings" yaml:"settings"`
}

type configStruct struct {
	InputDsn   string   `json:"input" yaml:"input"`
	Regexp     string   `json:"regexp" yaml:"regexp"`
	Period     string   `json:"period" yaml:"period"`
	Counts     []string `json:"counts" yaml:"counts"`
	Aggregates []string `json:"aggregates" yaml:"aggregates"`

	Filters []*Filter       `json:"filters" yaml:"filters"`
	Outputs []*outputConfig `json:"output" yaml:"output"`
}

//NewConfig parse config filepath and return new Config
func NewConfig(filepath string) (config Config, err error) {
	configStruct := new(configStruct)
	config.Aggregates = make(map[string]bool)
	config.Counts = make(map[string]bool)

	//we need to lock file in processlist for restore by file descriptor if delete in runtime
	_, err = os.Open(filepath)

	if err != nil {
		return config, err
	}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	//filename can doesn't have "yaml" substring. dirty hack. === in start check
	if strings.Contains(filepath, ".yaml") || bytes[0] == 45 && bytes[1] == 45 && bytes[2] == 45 {
		err = yaml.Unmarshal(bytes, &configStruct)
	} else {
		err = json.Unmarshal(bytes, &configStruct)
	}

	if err != nil {
		return config, err
	}

	config.InputDsn = configStruct.InputDsn
	config.Period, err = time.ParseDuration(configStruct.Period)
	if err != nil {
		return config, err
	}

	config.Rex, err = re.Compile(configStruct.Regexp)
	if err != nil {
		return config, err
	}

	config.Outputs = configStruct.Outputs

	for _, el := range configStruct.Counts {
		config.Counts[el] = true
	}

	for _, el := range configStruct.Aggregates {
		config.Aggregates[el] = true
	}

	for _, f := range configStruct.Filters {

		for _, filterItem := range f.Items {
			for _, metric := range filterItem.Metrics {

				switch {
				case metric == "min", metric == "max", metric == "len", metric == "avg",
					metric == "sum", metric == "sum_ps", metric == "ips", strings.Contains(metric, "cent_"):

					if !config.Aggregates[filterItem.Field] {
						err = fmt.Errorf("field \"%s\" must in in \"aggregates\" section"+
							" because you want metric \"%s\"",
							filterItem.Field, metric)
					}
				case metric == "uniq", metric == "uniq_ps", strings.Contains(metric, "cps_"), strings.Contains(metric, "percentage_"):
					if !config.Counts[filterItem.Field] {
						err = fmt.Errorf("field \"%s\" must in in \"counts\" section "+
							"because you want metric \"%s\"",
							filterItem.Field, metric)
					}
				}

			}
		}

		check(err)
		config.Filters = append(config.Filters, f)
	}

	if len(config.Filters) == 0 {
		err = errFiltersNotSet
	}

	if len(config.Outputs) == 0 {
		return config, errOutputNotSet
	}

	return config, err
}
