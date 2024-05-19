// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_proxy

import (
	// "case_proxy/infer"
	"case_proxy/rpc_scheduler"
	"context"
	"encoding/json"
	"fmt"
	// "io"
	"log"
	// "net/http"
	"sync"
	"time"
	triton "case_proxy/grpc-client"
	"encoding/binary"
	"bytes"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PRCProxy struct {
	hostMap      map[string]*triton.GRPCInferenceServiceClient
	lb           rpc_scheduler.RPCScheduler
	Ctx          context.Context
	sync.RWMutex // protect alive
	alive        map[string]bool
}
type Resp struct {
	Output  float32 `json:"output"`
	Perfect bool    `json:"perfect"`

	// Rank	int8 `json:"rank"`
	// KeysEvicted uint64 `json:"keys_evicted"`
	// Ratio				float64	`json:"ratio"`
}

type Statistics struct{
	Host string
	Perfect bool
	Du int64
	Psdu int64
	Total int64
}
var Datach = make(chan Statistics, 10000)


// func test(){
// 	name := "World"
// 	addr := "127.0.0.1:50051"
// 	// 连接gRPC服务器
//         conn, err := grpc.Dial( addr , grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		fmt.Println("connect error to: ",addr)
// 	}
// 	defer conn.Close()

//         // 实例化一个client对象，传入参数conn
// 	c := infer.NewMyServiceClient(conn)

// 	// 初始化上下文，设置请求超时时间为1秒
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	//延迟关闭请求会话
// 	defer cancel()

// 	// 调用SayHello方法，以请求服务，然后得到响应消息
// 	r, err := c.SayHello(ctx, &.HelloRequest{Name: name})
// 	if err != nil {
// 		fmt.Println("can not greet to: ",addr)
// 	} else {
// 		fmt.Println("response from server: ",r.GetMessage())
// 	}
// }

func (h *PRCProxy) Setlb(targetHosts []string, algorithm string) (
	*PRCProxy, error) {
	hosts := make([]string, 0)
	hostMap := make(map[string]*triton.GRPCInferenceServiceClient)
	// alive := make(map[string]bool)
	for _, targetHost := range targetHosts {
		// url, err := url.Parse(targetHost)
		// if err != nil {
		// 	return nil, err
		// }
		conn, err := grpc.Dial(targetHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println("connect error to: ", targetHost)
		}
		proxy := triton.NewGRPCInferenceServiceClient(conn)
		hostMap[targetHost] = &proxy
		hosts = append(hosts, targetHost)
	}

	lb, err := rpc_scheduler.Build(algorithm, hosts)
	if err != nil {
		return nil, err
	}
	h.lb = lb
	return h, nil
}

// NewHTTPProxy create  new reverse proxy with url and balancer algorithm
func NewPRCProxy(targetHosts []string, algorithm string) (
	*PRCProxy, error) {

	hosts := make([]string, 0)
	hostMap := make(map[string]*triton.GRPCInferenceServiceClient)
	alive := make(map[string]bool)
	for _, targetHost := range targetHosts {
		// url, err := url.Parse(targetHost)
		// if err != nil {
		// 	return nil, err
		// }
		conn, err := grpc.Dial(targetHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println("connect error to: ", targetHost)
		}
		proxy := triton.NewGRPCInferenceServiceClient(conn)
		hostMap[targetHost] = &proxy
		hosts = append(hosts, targetHost)
	}

	lb, err := rpc_scheduler.Build(algorithm, hosts)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	return &PRCProxy{
		hostMap: hostMap,
		lb:      lb,
		Ctx:     ctx,
		alive:   alive,
	}, nil
}


func (h *PRCProxy)FasthttpInferRequest(httpContext *fasthttp.RequestCtx) {
	// start := time.Now()
	var rr rpc_scheduler.Req
	if err := json.Unmarshal(httpContext.PostBody(), &rr); err != nil {
		httpContext.Error(err.Error(), fasthttp.StatusBadGateway)
		return
	}
	host, err := h.lb.Schedule("", rr)
	if err != nil {
		httpContext.Error(fmt.Sprintf("balance error: %s", err.Error()), fasthttp.StatusBadGateway)
		return
	}
	h.lb.Inc(host)
	defer h.lb.Done(host)
	c := h.hostMap[host]
	// result, err := cc.Infer(ctx, &infer.ReqData{Keys: rr.Keys[0]})
	result := ModelInferRequest(*c,rr.Keys[0],"embed_cache","1")
	if err != nil {
		fmt.Println("can not greet to: ", host,err)
		httpContext.Error(fmt.Sprintf("can not greet to: ", host,err), fasthttp.StatusInternalServerError)
	} 

	// fmt.Println(st)
	resp := Resp{Output: result[0], Perfect: false}
	// resp := Resp{}
	if err := json.NewEncoder(httpContext).Encode(resp); err != nil {
		httpContext.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	// h.hostMap[host].ServeHTTP(w, r)
	// end := time.Now()
	// st :=Statistics{Host:host,Perfect: result.Perfect,Du:result.Duration,Psdu: result.Psdu,Total: int64(end.Sub(start))  }
	// Datach <- st
}






// ServeHTTP implements a proxy to the http server
// func (h *PRCProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	start := time.Now()
// 	defer func() {
// 		if err := recover(); err != nil {
// 			log.Printf("proxy causes panic :%s", err)
// 			w.WriteHeader(http.StatusBadGateway)
// 			_, _ = w.Write([]byte(err.(error).Error()))
// 		}
// 	}()
// 	var rr rpc_scheduler.Req
// 	body, _ := io.ReadAll(r.Body)
// 	err := json.Unmarshal(body, &rr)
// 	if err != nil {
// 		return
// 	}
// 	host, err := h.lb.Schedule("", rr)
// 	// fmt.Println(h.lb.ToString())
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadGateway)
// 		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
// 		return
// 	}
// 	h.lb.Inc(host)
// 	defer h.lb.Done(host)
// 	c := h.hostMap[host]
// 	cc := *c
// 	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
// 	result, err := cc.Infer(ctx, &infer.ReqData{Keys: rr.Keys[0]})
// 	if err != nil {
// 		fmt.Println("can not greet to: ", host,err)
// 	} 

// 	// fmt.Println(st)
// 	resp := Resp{Output: result.Output, Perfect: result.Perfect}
// 	jsonData, err := json.Marshal(resp)
// 	if err != nil {
// 		// 处理错误
// 		http.Error(w, "Error serializing struct to JSON", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(jsonData)
// 	// h.hostMap[host].ServeHTTP(w, r)
// 	end := time.Now()
// 	st :=Statistics{Host:host,Perfect: result.Perfect,Du:result.Duration,Psdu: result.Psdu,Total: int64(end.Sub(start))  }
// 	Datach <- st
// }

func Preprocess(inputData0 []int32) []byte {

	var inputBytes0 []byte
	// Temp variable to hold our converted int32 -> []byte
	bs := make([]byte, 8)
	for i := 0; i < 26; i++ {
		binary.LittleEndian.PutUint64(bs, uint64(inputData0[i]))
		inputBytes0 = append(inputBytes0, bs...)
	}

	return inputBytes0
}

func ModelInferRequest(client triton.GRPCInferenceServiceClient, idx []int32, modelName string, modelVersion string) []float32 {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rawInput := Preprocess(idx)
	// fmt.Println(idx)
	// Create request input tensors
	inferInputs := []*triton.ModelInferRequest_InferInputTensor{
		&triton.ModelInferRequest_InferInputTensor{
			Name:     "IN",
			Datatype: "INT64",
			Shape:    []int64{1, 26},
		},
	}

	// Create request input output tensors
	inferOutputs := []*triton.ModelInferRequest_InferRequestedOutputTensor{
		&triton.ModelInferRequest_InferRequestedOutputTensor{
			// Name: "output",
			Name: "OUT",

		},
	}

	// Create inference request for specific model/version
	modelInferRequest := triton.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Inputs:       inferInputs,
		Outputs:      inferOutputs,
	}

	modelInferRequest.RawInputContents = append(modelInferRequest.RawInputContents, rawInput)

	// Submit inference request to server
	modelInferResponse, err := client.ModelInfer(ctx, &modelInferRequest)
	if err != nil {
		log.Fatalf("Error processing InferRequest: %v", err)
	}
	return Postprocess(modelInferResponse)
}


func readFP32(eightBytes []byte) float32 {
	buf := bytes.NewBuffer(eightBytes)
	var retval float32
	binary.Read(buf, binary.LittleEndian, &retval)
	return retval
}

// Convert output's raw bytes into int32 data (assumes Little Endian)
func Postprocess(inferResponse *triton.ModelInferResponse) []float32 {

	// outputBytes0 := inferResponse.RawOutputContents[0]
	// fmt.Println(len(outputBytes0))
	// outputData0 := make([]float32, 1)
	// for i := 0; i < 1; i++ {
	// 	outputData0[i] = readFP32(outputBytes0[i*4 : i*4+4])
	// }
	return []float32{1}
}