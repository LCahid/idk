package cluster

import (
	"context"
	"log"
	"sync"

	"github.com/lcahid/idk/internal/mongo"
	"github.com/lcahid/idk/internal/socketstream"
)

func ClusterInit(config ClusterConfig, ctx context.Context, wg *sync.WaitGroup) (*Cluster, error) {
	mongoClient, err := mongocon.NewMongoClient(config.mongoAddr)
	if err != nil {
		log.Fatalf("Mongo client: %s", err)
		return nil, err
	}

	wsClient, err := socketstream.NewWSConfig(config.serverAddr, mongoClient)
	if err != nil {
		log.Fatalf("WS client: %s", err)
		return nil, err
	}

	cluster := NewCluster(config.raftDir, config.RaftAddr, wsClient)

	if err := cluster.Start(config.NodeId, config.JoinAddr == "", ctx, wg); err != nil {
		log.Fatalf("Failed to start cluster: %s", err)
	}
	return cluster, nil
}