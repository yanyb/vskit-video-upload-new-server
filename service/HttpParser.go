package service

import (
	"bytes"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/vskit-tv/vlog/log"
	"io"
	"net"
	"strconv"
	"strings"
)

type ParseStage int

const (
	LineStage ParseStage = iota
	HeaderStage
	BodyStage
	DoneStage
)

const (
	MethodGet  = "GET"
	MethodPost = "POST"
)

var (
	ErrParseRequestLine   = errors.New("parse request line failed")
	ErrParseRequestHeader = errors.New("parse request header failed")
	ErrParseRequestBody   = errors.New("parse request body failed")
	ErrParseMethod        = errors.New("parse method failed")
	ErrParseContentLength = errors.New("parse content length failed")
)

type Parser struct {
	Id            string
	Conn          net.Conn
	N             int
	ParseIndex    int
	Stage         ParseStage
	Data          []byte
	Method        string
	Line          string
	Path          string
	Headers       map[string]string
	ContentLength int
	ErrCh         chan struct{}
	ErrConn       chan struct{}
	Done          chan Request
}

func NewParser(conn net.Conn) *Parser {
	return &Parser{
		Id:      uuid.NewV4().String(),
		Conn:    conn,
		Data:    bytePool.Get().([]byte),
		Headers: make(map[string]string),
		ErrCh:   make(chan struct{}),
		ErrConn: make(chan struct{}),
		Done:    make(chan Request),
	}
}

func (p *Parser) ConstructRequest() Request {
	req := RequestOk{
		Id:            p.Id,
		Conn:          p.Conn,
		Body:          p.Data[p.ParseIndex-p.ContentLength : p.ParseIndex],
		Method:        p.Method,
		Path:          p.Path,
		Headers:       p.Headers,
		ContentLength: p.ContentLength,
	}

	data := bytePool.Get().([]byte)
	copy(data, p.Data[p.ParseIndex:p.N])
	bytePool.Put(p.Data)

	p.N -= p.ParseIndex
	p.ParseIndex = 0
	p.Stage = LineStage

	p.Data = data
	p.Method = ""
	p.Line = ""
	p.Headers = make(map[string]string)
	p.ContentLength = 0

	return &req
}

func (p *Parser) Release() {
	fmt.Printf("parser realease id = %s\n", p.Id)
	p.Conn.Close()
	bytePool.Put(p.Data)
}

func (p *Parser) Run() {
	go func() {
		// 读取消息,阻塞
		for {
			err := p.read()
			if err != nil {
				log.Errorf("parse read err = %+v", err)
				break
			}

			if err := p.parse(); err != nil {
				p.ErrCh <- struct{}{}
				break
			}
		}

		fmt.Printf("exist parser id = %s\n", p.Id)
	}()
}

func (p *Parser) parse() error {
	prevStage := p.Stage
	prevHeaderLen := len(p.Headers)

	for {
		switch p.Stage {
		case LineStage:
			if err := p.parseRequestLine(); err != nil {
				log.Errorf("err = %+v,received  data: %v", err, p.Data[:p.N])
				return err
			}
			if err := p.parseMethodAndPath(); err != nil {
				log.Errorf("err = %+v,received  data: %v", err, p.Data[:p.N])
				return err
			}
		case HeaderStage:
			if err := p.parseRequestHeader(); err != nil {
				log.Errorf("err = %+v,received data: %v", err, p.Data[:p.N])
				return err
			}
		case BodyStage:
			if p.Method == "POST" {
				if err := p.parseContentLength(); err != nil {
					log.Errorf("err = %+v,received data: %v", err, p.Data[:p.N])
					return err
				}
				if p.ContentLength <= len(p.Data[p.ParseIndex:p.N]) {
					p.ParseIndex += p.ContentLength
					p.Stage += 1
				}
			} else {
				p.Stage += 1
			}
		case DoneStage:
			p.Done <- p.ConstructRequest()
			return nil
		}

		// 解析进度是否变化
		if prevStage == p.Stage && prevHeaderLen == len(p.Headers) {
			break
		} else {
			prevStage = p.Stage
			prevHeaderLen = len(p.Headers)
		}
	}

	return nil
}

func (p *Parser) read() error {
	m, err := p.Conn.Read(p.Data[p.N:])
	if err != nil {
		if err == io.EOF {
			fmt.Println("conn closed")
		} else {
			log.Errorf("conn read err")
		}
		return err
	}
	p.N += m
	return nil
}

func (p *Parser) parseRequestLine() error {
	i := bytes.IndexByte(p.Data[p.ParseIndex:p.N], '\n')
	if i == -1 {
		return nil
	}
	if p.Data[i-1] != '\r' {
		return ErrParseRequestLine
	}
	p.Line = string(p.Data[p.ParseIndex : p.ParseIndex+i-1])
	p.ParseIndex += i + 1
	p.Stage += 1
	return nil
}

func (p *Parser) parseRequestHeader() error {
	i := bytes.IndexByte(p.Data[p.ParseIndex:p.N], '\n')
	if i == -1 {
		return nil
	}
	if p.Data[p.ParseIndex+i-1] != '\r' {
		return ErrParseRequestHeader
	}
	// 空行
	if i == 1 {
		p.Stage += 1
	} else {
		fieldSperator := bytes.IndexByte(p.Data[p.ParseIndex:p.ParseIndex+i+1], ':')
		if fieldSperator == -1 {
			return ErrParseRequestHeader
		}
		key := strings.ToLower(string(p.Data[p.ParseIndex : p.ParseIndex+fieldSperator]))
		value := strings.ToLower(strings.TrimSpace(string(p.Data[p.ParseIndex+fieldSperator+1 : p.ParseIndex+i-1])))
		p.Headers[key] = value
	}
	p.ParseIndex += i + 1
	return nil
}

func (p *Parser) parseMethodAndPath() error {
	splits := strings.SplitN(p.Line, " ", 3)
	if len(splits) < 2 {
		return ErrParseMethod
	}
	if splits[0] != MethodPost && splits[0] != MethodGet {
		return ErrParseMethod
	}
	p.Method = splits[0]
	p.Path = splits[1]
	return nil
}

func (p *Parser) parseContentLength() error {
	if lenVal, ok := p.Headers["content-length"]; ok {
		if contentLength, err := strconv.Atoi(lenVal); err == nil {
			p.ContentLength = contentLength
			return nil
		}
	}
	return ErrParseContentLength
}

func (p *Parser) Size() int {
	return p.N - p.ParseIndex
}
