package service

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

type HttpService struct {
	addr     string
	listener net.Listener

	cluster Cluster
}

type Cluster interface{
	Join(addr, nodeId string) error
	GetLast() (string, error)
	Close() error
}

func NewHttpService(addr string, cluster Cluster) *HttpService {
	return &HttpService{
		addr:     addr,
		cluster:  cluster,
	}
}

func (h *HttpService) Start() error {
	server := http.Server{
		Handler: h,
	}

	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	h.listener = listener

	http.Handle("/", h)

	go func() {
		err := server.Serve(h.listener)
		if err != nil {
			log.Fatalf("HTTP serve error: %s", err)
		}
	}()

	return nil
}

func (h *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/join") {
		h.handleJoin(w, r)
	} else if r.URL.Path == "/last" {
		h.handleLast(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (h *HttpService) Close() error {
	if err := h.cluster.Close(); err != nil {
		return err
	}
	return h.listener.Close()
}

func (h *HttpService) Addr() net.Addr {
	return h.listener.Addr()
}

func (h *HttpService) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeId, ok := m["nodeId"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.cluster.Join(remoteAddr, nodeId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HttpService) handleLast(w http.ResponseWriter, _ *http.Request) {
	last, err := h.cluster.GetLast()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"last": last})
}