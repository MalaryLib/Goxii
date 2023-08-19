package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DescendentSetupServer struct {
	Server http.Server
	DestinationAddress *string
}

func (s *DescendentSetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "You can only use the GET_OUT method on this endpoint!", http.StatusForbidden)
	}
	dat, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var dp DestinationPayload
	json.Unmarshal(dat, &dp)

	dest := dp.DestinationAddress
	SubtleTextIndent(fmt.Sprintf("Destination Address: %s\n", dest), true)
	*s.DestinationAddress = dest
	s.Server.Close()
}

func (s *DescendentSetupServer) Listen(BindPort int, DestinationAddress *string) {
	s.DestinationAddress = DestinationAddress
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", s.handleSetup)

	s.Server = http.Server{Addr: fmt.Sprintf(":%d", BindPort)}
	s.Server.Handler = mux 

	s.Server.ListenAndServe()
}