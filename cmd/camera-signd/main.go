package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"darlinggo.co/trout"
)

type Server struct {
	Statuses map[string]Status `json:"statuses"`
	statusMu sync.RWMutex
	plug     Hs1xxPlug
}

type Status struct {
	CameraOn bool      `json:"cameraOn"`
	LastSync time.Time `json:"lastSync"`
}

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("Usage: camera-signd {OUTLET IP}")
		os.Exit(1)
	}

	s := &Server{
		Statuses: map[string]Status{},
		plug:     Hs1xxPlug{IPAddress: os.Args[1]},
	}
	go s.syncSignLoop(ctx)

	var router trout.Router
	router.Endpoint("/status").Methods(http.MethodGet).Handler(http.HandlerFunc(s.getStatusHandler))
	router.Endpoint("/status/{mac}").Methods(http.MethodPatch).Handler(http.HandlerFunc(s.patchStatusHandler))
	router.Endpoint("/status").Methods(http.MethodDelete).Handler(http.HandlerFunc(s.deleteStatusHandler))

	http.Handle("/", router)
	err := http.ListenAndServe(":9988", nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (s *Server) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	b, err := json.Marshal(s.Statuses)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Server) patchStatusHandler(w http.ResponseWriter, r *http.Request) {
	mac := trout.RequestVars(r).Get("mac")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	defer r.Body.Close()
	var status Status
	err = json.Unmarshal(b, &status)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	status.LastSync = time.Now()
	change := true
	s.statusMu.Lock()
	before, ok := s.Statuses[mac]
	if ok {
		change = before.CameraOn != status.CameraOn
	}
	s.Statuses[mac] = status
	s.statusMu.Unlock()
	if change {
		err = s.syncSign(r.Context())
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteStatusHandler(w http.ResponseWriter, r *http.Request) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	s.Statuses = map[string]Status{}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) syncSignLoop(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := s.syncSign(ctx)
			if err != nil {
				log.Println("error syncing sign:", err.Error())
				continue
			}
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				log.Println("context finished:", err.Error())
			}
			return
		}
	}
}

func (s *Server) syncSign(ctx context.Context) error {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()

	desiredOn := false
	for _, status := range s.Statuses {
		if time.Since(status.LastSync) > time.Minute*15 {
			continue
		}
		if status.CameraOn {
			desiredOn = true
			break
		}
	}
	if desiredOn {
		err := s.plug.TurnOn()
		if err != nil {
			return fmt.Errorf("error turning plug on: %w", err)
		}
		return nil
	}
	err := s.plug.TurnOff()
	if err != nil {
		return fmt.Errorf("error turning plug off: %w", err)
	}
	return nil
}
