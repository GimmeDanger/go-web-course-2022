package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	md5QuotaLimit  = 1
	multihashThNum = 6
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	chs := make([]chan interface{}, len(jobs)+1)
	for i, runner := range jobs {
		wg.Add(1)
		chs[i+1] = make(chan interface{}, MaxInputDataLen)
		go func(runner job, in chan interface{}, out chan interface{}, waiter *sync.WaitGroup) {
			defer waiter.Done()
			runner(in, out)
			close(out) // this stops waiting for data
		}(runner, chs[i], chs[i+1], wg)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	quotaCh := make(chan struct{}, md5QuotaLimit)
	for data := range in {
		wg.Add(1)
		go func(data int, out chan interface{}, quotaCh chan struct{}, waiter *sync.WaitGroup) {
			defer wg.Done()
			p0 := strconv.Itoa(data)
			// run md5 worker with quota
			p1OutCh := make(chan string)
			go func(input string, outCh chan<- string, quotaCh chan struct{}) {
				quotaCh <- struct{}{}
				outCh <- DataSignerMd5(input)
				<-quotaCh
			}(p0, p1OutCh, quotaCh)
			// run crc32 worker in parallel with md5
			p2OutCh := make(chan string)
			go func(input string, outCh chan<- string) {
				outCh <- DataSignerCrc32(input)
			}(p0, p2OutCh)
			// wait for p1 finished and run another crc32 worker
			p1 := <-p1OutCh
			p3OutCh := make(chan string)
			go func(input string, outCh chan<- string) {
				outCh <- DataSignerCrc32(input)
			}(p1, p3OutCh)
			// wait all and combine result
			p2 := <-p2OutCh
			p3 := <-p3OutCh
			res := p2 + "~" + p3
			fmt.Println(p0, "SingleHash", "data", p0)
			fmt.Println(p0, "SingleHash", "md5(data)", p1)
			fmt.Println(p0, "SingleHash", "crc32(md5(data))", p2)
			fmt.Println(p0, "SingleHash", "crc32(data)", p3)
			fmt.Println(p0, "SingleHash", "result", res)
			out <- res
		}(data.(int), out, quotaCh, wg)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data string, out chan interface{}, waiter *sync.WaitGroup) {
			defer waiter.Done()
			p0 := data
			// parallel run for Crc32 hashes compute
			var outChs [multihashThNum]chan string
			for th := 0; th < multihashThNum; th++ {
				p1 := strconv.Itoa(th) + p0
				outChs[th] = make(chan string)
				go func(input string, outCh chan<- string) {
					outCh <- DataSignerCrc32(input)
				}(p1, outChs[th])
			}
			// gather result
			res := string("")
			for th := 0; th < multihashThNum; th++ {
				p2 := <-outChs[th]
				res = res + p2
				fmt.Println(p0, "MultiHash", "crc32(th+step1))", th, p2)
			}
			fmt.Println(p0, "MultiHash", "result", res)
			out <- res
		}(data.(string), out, wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	res := []string{}
	for data := range in {
		res = append(res, data.(string))
	}
	sort.Strings(res)
	out <- strings.Join(res, "_")
}
