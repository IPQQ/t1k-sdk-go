go snserver SDK
===============

使用 golang 访问 snserver 检测功能的 API，基于 T1K v2 协议，提供了连接池管理、响应检测等高级功能。

## 1. 包结构

``` go
import (
	"git.in.chaitin.net/patronus/snserver/sdk/go/pkg/gosnserver"
	"git.in.chaitin.net/patronus/snserver/sdk/go/pkg/detection"
	"git.in.chaitin.net/patronus/snserver/sdk/go/pkg/t1k"
	"git.in.chaitin.net/patronus/snserver/sdk/go/pkg/misc"
)
```

一般来说，简单的使用只需要用到 `gosnserver` `detection` 两个包。前者包含了和 snserver 通信的主要逻辑，包括请求检测、响应检测、连接池、心跳发送等；后者规定了请求和响应的接口，并提供了检测过程中数据的保存上下文。

`t1k` 包实现了底层 T1K 协议通信。`misc` 包包括一些零碎的逻辑，例如时间戳生成、事件ID生成等，和一些调试工具。

## 2. 使用范例

### 2.1 检测请求

假设 `req` 是一个 `*http.Request`

``` go
server, err := gosnserver.New("169.254.0.5:8000")
panicIf(err)
defer server.Close()

result, err := server.DetectHttpRequest(req)
panicIf(err)

if result.Passed() {
	fmt.Println("Passed")
}
if result.Blocked() {
	fmt.Println("Blocked")
}
```

完整的检测代码参见 [detect_request](sdk/go/examples/detect_request/main.go)

### 2.2 检测请求和响应

为了将请求和响应相互关联，并暂存检测中需要到中间数据，需要一个 `detection.DetectionContext` 作为请求和响应的“容器”。

``` go
dc := detection.New()
detection.MakeHttpRequestInCtx(req, dc)
detection.MakeHttpResponseInCtx(rsp, dc)

server, err := gosnserver.New("169.254.0.5:8000")
panicIf(err)
defer server.Close()

resultReq, resultRsp, err := server.Detect(dc)
panicIf(err)
```

完整的检测代码参见 [detect_request_and_response](sdk/go/examples/detect_request_and_response/main.go)

### 2.3 不使用连接池检测

所有形如 `DetectSomething` 的函数都有不需要连接池的版本，直接输入一个 `io.ReadWriter` 即可：

``` go
conn, err := net.Dial("tcp", "169.254.0.5:8000")
panicIf(err)
defer conn.Close()

result, err := gosnserver.DetectHttpRequest(conn, req)
panicIf(err)
```

完整的检测代码参见 [detect_request_with_socket](sdk/go/examples/detect_request_with_socket/main.go)

### 2.4 嵌入集成请求检测到 http server

在 go 中，实现一个 http server 十分简单，其核心便是 `handler` 函数。在 `handler` 开头加入如下代码，即可嵌入雷池的请求检测能力：

``` go
result, err := snserver.DetectHttpRequest(req)
if err != nil {
	fmt.Printf("error in detection: %s\n", err)
} else {
	if result.Blocked() {
		fmt.Fprintf(w, "blocked\n")
		return
	}
}
```

在 [http_server_embedded](sdk/go/examples/http_server_embedded/main.go) 范例中，有一个完成的例子。在雷池单机环境运行后，会在 8090 端口开启一个示例服务器，我们可以用 curl 测试拦截的效果：

``` bash
# curl 127.0.0.1:8090/webshell.php
blocked
# curl 127.0.0.1:8090
hello
```

### 2.5 嵌入集成请求和响应检测到 http server

嵌入响应检测有很多方式，在 [http_server_embedded_with_response_detection](sdk/go/examples/http_server_embedded_with_response_detection/main.go) 示例中，给出了一个包装 `handler` 函数的方式：

``` go
func wrapHandlerFunc(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		dc := detection.New()

		// detect the request
		detection.MakeHttpRequestInCtx(req, dc)
		result, err := snserver.DetectRequestInCtx(dc)
		if err != nil {
			fmt.Printf("error in detection: %s\n", err)
		} else {
			if result.Blocked() {
				fmt.Fprintf(w, "blocked\n")
				return
			}
		}

		rec := httptest.NewRecorder()
		f(rec, req)
		rsp := rec.Result()

		// detect the response
		detection.MakeHttpResponseInCtx(rsp, dc)
		result, err = snserver.DetectResponseInCtx(dc)
		if err != nil {
			fmt.Printf("error in detection: %s\n", err)
		} else {
			if result.Blocked() {
				fmt.Fprintf(w, "blocked\n")
				return
			}
		}

		for key, values := range rsp.Header {
			w.Header()[key] = values
		}
		_, err = io.Copy(w, rsp.Body)
		rsp.Body.Close()
	}
}
```

### 2.6 绕过 go http 自定义请求和响应格式

golang 的 `http.Request` 和 `http.Response` 使用起来有诸多限制，有时候在使用一些第三方的网络框架，或者是用于转发请求时，会希望不去使用 golang 的 `net/http` 库，而是自己定义 `Request` `Response` 结构。`detection.Request` 和 `detection.Response` 都是简单的接口，只要封装成这些接口，就可以绕过 `net/http` 库接入任意的请求和响应。

下面是简单的封装自定义请求的例子：

``` go
type MyCustomRequest struct {
	dc *detection.DetectionContext
}

func MakeMyCustomRequestInCtx(dc *detection.DetectionContext) *MyCustomRequest {
	ret := &MyCustomRequest{
		dc: dc,
	}
	dc.Request = ret
	return ret
}

func (r *MyCustomRequest) Header() ([]byte, error) {
	return []byte(
		"POST /form.php HTTP/1.1\r\n" +
			"Host: a.com\r\n" +
			"Content-Length: 40\r\n" +
			"Content-Type: application/json\r\n\r\n",
	), nil
}

func (r *MyCustomRequest) Body() (uint32, io.ReadCloser, error) {
	body := "{\"name\": \"youcai\", \"password\": \"******\"}"
	return uint32(len(body)), ioutil.NopCloser(bytes.NewReader([]byte(body))), nil
}

func (r *MyCustomRequest) Extra() ([]byte, error) {
	return detection.GenRequestExtra(r.dc), nil
}
```

下面是简单的封装自定义响应的例子：

``` go
type MyCustomResponse struct {
	dc *detection.DetectionContext
}

func MakeMyCustomResponseInCtx(dc *detection.DetectionContext) *MyCustomResponse {
	ret := &MyCustomResponse{
		dc: dc,
	}
	dc.Response = ret
	return ret
}

func (r *MyCustomResponse) RequestHeader() ([]byte, error) {
	return r.dc.Request.Header()
}

func (r *MyCustomResponse) Header() ([]byte, error) {
	return []byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Length: 29\r\n" +
			"Content-Type: application/json\r\n\r\n",
	), nil
}

func (r *MyCustomResponse) Body() (uint32, io.ReadCloser, error) {
	body := "{\"err\": \"password-incorrect\"}"
	return uint32(len(body)), ioutil.NopCloser(bytes.NewReader([]byte(body))), nil
}

func (r *MyCustomResponse) Extra() ([]byte, error) {
	return detection.GenResponseExtra(r.dc), nil
}

func (r *MyCustomResponse) T1KContext() ([]byte, error) {
	return r.dc.T1KContext, nil
}
```

完整的检测代码参见 [custom_request_response](sdk/go/examples/custom_request_response/main.go)

