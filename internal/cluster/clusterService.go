package cluster

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

type Cluster struct {
	raftDir  string
	raftAddr string

	client wsClient

	raft *raft.Raft 

	logger *log.Logger
}

type wsClient interface {
	Start(ctx context.Context, wg *sync.WaitGroup)
	Apply(l *raft.Log) interface{}
	Snapshot() (raft.FSMSnapshot, error)
	Restore(io.ReadCloser) error
}

func NewCluster(raftDir, raftAddr string, client wsClient) *Cluster {
	return &Cluster{
		raftDir:  raftDir,
		raftAddr: raftAddr,
		logger: log.New(os.Stdout, "[store] ", log.LstdFlags),
		client: client,
	}
}

func (c *Cluster) Start(localId string, first bool, ctx context.Context, wg *sync.WaitGroup) error {
	config := raft.DefaultConfig()
	if localId == "" {
		config.LocalID = raft.ServerID(1)
	} else {
		config.LocalID = raft.ServerID(localId)
	}

	addr, err := net.ResolveTCPAddr("tcp", c.raftAddr)
	if err != nil {
		return err
	}
	
	trans, err := raft.NewTCPTransport(c.raftAddr, addr, 5, 5*time.Second, os.Stdout)
	if err != nil {
		return err
	}
	snaps, err := raft.NewFileSnapshotStore(c.raftDir, 3, os.Stderr)
	if err != nil {
		return err
	}

	logs := raft.NewInmemStore()
	stable := raft.NewInmemStore()

	ra, err := raft.NewRaft(config, c.client, logs, stable, snaps, trans)
	if err != nil {
		return err
	}
	c.raft = ra

	if first{
		c.logger.Println("bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: trans.LocalAddr(),
				},
			},
		}
		c.raft.BootstrapCluster(configuration)
	}
	wg.Add(1)
	go c.clientStart(ctx, wg)

	return nil
}


func (c *Cluster) Close() error {
	err := c.raft.Shutdown().Error()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cluster) clientStart(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			if c.raft.State() == raft.Leader{
				c.client.Start(ctx, wg)
				return
			}
		}
	}
}


func (c *Cluster) Join(addr, nodeId string) error {
	c.logger.Printf("join request from remote node %s",  addr)
	configFuture := c.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == raft.ServerAddress(addr) || srv.ID == raft.ServerID(nodeId) {
			c.logger.Printf("server %s already member of cluster", addr)
			return fmt.Errorf("server %s already member of cluster", addr)
		}
	}

	f := c.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(addr), 0, 0)
	if err := f.Error(); err != nil {
		return err
	}
	c.logger.Printf("node %s at %s joined successfully", raft.ServerID(nodeId), addr)
	return nil
}

func (c *Cluster) GetLast() (string, error) {
	configFuture := c.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return "", err
	}
	return string(len(configFuture.Configuration().Servers) + 1), nil
}

