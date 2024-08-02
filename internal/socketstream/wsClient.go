package socketstream

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/raft"
)

type wsClient struct {
	addr *string
	client *websocket.Conn
	interrupt chan os.Signal
	db dbClient
}

type dbClient interface {
	InsertData(map[string]string) error
	Close() error
}

func NewWSConfig(addr string, db dbClient) (*wsClient, error) {
	var u url.URL = url.URL{Scheme: "ws", Host: addr, Path: "/full-stream"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
		return nil, err
	}
	log.Printf("connecting to %s", u.String())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	return &wsClient{
		addr: &addr,
		client: c,
		interrupt: interrupt,
		db: db,
	}, nil
}

func (ws *wsClient) Start(wg *sync.WaitGroup) {
	go func() {
		defer ws.client.Close()
		wg.Add(1)
		defer wg.Done()
		for i := 0; true; i++{
			select {
			case <- ws.interrupt:
				log.Println("interrupt")
				err := ws.Close()
				if err != nil {
					log.Println("error closing ws connection")
				}
				return
			default:
				_, message, err := ws.client.ReadMessage()
				if err != nil {
					log.Println("read error:", err)
					return
				}
				decodedMessage := string(message)
				time.Sleep(5 * time.Second)
				log.Println("Record added to mongo")
				ws.db.InsertData(map[string]string{fmt.Sprint(i): decodedMessage})
			}
		}
	}()
}

func (ws *wsClient) Close() error{
	// err := ws.client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// if err != nil {
	// 	log.Println("write close error:", err)
	// 	return err
	// }
	err := ws.client.Close()
	if err != nil {
		log.Println("close error:", err)
		return err
	}
	return nil
}


func (c *wsClient) Apply(l *raft.Log) interface{} {
	return nil
}


func (c *wsClient) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}


func (c *wsClient) Restore(io.ReadCloser) error {
	return nil
}
