package cluster

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	DefaultHTTPAddr = "localhost:9025"
	DefaultRaftAddr = "localhost:9125"
	DefaultServerAddr = "10.20.1.232:4000"
	DefaultMongoAddr =  "localhost:27017"
)


type ClusterConfig struct {
	HttpAddr   string
	RaftAddr   string
	JoinAddr   string
	mongoAddr  string
	serverAddr string
	NodeId     string
	raftDir    string
}

func NewClusterConfig() *ClusterConfig {
	var httpAddr string
	var raftAddr string
	var joinAddr string
	var mongoAddr string
	var serverAddr string
	var nodeId string
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
	flag.Parse()

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

	return &ClusterConfig{
		HttpAddr:   httpAddr,
		RaftAddr:   raftAddr,
		JoinAddr:   joinAddr,
		mongoAddr:  mongoAddr,
		serverAddr: serverAddr,
		NodeId:     nodeId,
		raftDir:    raftDir,
	}
}