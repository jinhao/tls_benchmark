package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	//	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	log "seelog"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func do_reqs(addr string, reqs int, session_cache bool, ch chan int) {
	//config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true, ClientSessionCache: tls.NewLRUClientSessionCache(32)}
	cert2_b, _ := ioutil.ReadFile("cert2.pem")
	/*priv2_b, _ := ioutil.ReadFile("cert2.key")
	priv2, _ := x509.ParsePKCS1PrivateKey(priv2_b)

	cert := tls.Certificate{
		Certificate: [][]byte{cert2_b},
		PrivateKey:  priv2,
	}
	*/
	cert, err := x509.ParseCertificate(cert2_b)
	checkError(err)
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(cert)
	config := tls.Config{InsecureSkipVerify: false, RootCAs: rootCAs}
	if session_cache {
		config.ClientSessionCache = tls.NewLRUClientSessionCache(32)
	}

	tr := &http.Transport{
		TLSClientConfig:    &config,
		DisableCompression: true,
		DisableKeepAlives:  true,
		// TODO(jbd): Add dial timeout.
		TLSHandshakeTimeout: time.Duration(6000) * time.Millisecond,
		//Proxy:               http.ProxyURL(b.ProxyAddr),
		Dial: func(netw, addr string) (net.Conn, error) {
			lAddr, err := net.ResolveTCPAddr(netw, "172.16.30.3"+":0")
			if err != nil {
				return nil, err
			}

			rAddr, err := net.ResolveTCPAddr(netw, addr)
			if err != nil {
				return nil, err
			}
			conn, err := net.DialTCP(netw, lAddr, rAddr)
			if err != nil {
				return nil, err
			}
			deadline := time.Now().Add(35 * time.Second)
			conn.SetDeadline(deadline)
			return conn, nil
		},
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", addr, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.93 Safari/537.36")

	for i := 0; i < reqs; i++ {
		resp, err := client.Do(req)
		if err == nil {
			//size := resp.ContentLength
			//code := resp.StatusCode
			resp.Body.Close()
		} else {
			log.Warnf("client Do err:%s", err)
		}

	}
	ch <- 1
}

func vip_operate(ip_prefix string, ip_num int, up bool) {
	operate := "down"
	if up {
		operate = "up"
	}
	for i := 0; i < ip_num; i++ {
		cmd_ip_up := fmt.Sprintf("ifconfig eth0:%d %s.%d netmask 255.255.255.0 %s", i, ip_prefix, i, operate)
		cmd := exec.Command("/bin/sh", "-c", cmd_ip_up)
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("err:%s\n", err)
		} else {
			log.Warnf("vip_operate | ifconfig %s.%d %s success, out:%s", ip_prefix, i, operate, out)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log_open("./log/tls_benchmark.log")
	log.Warn("-----tls benchmark log------")

	conn := flag.Int("c", 10, "connection num")
	reqs := flag.Int("n", 100, "total num of requests(default 1000)")
	server_addr := flag.String("s", "https://172.16.91.101:443/", "server addr[default:127.0.0.1:443]")
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
	//vip_operate("172.16.30", 10, true)

	ch := make(chan int, *conn)

	for i := 0; i < *conn; i++ {
		go do_reqs(*server_addr, reqs_per_conn, *do_session_cache, ch)
	}

	for i := 0; i < *conn; i++ {
		<-ch
		log.Warnf("main | recv %d chan", i)
	}
}
