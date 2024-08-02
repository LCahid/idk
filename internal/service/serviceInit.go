package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/lcahid/idk/internal/cluster"
)

func ServiceInit(ctx context.Context, wg *sync.WaitGroup) (*HttpService, error) {
	config := cluster.NewClusterConfig()
	cluster, err := cluster.ClusterInit(*config, ctx, wg)
	if err != nil {
		return nil, err
	}
	httpService := NewHttpService(config.HttpAddr, cluster)
	if err := httpService.Start(); err != nil {
		log.Fatalf("Failed to start HTTP service: %s", err)
	}
	if config.JoinAddr != "" {
		if err := join(config.JoinAddr, config.RaftAddr, config.NodeId); err != nil {
			log.Fatalf("failed to join node at %s: %s", config.JoinAddr, err.Error())
		}
	}
	return httpService, nil
}

func join(joinAddr, raftAddr, nodeId string) error {
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