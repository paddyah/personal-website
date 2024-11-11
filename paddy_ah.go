package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles("tmpl/index.html", "tmpl/admin.html"))
var validHomePath = regexp.MustCompile("^/$")
var validAdminPath = regexp.MustCompile("^/admin/$")

//go:embed admin_username.txt
var admin_username string

//go:embed admin_password.txt
var admin_password string

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

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			if username == admin_username && password == admin_password {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func main() {
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/admin/", basicAuth(adminPageHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
