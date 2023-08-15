package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type MacIngestionPoint struct {
	MacAllowedMap map[string]bool
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	dat, err := os.ReadFile("./MacPage.html")
	if err != nil {
		log.Panicln(err)
		dat = []byte("Website!")
	}
	io.WriteString(w, string(dat))
}

func (m *MacIngestionPoint) ingestMacWithPass(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		Mac := strings.ToUpper((r.PostFormValue("Mac")))
		SudoPassword := (r.PostFormValue("SudoPassword"))

		h := sha256.New()
		h.Write([]byte(SudoPassword))
		if fmt.Sprintf("%x", h.Sum(nil)) == "0147dc0802060629357a33b1e98cc1c8b207d743eda9d03715e987c325d3d335"	{
			m.MacAllowedMap[Mac] = true
		}

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (m *MacIngestionPoint) StartServer() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/ingest-mac", m.ingestMacWithPass)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
}