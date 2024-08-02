package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/lcahid/idk/internal/cluster"
	"github.com/lcahid/idk/internal/mongo"
	"github.com/lcahid/idk/internal/service"
	"github.com/lcahid/idk/internal/socketstream"
)


const (
	DefaultHTTPAddr = "localhost:9025"
	DefaultRaftAddr = "localhost:9125"
	DefaultServerAddr = "192.168.100.172:4000"
	DefaultMongoAddr =  "localhost:27017"
)

var httpAddr string
var raftAddr string
var joinAddr string
var mongoAddr string
var serverAddr string
var nodeId string

func init() {
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&serverAddr, "saddr", DefaultServerAddr, "Set WS server address")
	flag.StringVar(&mongoAddr, "maddr", DefaultMongoAddr, "Set MongoDB server address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeId, "id", "1", "Set node ID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}


func main() {
	flag.Parse()
	wg := sync.WaitGroup{}

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No storage dir specified\n")
		os.Exit(1)
	}

	raftDir := flag.Arg(0)
	if raftDir == "" {
		log.Fatal("No storage dir specified\n")
	}
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalf("Failed to create path for Raft storage: %s", err.Error())
	}

	mongoClient, err := mongocon.NewMongoClient(mongoAddr)
	if err != nil {
		log.Fatalf("Mongo client: %s", err)
		return
	}
	defer mongoClient.Close()
	
	wsClient, err := socketstream.NewWSConfig(serverAddr, mongoClient)
	if err != nil {
		log.Fatalf("WS client: %s", err)
		return
	}
	defer wsClient.Close()

	cluster := cluster.NewCluster(raftDir, raftAddr, wsClient)

	if err := cluster.Start(nodeId, joinAddr == "", &wg); err != nil {
		log.Fatalf("Failed to start cluster: %s", err)
	}

	httpService := service.NewHttpService(httpAddr, cluster)
	if err := httpService.Start(); err != nil {
		log.Fatalf("Failed to start HTTP service: %s", err)
	}

	if joinAddr != "" {
		if err := join(joinAddr, raftAddr); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	wg.Wait()
}

func join(joinAddr, raftAddr string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "nodeId": nodeId})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}