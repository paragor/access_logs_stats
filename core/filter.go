package core

import (
	"github.com/blackbass1988/access_logs_stats/core/re"
	"log"
	"strings"
)

var regularExpressionRex = re.MustCompile(`[\[\]{}+*\\()]`)

//Filter matching input string
type Filter struct {
	Matcher *matcher `json:"filter" yaml:"filter"`
	Prefix  string   `json:"prefix" yaml:"prefix"`
	Items   []struct {
		Field   string   `json:"field" yaml:"field"`
		Metrics []string `json:"metrics" yaml:"metrics"`
	} `json:"items" yaml:"items"`
}

//MatchString matches a input string
func (f *Filter) MatchString(str string) bool {
	return f.Matcher.MatchString(str)
}

func (f *Filter) String() string {
	return f.Matcher.String()
}

type matcher struct {
	raw     string
	isRegex bool
	alwaysMatch bool
	matcher re.RegExp
}

//MatchString matches input string and return true if str was matches with filter and false if not
func (m *matcher) MatchString(str string) bool {

	if m.alwaysMatch {
		return true
	} else if m.isRegex {
		return m.matcher.MatchString(str)
	} else {
		return strings.Contains(str, m.raw)
	}
}

func (m *matcher) String() string {
	return m.raw
}

func newmatcher(str string) (matcher, error) {
	var err error
	m := matcher{}
	m.raw = str

	if str == ".+" || str == ".*" || str == "*" || str == "" {
		m.alwaysMatch = true
		log.Printf("filter [%s] was recognized as \"always match expression\"\n", str)
	} else if regularExpressionRex.MatchString(str) {
		m.isRegex = true
		log.Printf("filter [%s] was recognized as \"regular expersion\"\n", str)
		m.matcher, err = re.Compile(str)
	} else {
		log.Printf("filter [%s] was recognized as \"string match expression\"\n", str)
	}
	return m, err
}

func (m *matcher) UnmarshalJSON(data []byte) (err error) {
	*m, err = newmatcher(string(data[1 : len(data)-1]))
	return err
}

func (m *matcher) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {

	v := ""
	if err := unmarshal(&v); err != nil {
		return err
	}
	*m, err = newmatcher(v)
	return err
}
