package main

import (
	"log"
	"strings"
	"fmt"
	"net/http"
	"html/template"

	"github.com/gorilla/mux"
	"github.com/NorgannasAddOns/go-uuid"
)

type poll struct {
	ID string
	Name string
	Description string
	Choices map[string]*choice
}

type choice struct {
	ID string
	Name string
	Tally int
}

const pollTemplate = `<form method="post"><h1>{{.Name}}</h1><p>{{.Description}}</p><ul>{{ range $id, $choice := .Choices }}<li><input type="radio" name="choice" value="{{$choice.ID}}">{{$choice.Name}}</li>{{ end }}</il><input type="submit" name="submit" value="submit"></form>`
const pollResultsTemplate = `<h1>{{.Name}}</h1><p>{{.Description}}</p><ul>{{ range $id, $choice := .Choices }}<li>{{$choice.Name}} = {{$choice.Tally}}</li>{{ end }}</il>`

var polls map[string]*poll

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Rock The Vote</h1>")
}

func getPollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, ok := polls[vars["id"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	t := template.Must(template.New("pollTemplate").Parse(pollTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, p)
}

func postPollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, ok := polls[vars["id"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	cid := r.FormValue("choice")

	c, ok := p.Choices[cid]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	c.Tally = c.Tally + 1

	http.Redirect(w, r, "/r/" + vars["id"], http.StatusSeeOther)
}

func getResultsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, ok := polls[vars["id"]]

	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.Must(template.New("pollResultsTemplate").Parse(pollResultsTemplate))
	t.Execute(w, p)
}

func getNewPollHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<form method="post"><input type="text" name="title" placeholder="title"><br/><input type="text" name="description" placeholder="description"><br/><textarea name="choices" placeholder="choices, one per line"></textarea><input type="submit" name="submit" value="submit"></form>`)
}

func postNewPollHandler(w http.ResponseWriter, r *http.Request) {
	p := &poll{
		ID : uuid.New("P"),
		Name : r.FormValue("title"),
		Description : r.FormValue("description"),
		Choices: make(map[string]*choice),
	}
	for _, cs := range strings.Split(r.FormValue("choices"), "\n") {
		c := &choice{
			ID: uuid.New("C"),
			Name: cs,
			Tally: 0,
		}
		p.Choices[c.ID] = c
	}

	polls[p.ID] = p

	http.Redirect(w, r, "/p/" + p.ID, http.StatusSeeOther)
}

func main() {
	log.Println("hi")

	polls = make(map[string]*poll)

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/p/{id}", getPollHandler).Methods("GET")
	r.HandleFunc("/p/{id}", postPollHandler).Methods("POST")
	r.HandleFunc("/r/{id}", getResultsHandler).Methods("GET")
	r.HandleFunc("/n", getNewPollHandler).Methods("GET")
	r.HandleFunc("/n", postNewPollHandler).Methods("POST")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}