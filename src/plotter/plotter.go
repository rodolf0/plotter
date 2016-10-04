package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Data struct {
	Cells [][]string
}

type PublishCtx struct {
	m sync.Mutex
	c map[int32]chan Data
}

func (p *PublishCtx) Get(id int32) chan Data {
	p.m.Lock()
	defer p.m.Unlock()
	if c, ok := p.c[id]; ok {
		return c
	}
	return nil
}

func (p *PublishCtx) Register() (int32, chan Data) {
	p.m.Lock()
	defer p.m.Unlock()
	max := int32(0)
	for k, _ := range p.c {
		if k > max {
			max = k
		}
	}
	max += 1
	pubchan := make(chan Data)
	p.c[max] = pubchan
	return max, pubchan
}

func (p *PublishCtx) Close(id int32) {
	p.m.Lock()
	defer p.m.Unlock()
	if c, ok := p.c[id]; ok {
		close(c)
		delete(p.c, id)
	}
}


var PC *PublishCtx
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func plotter(w http.ResponseWriter, r *http.Request) {
	// lookup the publish channel on which we're sending the data to the browser
	wsid, wsok := r.Header[http.CanonicalHeaderKey("Workspace")]
	if !wsok {
		http.NotFound(w, r)
		return
	}
	workspace, _ := strconv.Atoi(wsid[0])
	pubchan := PC.Get(int32(workspace))
	if pubchan == nil {
		http.NotFound(w, r)
		return
	}
	cells, err := csv.NewReader(r.Body).ReadAll()
	if err != nil {
		log.Printf("Failed to parse csv input: %v\n", err)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	pubchan <- Data{Cells: cells}
}

func viewer(w http.ResponseWriter, r *http.Request) {
	type tpl struct {
		Wsaddr string
	}
	GraphHtml.Execute(w, tpl{Wsaddr: "ws://" + r.Host + "/ws"})
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	// upgrade connection to websocket
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	// register a data channel to ingest data from /plot endpoint
	workspace, pubchan := PC.Register()
	defer PC.Close(workspace)
	// send initial message to set workspace name
	d := []byte(fmt.Sprintf(`{"workspace": "Workspace %v"}`, workspace))
	if err := conn.WriteMessage(websocket.TextMessage, d); err != nil {
		log.Println(err)
		return
	}
	// consume all signaling messages from the ws
	conndead := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
				close(conndead)
				return
			}
		}
	}()
	// forward each request to plot
	for {
		select {
		case <-conndead:
			return
		case data := <-pubchan:
			jdata, err := json.Marshal(data)
			if err != nil {
				log.Printf("Error encoding data: %v\n", err)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, jdata); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func main() {
	PC = &PublishCtx{c: make(map[int32]chan Data)}
	http.HandleFunc("/", viewer)
	http.HandleFunc("/ws", wshandler)
	http.HandleFunc("/plot", plotter)
	log.Fatal(http.ListenAndServe(":7272", nil))
}
