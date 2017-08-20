package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/namsral/flag"

	pandora ".."
	"../bolt"
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
	notFoundView     *views.View
	homeView         *views.View
	factoidsView     *views.View
	factoidsEditView *views.View
)

func main() {
	c := bolt.NewClient(DBPath)
	if err := c.Open(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	Client = c

	notFoundView = views.New("base", "views/notfound.gohtml")
	homeView = views.New("base", "views/home.gohtml")
	factoidsView = views.New("base", "views/factoids.gohtml")
	factoidsEditView = views.New("base", "views/factoids_edit.gohtml")

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.HandleFunc("/", home)
	r.HandleFunc("/factoids", factoids)
	r.HandleFunc("/factoids/{id:[0-9]+}/edit", factoidsEdit)

	s := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	go s.ListenAndServe()
	defer s.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Web server is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := notFoundView.Render(w, struct{ ActivePage string }{ActivePage: ""})
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := homeView.Render(w, struct{ ActivePage string }{ActivePage: "home"})
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func factoids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	factoids, err := Client.FactoidRange(0, 50)
	if err != nil {
		return
	}
	err = factoidsView.Render(w, struct {
		ActivePage string
		Factoids   []*pandora.Factoid
	}{
		ActivePage: "factoids",
		Factoids:   factoids,
	})
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func factoidsEdit(w http.ResponseWriter, r *http.Request) {
	var f *pandora.Factoid
	vars := mux.Vars(r)
	strID, ok := vars["id"]

	if id, err := strconv.ParseUint(strID, 10, 64); !ok || err != nil {
		return
	} else if f, err = Client.Factoid(id); err != nil {
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err := factoidsEditView.Render(w, struct {
		ActivePage string
		Factoid    *pandora.Factoid
	}{
		ActivePage: "factoids",
		Factoid:    f,
	})
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
