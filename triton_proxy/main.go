package main

import (
	"case_proxy/conf"
	"case_proxy/load"
	"log"

	// "encoding/csv"
	"case_proxy/rpc_proxy"
	"encoding/json"
	"net/http"
	// "strconv"
	"sync"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/buaazp/fasthttprouter"
	// "github.com/gorilla/mux"
)

var config *conf.Config
var err error
var rpcProxy *rpc_proxy.PRCProxy
var StatisticsMap map[string]*DataStatistics = make(map[string]*DataStatistics)
var StatisticsList = make([]*rpc_proxy.Statistics, 0)

type DataStatistics struct {
	mu   sync.Mutex
	Data []*rpc_proxy.Statistics
}

var L sync.Mutex

func dataReceiver(ch chan rpc_proxy.Statistics) {
	for data := range ch {
		L.Lock()
		StatisticsList = append(StatisticsList, &data)
		L.Unlock()
	}
}

type mh struct{
	rpcProxy *rpc_proxy.PRCProxy
}

func fasthttpInferRequest(httpContext *fasthttp.RequestCtx) {


	// 锁定整个StatisticsMap，防止并发修改
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	if err := json.NewEncoder(httpContext).Encode(StatisticsList); err != nil {
		httpContext.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}



func (m *mh) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 设置HTTP响应的Content-Type
	w.Header().Set("Content-Type", "application/json")

	// 锁定整个StatisticsMap，防止并发修改
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	// 将StatisticsMap转换为JSON格式
	statsJSON, err := json.MarshalIndent(StatisticsList, "", " ")
	if err != nil {
		http.Error(w, "Failed to marshal statistics", http.StatusInternalServerError)
		return
	}

	// 发送JSON数据给客户端
	_, err = w.Write(statsJSON)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func main() {
	config, err = conf.ReadConfig("conf/config.yaml")
	if err != nil {
		log.Fatalf("read config error: %s", err)
	}
	for _, v := range config.Location[0].ProxyPass {
		StatisticsMap[v] = &DataStatistics{
			Data: make([]*rpc_proxy.Statistics, 0),
		}
	}
	go dataReceiver(rpc_proxy.Datach)
	load.LoadPartition(config.PartitionFile)
	fasthttpRouter := fasthttprouter.New()
	// router := mux.NewRouter()
	for _, l := range config.Location {
		rpcProxy, err = rpc_proxy.NewPRCProxy(l.ProxyPass, l.BalanceMode)
		if err != nil {
			log.Fatalf("create proxy error: %s", err)
		}
		// start health check

		fasthttpRouter.POST("/predict", rpcProxy.FasthttpInferRequest)
	}

	fasthttpRouter.GET("/st",fasthttpInferRequest)
	// name := "World"
	// svr := http.Server{
	// 	Addr:    ":" + strconv.Itoa(config.Port),
	// 	Handler: router,
	// }
		config.Print()
	err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", config.Port), fasthttpRouter.Handler)
	if err != nil {
		fmt.Println(err)
		return
	}
}
