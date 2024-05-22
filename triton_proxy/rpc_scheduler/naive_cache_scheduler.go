package rpc_scheduler

import (
	"case_proxy/cnt"
	"case_proxy/conf"
	// "fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"
)

type NaiveCache struct {
	BaseRPCScheduler
	i atomic.Uint64
	WarmUpTime int64
	Caches map[int] *ristretto.Cache
	rnd *rand.Rand
}
func init() {
	factories[NaiveCacheScheduler] = NewNaiveCache
}
func (*NaiveCache) ToString() string {
	panic("NaiveCache")
}

func NewNaiveCache(workers []string) RPCScheduler{
	cacheSize := conf.Conf.CacheSize
	for _,v := range(workers){
		cnt.Cnter.M[v] = 0
	}
	cache := &NaiveCache{
		Caches: make(map[int]*ristretto.Cache),
		BaseRPCScheduler: BaseRPCScheduler{
			Workers: workers,
			Cnt: cnt.Cnter,
		},
		WarmUpTime: conf.Conf.WarmUpTime,
		i: atomic.Uint64{},
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	config := &ristretto.Config{
		NumCounters: 1e7,     // Num keys to track frequency of (10M).
		MaxCost:     cacheSize, // Maximum cost of cache (1GB).
		BufferItems: 64,      // Number of keys per Get buffer.
	}
	for i :=0; i< len(workers);i++{
		rcache, err := ristretto.NewCache(config)
		if err != nil {
			return nil
		}
		cache.Caches[i] = rcache
	}
	return cache

}


func (n *NaiveCache) Schedule(_ string, r Req) (string, error) {
	keys := r.Keys
	index := n.i.Add(1)
	if index < uint64(n.WarmUpTime){
		cache := n.Caches[int(index)]
		// fmt.Println(cache.Len())
		for _, a := range keys{
			for _ ,aa := range a {
				cache.Set(aa,aa,1)
			}
			
		}
		// fmt.Println(cache.Len())
		host := n.Workers[index%uint64(len(n.Workers))]
		return host, nil
	}




	var hostIndex = make([]int,0)
	var maxCnt = -1



	for i :=0; i< len(n.Workers);i++{
			cache := n.Caches[i]
			cnt := 0
			for _, a := range keys{
				for jj:=0; jj<1;jj++{
						for _ ,aa := range a {
					if _,ok := cache.Get(aa); ok{
						cnt  = cnt + 1
					}
				}
				}

			}
			if cnt == maxCnt{
				hostIndex = append(hostIndex, i)
			}
			if cnt > maxCnt{
				maxCnt  = cnt
				hostIndex = make([]int,0)
				hostIndex = append(hostIndex, i)
			}
	}
	if len(hostIndex) == 0{
		index := n.rnd.Intn(len(n.Workers))
		return n.Workers[index],nil
	}
	curIndex := hostIndex[(rand.Intn(len(hostIndex)))]
	cache := n.Caches[curIndex]
	for _, a := range keys{
		for _ ,aa := range a {
			cache.Set(aa,aa,1)
		}
	}
	return n.Workers[curIndex],nil

}
