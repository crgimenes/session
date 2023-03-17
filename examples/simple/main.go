package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"crg.eti.br/go/session"
)

var (
	sc *session.Control

	//go:embed assets/*
	assets embed.FS
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	sid, sd, ok := sc.Get(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// renew session
	sc.Save(w, sid, sd)

	//////////////////////////

	index, err := assets.ReadFile("assets/index.html")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("index.html").Parse(string(index))
	if err != nil {
		log.Fatal(err)
	}

	// exec template
	err = t.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}

	// http.Redirect(w, r, "/payments", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		index, err := assets.ReadFile("assets/login.html")
		if err != nil {
			log.Fatal(err)
		}
		t, err := template.New("login.html").Parse(string(index))
		if err != nil {
			log.Fatal(err)
		}

		// exec template
		err = t.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	// login logic

	// create session
	sid, sd := sc.Create()

	// save session
	sc.Save(w, sid, sd)

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sid, _, ok := sc.Get(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// remove session
	sc.Delete(w, sid)

	http.Redirect(w, r, "/login", http.StatusFound)
}

func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	sid, sd, ok := sc.Get(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// renew session
	sc.Save(w, sid, sd)

	http.ServeFile(w, r, "payments.html")
}

func main() {
	const cookieName = "session_abcd_company"
	sc = session.New(cookieName)

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sc.RemoveExpired()
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/payments", paymentsHandler)

	s := &http.Server{
		Handler:        mux,
		Addr:           fmt.Sprintf(":%d", 8080),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Listening on port %d\n", 8080)
	log.Fatal(s.ListenAndServe())

}
