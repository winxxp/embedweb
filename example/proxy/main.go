package main

import (
	"flag"
	"github.com/winxxp/embedweb/proxy"
	"log"
	"net/http"
)

var (
	addr = flag.String("addr", ":11111", "http server address")
)

func main() {
	http.Handle("/app/proxy/", proxy.Handler())

	log.Println("start on: ", *addr)
	http.ListenAndServe(*addr, nil)
}
