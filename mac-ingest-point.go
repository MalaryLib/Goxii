package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type MacServer struct {
	MacAllowedMap map[string]bool
	ProxyDescendents []string
	Dp DestinationPayload
}

type DestinationPayload struct {
	DestinationAddress string
}

func (ms *MacServer) ReachOutToProxy() {
	dat, _ := json.Marshal(ms.Dp)
	url := fmt.Sprintf("http://%s/setup", ms.ProxyDescendents[0])
	http.Post(url, "application/json", bytes.NewReader(dat))
}

func (ms *MacServer) HandleProxyChild(w http.ResponseWriter, r *http.Request) {
	if len(ms.ProxyDescendents) == 0 {
		defer r.Body.Close()
	}
	dat, err := json.Marshal(ms.Dp)
	url := fmt.Sprintf("http://%s/setup", ms.ProxyDescendents[0])
	resp, err := http.Post(url, "application/json", bytes.NewReader(dat))
	check(err)
	defer resp.Body.Close()
}

func (ms *MacServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	dat, err := os.ReadFile("./templates/MacPage.html")
	if err != nil {
		dat = []byte("Website!")
	}
	io.WriteString(w, string(dat))
}

func (ms*MacServer) ingestMacWithPass(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		Mac := strings.ToUpper((r.PostFormValue("Mac")))
		SudoPassword := (r.PostFormValue("SudoPassword"))

		h := sha256.New()
		h.Write([]byte(SudoPassword))
		if fmt.Sprintf("%x", h.Sum(nil)) == "0147dc0802060629357a33b1e98cc1c8b207d743eda9d03715e987c325d3d335"	{
			ms.MacAllowedMap[Mac] = true
			WarningText(fmt.Sprintf("Allowed: %s\n", Mac))
		}

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}


func (ms *MacServer) StartServer(BindAddres string, BindPort int) {
	if len(ms.ProxyDescendents) > 0 {
		ms.ReachOutToProxy()
	}
	SubtleText("Activating the Mac Ingestion Server!\n")
	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.HandleRoot)
	mux.HandleFunc("/ingest-mac", ms.ingestMacWithPass)
	mux.HandleFunc("/activate-chain", ms.HandleProxyChild)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", BindAddres, BindPort), mux)
	if err != nil {
		panic(err)
	}
}