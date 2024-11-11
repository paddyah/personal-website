package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles("tmpl/index.html", "tmpl/admin.html"))
var validHomePath = regexp.MustCompile("^/$")
var validAdminPath = regexp.MustCompile("^/admin/$")

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	m := validHomePath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "index")
}

func adminPageHandler(w http.ResponseWriter, r *http.Request) {
	m := validAdminPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "admin")
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/admin/", adminPageHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
