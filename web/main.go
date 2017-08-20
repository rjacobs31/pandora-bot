package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	r.HandleFunc("/factoids/{id:[0-9]+}/edit", factoidsEdit).Methods("GET")
	r.HandleFunc("/factoids/{id:[0-9]+}/edit", factoidsEditPost).Methods("POST")

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
	var (
		err error
		f   *pandora.Factoid
		id  uint64
	)
	vars := mux.Vars(r)
	if strID, ok := vars["id"]; ok {
		id, err = strconv.ParseUint(strID, 10, 64)
		if err == nil {
			f, _ = Client.Factoid(id)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	err = factoidsEditView.Render(w, struct {
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

func factoidsEditPost(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		f   *pandora.Factoid
		id  uint64
	)
	vars := mux.Vars(r)
	if strID, ok := vars["id"]; ok {
		id, err = strconv.ParseUint(strID, 10, 64)
		if err == nil {
			f, _ = Client.Factoid(id)
		}
	}

	if f != nil {
		r.ParseMultipartForm(10 * 1024)

		if trig := r.PostFormValue("trigger"); trig != "" && trig != f.Trigger {
			f.Trigger = trig
			f.DateEdited = time.Now()
		}

		i := 0
		for {
			v := r.PostFormValue("res_" + strconv.Itoa(i) + "_response")
			if v == "" || i >= len(f.Responses) {
				break
			} else if v != f.Responses[i].Response {
				f.Responses[i].Response = v
				f.Responses[i].DateEdited = time.Now()
			}
			i++
		}

		err = Client.PutFactoid(id, f)
	}

	if err == nil {
		http.Redirect(w, r, "/factoids", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	factoidsEditView.Render(w, struct {
		ActivePage string
		Error      error
		Factoid    *pandora.Factoid
	}{
		ActivePage: "factoids",
		Error:      err,
		Factoid:    f,
	})
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
