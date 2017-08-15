package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/namsral/flag"

	pandora ".."
	"./views"
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

var (
	homeView     *views.View
	notFoundView *views.View
)

func main() {
	// c := bolt.NewClient(DBPath)
	// if err := c.Open(); err != nil {
	// 	log.Fatal(err)
	// }
	// defer c.Close()
	// Client = c
	notFoundView = views.New("base", "views/notfound.gohtml")
	homeView = views.New("base", "views/home.gohtml")

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
	err := notFoundView.Render(w, nil)
	if err != nil {
		fmt.Println("Error: %v", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := homeView.Render(w, nil)
	if err != nil {
		fmt.Println("Error: %v", err)
	}
}
