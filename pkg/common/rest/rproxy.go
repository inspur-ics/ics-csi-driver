/*
Copyright (c) 2019 Inspur, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations
under the License.
*/

package rest

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const sessionTimeout = 30 * time.Second

// RestProxy - request client for any REST API
type RestProxy struct {
	addr          string
	port          int
	authToken     string
	httpRestProxy *http.Client
	prot          string
	user          string
	pass          string
	tries         int

	mu        sync.Mutex
	requestID int64
	timeout   int64
	sessionID string
}

// RestProxyInterface - request client interface
type RestProxyInterface interface {
	Send(method, path string, req interface{}) (int, []byte, error)
	Login() (int, []byte, error)
}

func (rp *RestProxy) Login() (int, []byte, error) {
	var res *http.Response
	var err error

	rp.mu.Lock()
	rp.requestID++
	rp.mu.Unlock()

	method := "POST"
	path := "authentication"
	addr := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addr, rp.port, path)

	klog.V(4).Infof("Send %s request to %s\n", method, addr)

	// send login request data as json
	var reader io.Reader

	loginReq := LoginReq{
		Username: rp.user,
		Password: rp.pass,
		Locale:   "cn",
		Domain:   "internal",
		Captcha:  "",
	}

	jdata, err := json.Marshal(loginReq)
	if err != nil {
		return 0, nil, err
	}

	reader = strings.NewReader(string(jdata))

	req, err := http.NewRequest(method, addr, reader)
	req.SetBasicAuth(rp.user, rp.pass)
	if err != nil {
		klog.Errorf("Unable to create req: %s", err)
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("version", "5.6")
	res, err = rp.httpRestProxy.Do(req)
	if err != nil {
		klog.Errorf("Request failed with error: %+v", err)
		return 0, nil, err
	}

	defer res.Body.Close()

	if err != nil {
		klog.Errorf("Request error: %+v", err)
		return 0, nil, err
	}

	// validate response body
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Response failure: %+v", err)
		err = status.Error(codes.Internal, "Unable to process response")
		return res.StatusCode, nil, err
	}

	var loginRsp LoginRsp
	if err := json.Unmarshal(bodyBytes, &loginRsp); err != nil {
		klog.Error("LoginRsp json unmarshal failed.")
	} else {
		rp.sessionID = loginRsp.SessionId
		//klog.V(4).Infof("Login SessionID:%s  RSP:%+v\n", rp.sessionID, loginRsp)
	}

	return res.StatusCode, bodyBytes, err
}

func (rp *RestProxy) Send(method, path string, data interface{}) (int, []byte, error) {
	var res *http.Response
	var err error

	rp.mu.Lock()
	rp.requestID++
	rp.mu.Unlock()

	addr := fmt.Sprintf("%s://%s:%d/%s", rp.prot, rp.addr, rp.port, path)

	klog.V(4).Infof("Send %s request to %s\n", method, addr)

	// send request data as json
	var reader io.Reader
	if data == nil {
		reader = nil
	} else {
		jdata, err := json.Marshal(data)
		if err != nil {
			return 0, nil, err
		}
		reader = strings.NewReader(string(jdata))
	}

	//rp.l.Debugf("Url %+v", addr)

	req, err := http.NewRequest(method, addr, reader)
	req.SetBasicAuth(rp.user, rp.pass)
	if err != nil {
		klog.Errorf("Unable to create req: %s", err)
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("version", "5.6")
	req.Header.Set("Authorization", rp.sessionID)
	res, err = rp.httpRestProxy.Do(req)
	if err != nil {
		klog.Errorf("Request failed with error: %+v", err)
		return 0, nil, err
	}

	defer res.Body.Close()

	if err != nil {
		klog.Errorf("Request error: %+v", err)
		return 0, nil, err
	}

	// validate response body
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		klog.Errorf("Response failure: %+v", err)
		err = status.Error(codes.Internal, "Unable to process response")
		return res.StatusCode, nil, err
	}
	return res.StatusCode, bodyBytes, err
}

type RestProxyCfg struct {
	Addr string
	Port int
	User string
	Pass string
}

var DefaultRestCfg RestProxyCfg

// TODO: implement sessions
func NewRestProxy() (ri RestProxyInterface, err error) {

	var timeoutDuration time.Duration
	idleTimeOut := "30s"
	//timeoutDuration, err = time.ParseDuration(cfg.IdleTimeOut)
	timeoutDuration, err = time.ParseDuration("30s")
	if err != nil {
		klog.Errorf("Uncorrect IdleTimeOut value: %s, Error %s", idleTimeOut, err)
		return nil, err
	}

	tr := &http.Transport{
		IdleConnTimeout: sessionTimeout,
		TLSClientConfig: &tls.Config{
			// Connect without checking certificate
			InsecureSkipVerify: true,
		},
	}

	httpRestProxy := &http.Client{
		Transport: tr,
		Timeout:   timeoutDuration,
	}

	ri = &RestProxy{
		addr:          DefaultRestCfg.Addr,
		port:          DefaultRestCfg.Port,
		httpRestProxy: httpRestProxy,
		requestID:     0,
		prot:          "https",
		user:          DefaultRestCfg.User,
		pass:          DefaultRestCfg.Pass,
		tries:         3,
	}

	stat, _, err := ri.Login()
	if err != nil {
		klog.Error("Internal failure in communication with ics")
	} else {
		klog.V(4).Infof("response stat:%v\n", stat)
	}

	return ri, err
}
