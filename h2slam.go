// The slam tool hits an HTTP/2 server with a lot of load over a single TCP connection.
//
// Run with GODEBUG=http2debug=1 or =2 to see debug info.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

var (
	host = flag.String("host", "", "hostname to hit")
	path = flag.String("path", "/image/jpeg", "path to hit on server")
)

var hc *http.Client

func main() {
	flag.Parse()
	if *host == "" {
		log.Fatalf("missing required --host flag")
	}
	c, err := tls.Dial("tcp", net.JoinHostPort(*host, "443"), &tls.Config{
		NextProtos: []string{http2.NextProtoTLS},
	})
	if err != nil {
		log.Fatal(err)
	}
	tr := &http2.Transport{}
	cc, err := tr.NewClientConn(c)
	if err != nil {
		log.Fatal(err)
	}
	hc = &http.Client{Transport: cc}

	for i := 0; i < 40; i++ {
		go loop()
	}
	select {}
}

func loop() {
	url := fmt.Sprintf("https://%s%s", *host, *path)
	for {
		res, err := hc.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if res.ProtoMajor != 2 {
			panic("not 2")
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()
		fmt.Println(len(body))
	}
}
