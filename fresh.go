// Copyright 2018 tree xie
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

package fresh

import (
	"net/http"

	"github.com/vicanso/cod"
	"github.com/vicanso/fresh"
)

type (
	// Config fresh config
	Config struct {
		Skipper cod.Skipper
	}
)

// NewDefault create a default ETag middleware
func NewDefault() cod.Handler {
	return New(Config{})
}

// New create a fresh checker
func New(config Config) cod.Handler {
	skipper := config.Skipper
	if skipper == nil {
		skipper = cod.DefaultSkipper
	}
	return func(c *cod.Context) (err error) {
		if skipper(c) {
			return c.Next()
		}
		err = c.Next()
		if err != nil {
			return
		}
		// 如果空数据或者已经是304，则跳过
		bodyBuf := c.BodyBuffer
		if bodyBuf == nil || bodyBuf.Len() == 0 || c.StatusCode == http.StatusNotModified {
			return
		}

		// 如果非GET HEAD请求，则跳过
		method := c.Request.Method
		if method != http.MethodGet && method != http.MethodHead {
			return
		}

		// 如果响应状态码不为0 而且( < 200 或者 >= 300)，则跳过
		// 如果未设置状态码，最终则为200
		statusCode := c.StatusCode
		if statusCode != 0 &&
			(statusCode < http.StatusOK ||
				statusCode >= http.StatusMultipleChoices) {
			return
		}

		// 304的处理
		if fresh.Fresh(c.Request.Header, c.Headers) {
			c.NotModified()
		}
		return
	}
}
