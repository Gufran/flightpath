package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"log"
	"net/http"
	"strings"
	"sync"
)

type DebugServer struct {
	mx    *sync.Mutex
	state cache.SnapshotCache
	node  string
}

func StartDebugServer(port int, node string, c cache.SnapshotCache) {
	s := &DebugServer{
		mx:    &sync.Mutex{},
		state: c,
		node:  node,
	}

	go s.ListenAndServe(port)
}

func (d *DebugServer) ListenAndServe(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.dump)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("starting debug server on %s", addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("debug http server crashed with error. %s", err)
	}
}

func (d *DebugServer) sendResp(resp http.ResponseWriter, kind string) {
	snap, err := d.state.GetSnapshot(d.node)
	if err != nil {
		log.Printf("failed to retrieve snapshot from XDS cache. %s", err)
		http.Error(resp, "failed to retrieve state", http.StatusInternalServerError)
		return
	}

	var data interface{}

	switch kind {
	case "clusters":
		data = snap.Clusters
	case "endpoints":
		data = snap.Endpoints
	case "listeners":
		data = snap.Listeners
	case "routes":
		data = snap.Routes
	case "", "all":
		data = snap
	default:
		http.Error(resp, "configuration type "+kind+" is not understood. only 'clusters', 'endpoints', 'listeners' and 'routes' are supported", http.StatusNotFound)
		return
	}

	enc := json.NewEncoder(resp)
	enc.SetIndent("", "  ")
	err = enc.Encode(data)
	if err != nil {
		log.Printf("Debug server failed to send response. %s", err)
	}
}

func (d *DebugServer) dump(resp http.ResponseWriter, req *http.Request) {
	kind := strings.TrimLeft(req.URL.Path, "/")
	d.sendResp(resp, kind)
}
