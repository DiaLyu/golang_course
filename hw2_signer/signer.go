package main

import (
	"sync"
	"fmt"
	"strconv"
	"strings"
	"sort"
)

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup

	in := make(chan interface{})

	for _, jb := range jobs {
		out := make(chan interface{})
		wg.Add(1)

		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(jb, in, out)

		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	var md5Mutex sync.Mutex
	var wg sync.WaitGroup

	for val := range in {
		data := fmt.Sprintf("%v", val)
		wg.Add(1)

		go func(data string) {
			defer wg.Done()

			md5Mutex.Lock()
			md5 := DataSignerMd5(data)
			md5Mutex.Unlock()

			var crcData, crcMd5 string
			var innerWg sync.WaitGroup
			innerWg.Add(2)

			go func() {
				defer innerWg.Done()
				crcData = DataSignerCrc32(data)
			}()

			go func() {
				defer innerWg.Done()
				crcMd5 = DataSignerCrc32(md5)
			}()

			innerWg.Wait()
			out <- crcData + "~" + crcMd5
		}(data)
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup

	for val := range in {
		data := val.(string)
		wg.Add(1)

		go func(data string) {
			defer wg.Done()

			results := make([]string, 6)
			var innerWg sync.WaitGroup
			innerWg.Add(6)

			for i := 0; i < 6; i++ {
				i := i
				go func() {
					defer innerWg.Done()
					results[i] = DataSignerCrc32(strconv.Itoa(i) + data)
				}()
			}

			innerWg.Wait()
			out <- strings.Join(results, "")
		}(data)
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string

	for val := range in {
		results = append(results, val.(string))
	}

	sort.Strings(results)
	out <- strings.Join(results, "_")
}

