//The runner for MonitorNode

package main

import (
	"flag"
	"fmt"
	"github.com/cmu440-F15/paxosapp/monitor"
	"log"
	"strings"
)

var (
	port       = flag.String("port", "", "port for monitor node")
	opt        = flag.String("opt", "", "option for monitor node, either crawl or pagerank")
	num        = flag.Int("num", 10, "number of webpages to crawl")
	url        = flag.String("url", "http://news.google.com", "root url to start crawling with")
	masterPort = flag.String("masterPort", "", "port for master node")
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func main() {
	flag.Parse()
	defer fmt.Println("Leaving main")
	fmt.Println("Before")
	myHostPort := "localhost:" + *port
	masterHostPort := strings.Split(*masterPort, ",")
	fmt.Println("mrunner creates monitor node on", myHostPort, "with master on ", *masterPort)

	// Create and start the Paxos Node.
	_, err := monitor.NewMonitorNode(myHostPort, masterHostPort)
	if err != nil {
		log.Fatalln("Failed to create monitor node:", err)
	}

	select {}
}
