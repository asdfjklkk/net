package http

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type HttpClient struct {
	headers map[string]string
	//
	TLSInsecureSkipVerify bool
	TLSHandshakeTimeout   time.Duration
	DialTimeout           time.Duration
	DialKeepAlive         time.Duration
	Cookie                *cookiejar.Jar
	Timeout               time.Duration
	UserAgent             string
	ContentType           string
	BindIp                string
	Proxy                 string
}

func NewHttpClient() (returnValue HttpClient) {
	returnValue = HttpClient{}
	if returnValue.Cookie == nil {
		returnValue.Cookie, _ = cookiejar.New(nil)
	}
	if returnValue.headers == nil {
		returnValue.headers = make(map[string]string)
	}
	return
}

func (obj HttpClient) GetHeader(name string) (returnValue string) {
	returnValue = obj.headers[name]
	return
}

func (obj HttpClient) SetHeader(name string, value string) {
	obj.headers[name] = value
}

func (obj HttpClient) RemoveHeader(name string) {
	delete(obj.headers, name)
}

func (obj HttpClient) ClearHeaders() {
	for k, _ := range obj.headers {
		delete(obj.headers, k)
	}
}

func (obj HttpClient) GetResponseData(method string, u string, postData []byte) (returnValue HttpResponseData) {
	defer func() {
		if err := recover(); err != nil {
			returnValue.Error = fmt.Errorf("%v", err)
		}
	}()
	method = strings.ToUpper(method)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: obj.TLSInsecureSkipVerify},
	}
	tr.TLSHandshakeTimeout = obj.TLSHandshakeTimeout
	//代理伺服器 - 開始
	if obj.Proxy != "" {
		proxyUrl, err := url.Parse(obj.Proxy)
		if err != nil {
			returnValue.Error = err
			return
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	//代理伺服器 - 結束
	//綁定 IP - 開始
	if obj.BindIp != "" {
		localAddr, err := net.ResolveIPAddr("ip", obj.BindIp)
		if err != nil {
			returnValue.Error = err
			return
		}
		localTCPAddr := net.TCPAddr{
			IP: localAddr.IP,
		}
		tr.Dial = (&net.Dialer{
			LocalAddr: &localTCPAddr,
			Timeout:   obj.DialTimeout * time.Millisecond,
			KeepAlive: obj.DialKeepAlive * time.Millisecond,
		}).Dial
	}
	//綁定 IP - 結束
	//Request - 開始
	request, err := http.NewRequest(method, u, bytes.NewBuffer(postData))
	if err != nil {
		returnValue.Error = err
		return
	}
	if obj.UserAgent != "" {
		request.Header.Set("User-Agent", obj.UserAgent)
	}
	if obj.ContentType != "" && method == "POST" {
		request.Header.Set("Content-Type", obj.ContentType)
	}
	for k, v := range obj.headers {
		request.Header.Set(k, v)
	}
	//Request - 結束
	//HttpClient - 開始
	client := &http.Client{}
	client.Timeout = obj.Timeout
	client.Jar = obj.Cookie
	client.Transport = tr
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if obj.UserAgent != "" {
			req.Header.Set("User-Agent", obj.UserAgent)
		}
		if obj.ContentType != "" && method == "POST" {
			req.Header.Set("Content-Type", obj.ContentType)
		}
		for k, v := range obj.headers {
			req.Header.Set(k, v)
		}
		return nil
	}
	//HttpClient - 結束
	response, err := client.Do(request)
	if err != nil {
		returnValue.Error = err
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		returnValue.Error = err
		return
	}
	returnValue.Status = response.Status
	returnValue.StatusCode = response.StatusCode
	returnValue.Proto = response.Proto
	returnValue.ProtoMajor = response.ProtoMajor
	returnValue.ProtoMinor = response.ProtoMinor
	returnValue.Header = response.Header
	returnValue.ContentLength = response.ContentLength
	returnValue.Body = contents
	returnValue.Request = response.Request
	return
}
