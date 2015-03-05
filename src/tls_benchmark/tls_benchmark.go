package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	log "seelog"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func do_reqs(addr string, reqs int, session_cache bool, ch chan int) {
	cert2_b, _ := ioutil.ReadFile("cert2.pem")
	priv2_b, _ := ioutil.ReadFile("cert2.key")
	priv2, _ := x509.ParsePKCS1PrivateKey(priv2_b)

	cert := tls.Certificate{
		Certificate: [][]byte{cert2_b},
		PrivateKey:  priv2,
	}

	//config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true, ClientSessionCache: tls.NewLRUClientSessionCache(32)}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	if session_cache {
		config.ClientSessionCache = tls.NewLRUClientSessionCache(32)
	}
	for i := 0; i < reqs; i++ {
		conn, err := tls.Dial("tcp", addr, &config)
		if err != nil {
			log.Errorf("client: dial: %s", err)
			ch <- 1
			return
		}
		defer conn.Close()

		message := "GET / HTTP/1.1\r\n\r\n"
		n, err := io.WriteString(conn, message)
		if err != nil {
			log.Errorf("client: write: %s %s", err, n)
		}

		reply := make([]byte, 256)
		n, err = conn.Read(reply)
		if err != nil {
			log.Errorf("conn Read err:%s", err)
		}

	}
	ch <- 1
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log_open("./log/tls_benchmark.log")
	log.Warn("-----tls benchmark log------")

	conn := flag.Int("c", 10, "connection num")
	reqs := flag.Int("n", 100, "total num of requests(default 1000)")
	server_addr := flag.String("s", "127.0.0.1:443", "server addr[default:127.0.0.1:443]")
	do_session_cache := flag.Bool("session-cache", false, "tls session cache[default false]")
	//reqs_per_conn := 1 / 1;

	flag.Parse()
	reqs_per_conn := int(*reqs) / int(*conn)

	log.Warnf("main | cpu num:%d", runtime.NumCPU())
	if *do_session_cache {
		log.Warnf("main | session cache: true")
	} else {

		log.Warnf("main | session cache: false")
	}
	log.Warnf("main | c:%d", *conn)
	log.Warnf("main | n:%d", *reqs)
	log.Warnf("main | server_addr:%s", *server_addr)
	log.Warnf("main | reqs_p_c:%d", reqs_per_conn)

	ch := make(chan int, *conn)

	for i := 0; i < *conn; i++ {
		go do_reqs(*server_addr, reqs_per_conn, *do_session_cache, ch)
	}

	for i := 0; i < *conn; i++ {
		<-ch
		log.Warnf("main | recv %d chan", i)
	}
}
