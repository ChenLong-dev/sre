package test

import (
	"net/http"
	"net/http/pprof"
)

const (
	// 默认单测pprof地址
	DefaultUnitPProfAddr = "localhost:8089"
)

// 后台运行pprof
func RunPProfInBackground(addr string) {
	if addr == "" {
		addr = DefaultUnitPProfAddr
	}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		_ = http.ListenAndServe(addr, mux)
	}()
}
