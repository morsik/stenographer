// Copyright 2014 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package httplog wraps a http.ResponseWriter to allow request logging.
package httplog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type httpLog struct {
	r      *http.Request
	w      http.ResponseWriter
	nBytes int
	code   int
	err    error
	start  time.Time
	body   string
}

// New returns a new ResponseWriter which provides a nice
// String() method for easy printing.  The expected usage is:
//   func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//     w = httplog.New(w, r, false)
//     defer log.Print(w)  // Prints out useful information about request AND response
//     ... do stuff ...
//   }
func New(w http.ResponseWriter, r *http.Request, logRequestBody bool) http.ResponseWriter {
	h := &httpLog{w: w, r: r, start: time.Now(), code: http.StatusOK}
	if logRequestBody {
		var buf bytes.Buffer
		_, h.err = io.Copy(&buf, r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(&buf)
		h.body = fmt.Sprintf(" RequestBody:%q", buf.String())
	}
	return h
}

// Header implements http.ResponseWriter.
func (h *httpLog) Header() http.Header {
	return h.w.Header()
}

// Write implements http.ResponseWriter and io.Writer.
func (h *httpLog) Write(data []byte) (int, error) {
	n, err := h.w.Write(data)
	h.nBytes += n
	if err != nil && h.err == nil {
		h.err = err
	}
	return n, err
}

// WriteHeader implements http.ResponseWriter.
func (h *httpLog) WriteHeader(code int) {
	h.code = code
	h.w.WriteHeader(code)
}

// String implements fmt.Stringer.
func (h *httpLog) String() string {
	var errstr string
	if h.err != nil {
		errstr = h.err.Error()
	}
	return fmt.Sprintf("Requester:%q Request:\"%v %v %v\" Time:%v Bytes:%v Code:%q Err:%q%v",
		h.r.RemoteAddr,
		h.r.Method,
		h.r.URL,
		h.r.Proto,
		time.Since(h.start),
		h.nBytes,
		http.StatusText(h.code),
		errstr,
		h.body)
}