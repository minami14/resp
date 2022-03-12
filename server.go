package resp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	readerPool sync.Pool
	writerPool sync.Pool
)

type Server struct {
	Logger            *log.Logger
	Handler           func(*Request, *ResponseWriter) error
	commands          map[string]func(*Request, *ResponseWriter) error
	BadRequestHandler func(error, *ResponseWriter)
}

func (s *Server) Serve(l net.Listener) error {
	for {
		con, err := l.Accept()
		if err != nil {
			s.Logger.Printf("resp: Accept error: %v", err)
		}

		c := s.newConn(con)
		go c.serve()
	}
}

func (s *Server) newConn(con net.Conn) *conn {
	return &conn{
		server: s,
		con:    con,
		br:     newReader(con),
		bw:     newWriter(con),
	}
}

func (s *Server) Handle(command string, handler func(*Request, *ResponseWriter) error) {
	command = strings.ToUpper(command)
	if s.commands == nil {
		s.commands = map[string]func(*Request, *ResponseWriter) error{}
	}
	s.commands[command] = handler
	s.Handler = func(r *Request, rw *ResponseWriter) error {
		commandUpper := strings.ToUpper(r.Command)
		cmd, ok := s.commands[commandUpper]
		if !ok {
			return rw.WriteError(fmt.Sprintf("ERR unknown command `%v`", r.Command))
		}
		return cmd(r, rw)
	}
}

func newReader(r io.Reader) *bufio.Reader {
	if v := readerPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func newWriter(w io.Writer) *bufio.Writer {
	if v := writerPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriter(w)
}

type conn struct {
	server *Server
	con    net.Conn
	br     *bufio.Reader
	bw     *bufio.Writer
}

func (c *conn) serve() {
	defer readerPool.Put(c.br)
	defer writerPool.Put(c.bw)
	defer c.con.Close()
	for {
		req, err := ParseRequest(c.br)
		wr := &ResponseWriter{writer: c.bw}
		if err != nil {
			if c.server.BadRequestHandler != nil {
				c.server.BadRequestHandler(err, wr)
			}
			break
		}
		err = c.server.Handler(req, wr)
		if err != nil {
			break
		}
	}
}
