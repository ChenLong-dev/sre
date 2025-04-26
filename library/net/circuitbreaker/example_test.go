package circuitbreaker

import "context"

func ExampleNewThresholdBreaker() {
	breaker := NewThresholdBreaker(10)
	err := breaker.Call(context.Background(), func() error {
		return nil
	}, 0)
	if err != nil {
		return
	}
}

func ExampleNewBreakerGroup() {
	vipBreaker := NewThresholdBreaker(10)
	group := NewBreakerGroup()
	group.Add("vip", vipBreaker)

	breaker := group.Get("vip")
	if breaker == nil {
		return
	}
	err := breaker.Call(context.Background(), func() error {
		return nil
	}, 0)
	if err != nil {
		return
	}
}
