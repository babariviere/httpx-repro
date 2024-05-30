package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/foxcpp/go-mockdns"
	"github.com/projectdiscovery/goflags"
	customport "github.com/projectdiscovery/httpx/common/customports"
	"github.com/projectdiscovery/httpx/runner"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	addr, _ := ts.Listener.Addr().(*net.TCPAddr)

	dns, _ := mockdns.NewServer(map[string]mockdns.Zone{
		"test.com.": {
			A: []string{"127.0.0.1"},
		},
	}, false)
	targetResolver := &net.Resolver{}
	dns.PatchNet(targetResolver)

	ports := customport.CustomPorts{}
	ports.Set(strconv.Itoa(addr.Port))
	fmt.Println("Target port:", addr.Port)
	fmt.Println("DNS addr:", dns.LocalAddr())
	options := runner.Options{
		Methods:         "GET",
		InputTargetHost: goflags.StringSlice{"test.com"},
		CustomPorts:     ports,
		Timeout:         5,
		StatusCode:      true,
		ContentLength:   true,
		Threads:         10,
		TechDetect:      true,
		Silent:          true,
		NoFallback:      true,
		ExtractTitle:    true,
		OnResult: func(hr runner.Result) {
			fmt.Println("SHOULD BE PRINTED:", hr)
		},

		Resolvers: []string{dns.LocalAddr().String()},
	}

	if err := options.ValidateOptions(); err != nil {
		panic("invalid options")
	}

	httpxRunner, err := runner.New(&options)
	if err != nil {
		panic("runner failed")
	}
	defer httpxRunner.Close()
	httpxRunner.RunEnumeration()
}
