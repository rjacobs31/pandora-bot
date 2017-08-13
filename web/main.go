package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/namsral/flag"

	pandora ".."
	bolt "../bolt"
)

// Variables used for command line parameters
var (
	DBPath string
	Client pandora.DataClient
)

func init() {
	flag.String(flag.DefaultConfigFlagname, "", "path to config file")
	flag.StringVar(&DBPath, "pandora_db", "pandora.boltdb", "BoltDB database file path")
	flag.Parse()
}

func main() {
	c := bolt.NewClient(DBPath)
	if err := c.Open(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	Client = c

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	http.ListenAndServe(":3000", r)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Web server is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "Sorry, but I don't know that page. :(")
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Welcome to my website!</h1>")
}
