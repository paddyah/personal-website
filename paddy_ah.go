package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles("tmpl/index.html", "tmpl/admin.html", "tmpl/blog_admin.html", "tmpl/edit_blog_post.html"))
var validPath = regexp.MustCompile("^(/admin|^)/(blog/)?(create/)?$")

//go:embed admin_username.txt
var admin_username string

//go:embed admin_password.txt
var admin_password string

func checkPath(w http.ResponseWriter, r *http.Request) bool {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return false
	}
	return true
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "index")
	}
}

func adminPageHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "admin")
	}
}

func blogAdminHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "blog_admin")
	}
}

func blogCreateHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "edit_blog_post")
	}
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
	http.HandleFunc("/admin/blog/", basicAuth(blogAdminHandler))
	http.HandleFunc("/admin/blog/create/", basicAuth(blogCreateHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
