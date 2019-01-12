package git

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

func registerManagedHttp() error {
	registeredSmartTransport, err := NewRegisteredSmartTransport("http", true, httpSmartSubtransportFactory)
	if err != nil {
		registeredSmartTransport.Free()
	}
	return err
}

func registerManagedHttps() error {
	registeredSmartTransport, err := NewRegisteredSmartTransport("https", true, httpSmartSubtransportFactory)
	if err != nil {
		registeredSmartTransport.Free()
	}
	return err
}

func httpSmartSubtransportFactory(remote *Remote, transport *Transport) (SmartSubtransport, error) {
	var proxyFn func(*http.Request) (*url.URL, error)
	proxyOpts, err := transport.SmartProxyOptions()
	if err != nil {
		return nil, err
	}
	switch proxyOpts.Type {
	case ProxyTypeNone:
		proxyFn = nil
	case ProxyTypeAuto:
		proxyFn = http.ProxyFromEnvironment
	case ProxyTypeSpecified:
		parsedUrl, err := url.Parse(proxyOpts.Url)
		if err != nil {
			return nil, err
		}

		proxyFn = http.ProxyURL(parsedUrl)
	}

	return &httpSmartSubtransport{
		transport: transport,
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: proxyFn,
			},
		},
	}, nil
}

type httpSmartSubtransport struct {
	transport *Transport
	client    *http.Client
}

func (t *httpSmartSubtransport) Action(url string, action SmartServiceAction) (SmartSubtransportStream, error) {
	var req *http.Request
	var err error
	switch action {
	case SmartServiceActionUploadpackLs:
		req, err = http.NewRequest("GET", url+"/info/refs?service=git-upload-pack", nil)

	case SmartServiceActionUploadpack:
		req, err = http.NewRequest("POST", url+"/git-upload-pack", nil)
		if err != nil {
			break
		}
		req.Header.Set("Content-Type", "application/x-git-upload-pack-request")

	case SmartServiceActionReceivepackLs:
		req, err = http.NewRequest("GET", url+"/info/refs?service=git-receive-pack", nil)

	case SmartServiceActionReceivepack:
		req, err = http.NewRequest("POST", url+"/info/refs?service=git-upload-pack", nil)
		if err != nil {
			break
		}
		req.Header.Set("Content-Type", "application/x-git-receive-pack-request")

	default:
		err = errors.New("unknown action")
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "git/2.0 (git2go)")

	stream := newManagedHttpStream(t, req)
	if req.Method == "POST" {
		stream.recvReply.Add(1)
		stream.sendRequestBackground()
	}

	return stream, nil
}

func (t *httpSmartSubtransport) Close() error {
	return nil
}

func (t *httpSmartSubtransport) Free() {
	t.client = nil
}

type httpSmartSubtransportStream struct {
	owner       *httpSmartSubtransport
	req         *http.Request
	resp        *http.Response
	reader      *io.PipeReader
	writer      *io.PipeWriter
	sentRequest bool
	recvReply   sync.WaitGroup
	httpError   error
}

func newManagedHttpStream(owner *httpSmartSubtransport, req *http.Request) *httpSmartSubtransportStream {
	r, w := io.Pipe()
	return &httpSmartSubtransportStream{
		owner:  owner,
		req:    req,
		reader: r,
		writer: w,
	}
}

func (self *httpSmartSubtransportStream) Read(buf []byte) (int, error) {
	if !self.sentRequest {
		self.recvReply.Add(1)
		if err := self.sendRequest(); err != nil {
			return 0, err
		}
	}

	if err := self.writer.Close(); err != nil {
		return 0, err
	}

	self.recvReply.Wait()

	if self.httpError != nil {
		return 0, self.httpError
	}

	return self.resp.Body.Read(buf)
}

func (self *httpSmartSubtransportStream) Write(buf []byte) (int, error) {
	if self.httpError != nil {
		return 0, self.httpError
	}
	return self.writer.Write(buf)
}

func (self *httpSmartSubtransportStream) Free() {
	if self.resp != nil {
		self.resp.Body.Close()
	}
}

func (self *httpSmartSubtransportStream) sendRequestBackground() {
	go func() {
		self.httpError = self.sendRequest()
	}()
	self.sentRequest = true
}

func (self *httpSmartSubtransportStream) sendRequest() error {
	defer self.recvReply.Done()
	self.resp = nil

	var resp *http.Response
	var err error
	var userName string
	var password string
	for {
		req := &http.Request{
			Method: self.req.Method,
			URL:    self.req.URL,
			Header: self.req.Header,
		}
		if req.Method == "POST" {
			req.Body = self.reader
			req.ContentLength = -1
		}

		req.SetBasicAuth(userName, password)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()

			cred, err := self.owner.transport.SmartCredentials("", CredTypeUserpassPlaintext)
			if err != nil {
				return err
			}
			defer cred.Free()

			userName, password, err = cred.GetUserpassPlaintext()
			if err != nil {
				return err
			}

			continue
		}

		// Any other error we treat as a hard error and punt back to the caller
		resp.Body.Close()
		return fmt.Errorf("Unhandled HTTP error %s", resp.Status)
	}

	self.sentRequest = true
	self.resp = resp
	return nil
}
