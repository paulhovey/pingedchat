package main

import (
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"pingedchat/mailWebhooks"
	"pingedchat/pcDatabase"
	"time"
)

var (
	// for logging
	TRACE = log.New(os.Stdout, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

const (
	PORT       = ":443"
	PRIV_KEY   = "./PC_private_unencrypted.key"
	PUBLIC_KEY = "./PC_public.pem"
)

func redir(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "https://FILL_ME.com", http.StatusMovedPermanently)
}

func redirHTTP() {
	if err := http.ListenAndServe(":80", http.HandlerFunc(redir)); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func textWebhooksLoop() {
	if err := http.ListenAndServe(":8118", http.HandlerFunc(mailWebhooks.MandrillTextingPost)); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func emailWebhooksLoop() {
	if err := http.ListenAndServe(":8119", http.HandlerFunc(mailWebhooks.MandrillEmailPost)); err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func sockjsServerLoop() {
	// start sockjs server
	sockjsOptions := sockjs.DefaultOptions
	sockjsOptions.SockJSURL = "https://FILL_ME.com/vendor/plugins/sockjs.min.js"
	http.Handle("/ws/", sockjs.NewHandler("/ws", sockjsOptions, sockHandler))
	http.Handle("/", http.FileServer(http.Dir("web/")))
	if err := http.ListenAndServeTLS(PORT, PUBLIC_KEY, PRIV_KEY, nil); err != nil {
		ERROR.Printf("ListenAndServe:", err)
	}
}

func main() {
	// to disable trace messages
	log.SetOutput(ioutil.Discard)
	// http.HandleFunc("/ws", wsHandler)
	// start ADM token gette
	go pcDatabase.Protect(pcDatabase.ADMinit)
	go pcDatabase.Protect(pcDatabase.StartMessagesTicker)
	// start HTTP redirect
	go pcDatabase.Protect(redirHTTP)
	go pcDatabase.Protect(textWebhooksLoop)
	go pcDatabase.Protect(emailWebhooksLoop)
	go pcDatabase.Protect(sockjsServerLoop)

	// go exits when the main function is done, so I guess we'll just keep this party going forever!
	for {
		time.Sleep(time.Hour * 96) // 4 days
	}
}
