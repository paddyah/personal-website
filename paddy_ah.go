package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("tmpl/index.html", "tmpl/admin.html", "tmpl/blog_admin.html", "tmpl/edit_blog_post.html"))
var validPath = regexp.MustCompile("^(/admin|^)/(blog/)?(create/|save/|edit/|delete/)?(.+)?$")

//go:embed admin_username.txt
var admin_username string

//go:embed admin_password.txt
var admin_password string

type BlogPostEntry struct {
	Title string
	Url   string
}

type BlogPost struct {
	Title string
	Body  string
}

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
		renderTemplate(w, "index", nil)
	}
}

func adminPageHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "admin", nil)
	}
}

func blogAdminHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		files, err := os.ReadDir("posts")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		var blogPosts []BlogPostEntry
		blogPosts = make([]BlogPostEntry, 0)
		for _, file := range files {
			post := BlogPostEntry{Title: strings.TrimSuffix(file.Name(), ".md"), Url: "edit/" + file.Name()}
			blogPosts = append(blogPosts, post)
		}
		renderTemplate(w, "blog_admin", blogPosts)
	}
}

func blogCreateHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		renderTemplate(w, "edit_blog_post", nil)
	}
}

func blogEditHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
	}
	fileName := m[4]
	post, err := os.ReadFile("posts/" + fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	blogPost := BlogPost{Title: strings.TrimSuffix(fileName, ".md"), Body: string(post)}
	renderTemplate(w, "edit_blog_post", blogPost)
}

func blogSaveHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		switch r.Method {
		case "POST":
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			post := r.FormValue("blogPost")
			title := r.FormValue("title")
			// this doesn't have to optimized at all so an "edit" can just be a deletion and creation
			if r.FormValue("oldTitle") != "" {
				err := os.Remove("posts/" + r.FormValue("oldTitle") + ".md")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				f, err := os.Create("posts/" + title + ".md")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				_, err = f.WriteString(post)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					f.Close()
				}
				// over eager closing just to be careful
				f.Close()
			} else {
				f, err := os.Create("posts/" + time.Now().Format("2006-01-02") + " - " + title + ".md")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				_, err = f.WriteString(post)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					f.Close()
				}
				// over eager closing just to be careful
				f.Close()
			}
			http.Redirect(w, r, "/admin/blog/", http.StatusSeeOther)
		default:
			http.NotFound(w, r)
		}
	}
}

func blogDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if checkPath(w, r) {
		switch r.Method {
		case "POST":
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			title := r.FormValue("title")
			err = os.Remove("posts/" + title + ".md")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.NotFound(w, r)
		}
		http.Redirect(w, r, "/admin/blog/", http.StatusSeeOther)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
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
	http.HandleFunc("/admin/blog/save/", basicAuth(blogSaveHandler))
	http.HandleFunc("/admin/blog/edit/", basicAuth(blogEditHandler))
	http.HandleFunc("/admin/blog/delete/", basicAuth(blogDeleteHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
