package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Location struct {
	// can also add an id for multiple locations
	// this complicates things
	mu     sync.Mutex
	Orders map[string]*Order `json:"orders,omitempty"`
}

type Order struct {
	OrderID string    `json:"order_id"`
	History []*Coords `json:"history,omitempty"`
}

type Coords struct {
	Lat float32 `json:"lat"`
	Lng float32 `json:"lng"`
}

func writeAndLog(w http.ResponseWriter, out []byte) {
	log.Println(string(out))
	fmt.Fprintf(w, string(out))
}

func (l *Location) GetAll(w http.ResponseWriter, req *http.Request) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, _ := json.MarshalIndent(l, "", "  ")
	writeAndLog(w, b)
}

func (l *Location) PlaceOrder(w http.ResponseWriter, req *http.Request) {
	log.Println("hi")
	var x Order
	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
	err := json.NewDecoder(req.Body).Decode(&x)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.Orders[x.OrderID]; ok {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("Duplicate order: " + x.OrderID + "\n"))
		return
	} else {
		l.Orders[x.OrderID] = &x
		b, _ := json.MarshalIndent(&x, " ", "  ")
		writeAndLog(w, b)
	}
}

func (l *Location) DeleteOrder(w http.ResponseWriter, req *http.Request) {
	var x Order
	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
	err := json.NewDecoder(req.Body).Decode(&x)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.Orders[x.OrderID]; ok {
		delete(l.Orders, x.OrderID)
		fmt.Fprintf(w, "Removed order: %v", x.OrderID)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(x.OrderID + " does not exist.\n"))
		return
	}
}

func (l *Location) AddLocation(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	oid := vars["order_id"]

	var x Coords
	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)
	err := json.NewDecoder(req.Body).Decode(&x)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.Orders[oid]; ok {
		l.Orders[oid].History = append(l.Orders[oid].History, &x)
		fmt.Fprintf(w, "Added coordinates to order %s", oid)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(oid + " does not exist.\n"))
		return
	}
}

func (l *Location) GetHistory(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	order_id := vars["order_id"]
	max_param := vars["max"]
	var max int = 0
	var err error

	if max_param != "" {
		max, err = strconv.Atoi(max_param)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid parameter max: " + max_param + "\n"))
			return
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if val, ok := l.Orders[order_id]; ok {
		x := *val
		if max != 0 {
			if max < len(x.History) && max != 0 {
				x.History = x.History[len(x.History)-max:]
			}
		}
		b, _ := json.MarshalIndent(x, " ", "  ")
		fmt.Fprintf(w, string(b))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(order_id + " does not exist.\n"))
		return
	}
}

func main() {
	listen_on, ok := os.LookupEnv("LHPORT")
	if !ok {
		listen_on = "8080"
	}

	r := mux.NewRouter()

	loc := &Location{}
	loc.Orders = make(map[string]*Order, 0)

	r.HandleFunc("/location", loc.PlaceOrder).Methods("POST")

	r.HandleFunc("/location/{order_id}", loc.DeleteOrder).Methods("DELETE")
	r.HandleFunc("/location/{order_id}", loc.AddLocation).Methods("PUT")
	r.HandleFunc("/location/{order_id}", loc.GetHistory).Methods("GET").Queries("max", "{max}")

	r.HandleFunc("/location/{order_id}", loc.GetHistory).Methods("GET")

	r.HandleFunc("/", loc.GetAll).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "localhost:" + listen_on,
	}

	log.Fatal(srv.ListenAndServe())
}
