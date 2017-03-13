package input

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"regexp"
)

const UDP_SAFE_PACK_SIZE = 2048

var (
	INCORRENCT_DSN   = errors.New("Incorrect DSN")
	UNKNOWN_PROTOCOL = errors.New("Unknown protocol")
)

type SyslogInputReader struct {
	BufferedReader

	chanLock chan bool
	buffer   []byte

	protocol    string
	listen      string
	application string

	acceptor Acceptor

	parser *syslogParser
}

func CreateSyslogInputReader(dsn string) (r *SyslogInputReader, err error) {

	//dsn examples::
	// syslog:udp:binding_ip:binding_port/application
	// syslog:tcp:binding_ip:binding_port/application
	// syslog:tcp:binding_ip:binding_port/application

	r = &SyslogInputReader{}
	r.chanLock = make(chan bool, 1)
	r.buffer = []byte{}
	r.parser, err = NewSyslogParser()
	check(err)
	//read dsn
	err = r.parseDsn(dsn)
	if err != nil {
		return
	}

	//create udp or tcp server
	err = r.startServer()

	return
}

func (r *SyslogInputReader) parseDsn(dsn string) (err error) {
	r.protocol, r.listen, r.application, err = ParseSyslogDsn(dsn)
	return
}

func (r *SyslogInputReader) ReadToBuffer() {

	if r.protocol == "udp" {
		r.readToBufferUDP()
	} else {
		r.readToBufferTCP()
	}

}

func (r *SyslogInputReader) readToBufferTCP() {
	// accept new connect and send it to handler
	for {
		conn, err := r.acceptor.Accept()
		check(err)
		go r.handleConnectionTCP(conn)
	}
}

func (r *SyslogInputReader) readToBufferUDP() {
	//create one listener and handle messages in one loop
	conn, err := r.acceptor.Accept()
	check(err)
	r.handleConnectionUDP(conn)
}

func (r *SyslogInputReader) FlushBuffer() []byte {
	r.chanLock <- true
	buffer := r.buffer
	r.buffer = []byte{}
	<-r.chanLock
	return buffer
}

func (r *SyslogInputReader) Close() {
	r.acceptor.Close()
}

func (r *SyslogInputReader) startServer() (err error) {

	var (
		acceptor Acceptor
	)

	if r.protocol == "udp" {
		acceptor, err = r.getUDPAcceptor()
		if err != nil {
			return err
		}
	} else {
		acceptor, err = r.getTCPAcceptor()
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	r.acceptor = acceptor

	return
}

func (r *SyslogInputReader) getUDPAcceptor() (acceptor Acceptor, err error) {

	addr, err := net.ResolveUDPAddr(r.protocol, r.listen)
	if err != nil {
		return acceptor, err
	}
	udpconn, err := net.ListenUDP(r.protocol, addr)
	acceptor = &udpAcceptor{acceptor: udpconn}

	return acceptor, err
}

func (r *SyslogInputReader) getTCPAcceptor() (acceptor Acceptor, err error) {

	addr, err := net.ResolveTCPAddr(r.protocol, r.listen)
	if err != nil {
		return acceptor, err
	}
	tcpl, err := net.ListenTCP(r.protocol, addr)
	acceptor = &tcpAcceptor{acceptor: tcpl}
	return acceptor, err
}

func (r *SyslogInputReader) handleConnectionUDP(conn net.Conn) {
	var (
		read int
		err  error
		b    []byte
	)
	b = make([]byte, UDP_SAFE_PACK_SIZE)
	defer conn.Close()

	for {
		read, err = conn.Read(b)
		if err != nil {
			log.Println(err)
		}
		bytesBuf := b[0:read]
		r.chanLock <- true
		if r.appendToBuffer(bytesBuf) {
			r.buffer = append(r.buffer, '\n')
		}
		<-r.chanLock
		//log.Println(string(r.buffer))
	}
}

func (r *SyslogInputReader) handleConnectionTCP(conn net.Conn) {
	buffer := bufio.NewReader(conn)
	defer conn.Close()

	for {
		bytesBuf, err := buffer.ReadBytes('\n')
		if err == io.EOF {
			r.chanLock <- true
			if r.appendToBuffer(bytesBuf) {
				r.buffer = append(r.buffer, '\n')
			}
			<-r.chanLock
			break
		} else if err != nil {
			check(err)
		}
		r.chanLock <- true
		r.appendToBuffer(bytesBuf)
		<-r.chanLock
	}
	//log.Println(string(r.buffer))
}

func (r *SyslogInputReader) appendToBuffer(byteBuf []byte) bool {
	if len(byteBuf) == 0 {
		return false
	}

	//parse message.
	m, err := r.parser.ParseSyslogMsg(string(byteBuf))

	if err == UNKNOWN_INPUT_STRING_FORMAT {
		log.Println(UNKNOWN_INPUT_STRING_FORMAT, string(byteBuf))
	}
	//Filter by application
	if m.Application != r.application {
		return false
	}

	r.buffer = append(r.buffer, m.Message...)
	return true
}

func ParseSyslogDsn(dsn string) (protocol string, listen string, application string, err error) {
	re, err := regexp.Compile(`(syslog):([a-zA-Z0-9]+):([^/]+)/(\S+)`)
	if err != nil {
		return
	}

	matches := re.FindStringSubmatch(dsn)
	if len(matches) != 5 {
		err = INCORRENCT_DSN
		return
	}
	protocol = matches[2]
	listen = matches[3]
	application = matches[4]

	if protocol != "udp" && protocol != "tcp" {
		err = UNKNOWN_PROTOCOL
	}

	return
}

type Acceptor interface {
	Accept() (net.Conn, error)
	Close() error
}

type tcpAcceptor struct {
	acceptor *net.TCPListener
}

func (l *tcpAcceptor) Accept() (net.Conn, error) {
	return l.acceptor.Accept()
}

func (l *tcpAcceptor) Close() error {
	return l.acceptor.Close()
}

type udpAcceptor struct {
	acceptor *net.UDPConn
}

func (l *udpAcceptor) Accept() (net.Conn, error) {
	return l.acceptor, nil
}

func (l *udpAcceptor) Close() error {
	return l.acceptor.Close()
}
