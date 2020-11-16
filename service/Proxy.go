package service

import (
	"github.com/vskit-tv/vlog/log"
	"gopkg.in/resty.v1"
)

const (
	MaxRetryTimes = 3
)

var proxy = NewProxy()

type (
	Proxy struct {
		client    *resty.Client
		proxyHost string
	}
	Options struct {
		Method  string
		Path    string
		Headers map[string]string
		Body    string
	}
	Option func(*Options)
)

func Method(method string) Option {
	return func(o *Options) {
		o.Method = method
	}
}

func Path(path string) Option {
	return func(o *Options) {
		o.Path = path
	}
}

func Headers(headers map[string]string) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}

func Body(body string) Option {
	return func(o *Options) {
		o.Body = body
	}
}

func NewProxy() *Proxy {
	proxy := &Proxy{
		client:    resty.New(),
		proxyHost: "api.mylichking.com",
	}
	proxy.client.SetRetryCount(MaxRetryTimes)
	return proxy
}

func (p *Proxy) getUpstreamHost() string {
	return p.proxyHost
}

func ForwardRequest(options ...Option) *Response {
	var (
		resp *resty.Response
		err  error
		opts = Options{}
	)
	for _, option := range options {
		option(&opts)
	}
	url := "http://" + proxy.getUpstreamHost() + opts.Path
	req := proxy.client.R().SetHeaders(opts.Headers)
	if opts.Method == MethodGet {
		resp, err = req.Get(url)
	} else {
		resp, err = req.SetBody(opts.Body).Post(url)
	}
	if err != nil {
		log.Errorf("forward request failed err = %+v", err)
	}

	if err != nil || resp.StatusCode() != StatusOK {
		return &Response{
			Code:    StatusBadGateway,
			Headers: make(map[string]string),
		}
	}
	return &Response{
		Code:    StatusOK,
		Headers: make(map[string]string),
		Body:    resp.String(),
	}
}
