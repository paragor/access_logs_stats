package input

import (
	"errors"
	"regexp"
)

//todo make normal parser
type syslogMessage struct {
	Priority    string
	Date        string
	Hostname    string
	Application string
	Message     string
}

var UNKNOWN_INPUT_STRING_FORMAT = errors.New("UNKNOWN_INPUT_STRING_FORMAT")

func NewSyslogParser() (p *syslogParser, err error) {
	p = new(syslogParser)
	p.reg1, err = regexp.Compile(`<(\d+)>(\S+\s+\w+\s+\S+)\s+(\S+)\s+(\w+):\s*(.+)`)
	if err != nil {
		return
	}
	p.reg2, err = regexp.Compile(`<(\d+)>\s*([0-9\-T:.+]+)\s*(\S+)\s*(\S+)[^\]]+\]\s*(\S+)`)

	return p, err
}

type syslogParser struct {
	reg1 *regexp.Regexp
	reg2 *regexp.Regexp
}

func (s *syslogParser) ParseSyslogMsg(str string) (m syslogMessage, err error) {
	var matches []string

	//examples of syslog message
	//<9>Oct  5 13:46:36 fzozo: fooo
	//<149>Oct  7 13:51:20 node2.drom.ru nginx: s.auto.drom.ru s.rdrom.ru 31.173.227.172 - [2016-10-07T13:51:20+10:00] GET "/1/catalog/photos/generations/toyota_passo_g779.jpg?17911" HTTP/1.1 200 4333 "http://www.drom.ru/catalog/toyota/passo/" "Mozilla/5.0 (iPhone; CPU iPhone OS 9_3_5 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13G36 Safari/601.1" 0.000 381 "-" "-" HIT "-" - -56 Safari/602.1" 0.000 644 "-" "-" HIT "-" f87a666AIsfgMIILIrOCyVLlE10KQ0ab -0a7 19665144%3Aabea9d3121a0dba90d9279d31648c8fd
	//<12>1 2016-10-06T09:58:23.079047+10:00 salionov-drom zzz - - [timeQuality tzKnown="1" isSynced="1" syncAccuracy="1037287"] fooo

	m = syslogMessage{}
	matches = s.reg1.FindStringSubmatch(str)
	if len(matches) != 0 {
		m.Priority, m.Date, m.Hostname, m.Application, m.Message = matches[1], matches[2], matches[3], matches[4], matches[5]
	} else {
		matches := s.reg2.FindStringSubmatch(str)
		//log.Println(matches)
		if len(matches) != 0 {
			m.Priority, m.Date, m.Hostname, m.Application, m.Message = matches[1], matches[2], matches[4], matches[5], matches[6]
		} else {
			err = UNKNOWN_INPUT_STRING_FORMAT
		}
	}
	return m, err
}