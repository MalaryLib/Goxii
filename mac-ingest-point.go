package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type MacServer struct {
	MacAllowedMap map[string]bool
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
			fmt.Printf("Allowed: %s\n", Mac)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}


func (ms *MacServer) StartServer(BindPort int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.HandleRoot)
	mux.HandleFunc("/ingest-mac", ms.ingestMacWithPass)

	err := http.ListenAndServe(fmt.Sprintf(":%d", BindPort), mux)
	if err != nil {
		panic(err)
	}
}