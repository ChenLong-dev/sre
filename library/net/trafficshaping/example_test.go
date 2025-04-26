package trafficshaping

import (
	"fmt"
	"sync"
	"time"
)

func ExampleNewPipeline_qps_reject() {
	p, err := NewPipeline([]*Rule{
		{
			Type:            QPS,
			ControlBehavior: Reject,
			Limit:           1,
		},
	})
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				_ = p.Do(func() {
					fmt.Println(i, time.Now().Format("15:04:05.000"), "passed")
					time.Sleep(500 * time.Millisecond)
				})
			}
		}(i)
	}
	wg.Wait()
}

func ExampleNewPipeline_concurrency_reject() {
	p, err := NewPipeline([]*Rule{
		{
			Type:            Concurrency,
			ControlBehavior: Reject,
			Limit:           1,
		},
	})
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				_ = p.Do(func() {
					fmt.Println(i, time.Now().Format("15:04:05.000"), "passed")
					time.Sleep(500 * time.Millisecond)
					fmt.Println(i, time.Now().Format("15:04:05.000"), "finished")
				})
			}
		}(i)
	}
	wg.Wait()
}

func ExampleNewPipeline_waiting() {
	p, err := NewPipeline([]*Rule{
		{
			Type:            QPS,
			ControlBehavior: Waiting,
			Limit:           1,
			MaxWaitingTime:  time.Second * 3,
		},
	})
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				_ = p.Do(func() {
					fmt.Println(i, time.Now().Format("15:04:05.000"), "passed")
					time.Sleep(500 * time.Millisecond)
				})
			}
		}(i)
	}
	wg.Wait()
}

func ExampleNewPipeline_multi() {
	p, err := NewPipeline([]*Rule{
		{
			Type:            QPS,
			ControlBehavior: Reject,
			Limit:           10,
		},
		{
			Type:            Concurrency,
			ControlBehavior: Reject,
			Limit:           1,
		},
	})
	if err != nil {
		panic(err)
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				_ = p.Do(func() {
					fmt.Println(i, time.Now().Format("15:04:05.000"), "passed")
					time.Sleep(500 * time.Millisecond)
				})
			}
		}(i)
	}
	wg.Wait()
}
