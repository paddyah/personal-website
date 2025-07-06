package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

var templates = template.Must(template.ParseFiles("tmpl/index.html",
	"tmpl/admin.html",
	"tmpl/blog_admin.html",
	"tmpl/edit_blog_post.html",
	"tmpl/render_post.html",
	"tmpl/blog_list.html",
	"tmpl/links.html",
))

var validPath = regexp.MustCompile("^/(?:about/|links/|blog/(?:view/(.+))?)?$|^/admin/(?:blog/(?:edit/(.+)|delete/|save/|create/)?)?")

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

func checkErr(w http.ResponseWriter, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", nil)
}

func aboutPageHandler(w http.ResponseWriter, r *http.Request) {
	post, err := os.ReadFile("post_html/" + "about_me.html")
	checkErr(w, err)

	renderTemplate(w, "render_post", template.HTML(post))
}

func linkPageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "links", nil)
}

func blogHomeHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("posts")
	checkErr(w, err)
	var blogPosts []BlogPostEntry
	blogPosts = make([]BlogPostEntry, 0)
	for _, file := range files {
		// crude drafts implementation.
		// due to ease of implementation and ability to migrate to more robust solution if necessary this is fine enough to use right now
		if !strings.Contains(file.Name(), "DRAFT") {
			post := BlogPostEntry{Title: strings.TrimSuffix(file.Name(), ".md"), Url: "view/" + strings.TrimSuffix(file.Name(), ".md")}
			blogPosts = append(blogPosts, post)
		}
	}
	slices.SortFunc(blogPosts, func(a, b BlogPostEntry) int {
		return strings.Compare(strings.ToLower(b.Title), strings.ToLower(a.Title))
	})
	renderTemplate(w, "blog_list", blogPosts)
}

func blogViewHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
	}
	fileName := m[1]
	post, err := os.ReadFile("post_html/" + fileName + ".html")
	checkErr(w, err)

	renderTemplate(w, "render_post", template.HTML(post))
}

func adminPageHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin", nil)
}

func blogAdminHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("posts")
	checkErr(w, err)
	var blogPosts []BlogPostEntry
	blogPosts = make([]BlogPostEntry, 0)
	for _, file := range files {
		post := BlogPostEntry{Title: strings.TrimSuffix(file.Name(), ".md"), Url: "edit/" + file.Name()}
		blogPosts = append(blogPosts, post)
	}
	renderTemplate(w, "blog_admin", blogPosts)
}

func blogCreateHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "edit_blog_post", nil)
}

func blogEditHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
	}
	fileName := m[2]
	post, err := os.ReadFile("posts/" + fileName)
	checkErr(w, err)
	blogPost := BlogPost{Title: strings.TrimSuffix(fileName, ".md"), Body: string(post)}
	renderTemplate(w, "edit_blog_post", blogPost)
}

func blogSaveHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		err := r.ParseForm()
		checkErr(w, err)
		post := r.FormValue("blogPost")
		title := r.FormValue("title")
		postHTML := r.FormValue("hiddenHTML")
		// this doesn't have to optimized at all so an "edit" can just be a deletion and creation
		if r.FormValue("oldTitle") != "" {
			// removing old values
			err := os.Remove("posts/" + r.FormValue("oldTitle") + ".md")
			checkErr(w, err)
			err = os.Remove("post_html/" + r.FormValue("oldTitle") + ".html")
			checkErr(w, err)

			// saving new value without automatically generating a time for the title
			f, err := os.Create("posts/" + title + ".md")
			checkErr(w, err)
			_, err = f.WriteString(post)
			checkErr(w, err)
			f.Close()

			f, err = os.Create("post_html/" + title + ".html")
			checkErr(w, err)
			_, err = f.WriteString(postHTML)
			checkErr(w, err)
			f.Close()
		} else {
			fileName := time.Now().Format("2006-01-02") + " - " + title

			// saving markdown
			f, err := os.Create("posts/" + fileName + ".md")
			checkErr(w, err)
			_, err = f.WriteString(post)
			checkErr(w, err)
			f.Close()

			// saving html
			f, err = os.Create("post_html/" + fileName + ".html")
			checkErr(w, err)
			_, err = f.WriteString(postHTML)
			checkErr(w, err)
			f.Close()
		}
		http.Redirect(w, r, "/admin/blog/", http.StatusSeeOther)
	default:
		http.NotFound(w, r)
	}
}

func blogDeleteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		err := r.ParseForm()
		checkErr(w, err)
		title := r.FormValue("title")
		err = os.Remove("posts/" + title + ".md")
		checkErr(w, err)
		err = os.Remove("post_html/" + title + ".html")
		checkErr(w, err)
	default:
		http.NotFound(w, r)
	}
	http.Redirect(w, r, "/admin/blog/", http.StatusSeeOther)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	checkErr(w, err)
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

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r)
	}
}

func main() {
	http.HandleFunc("/", makeHandler(homePageHandler))
	http.HandleFunc("/about/", makeHandler(aboutPageHandler))
	http.HandleFunc("/links/", makeHandler(linkPageHandler))
	http.HandleFunc("/blog/", makeHandler(blogHomeHandler))
	http.HandleFunc("/blog/view/", makeHandler(blogViewHandler))
	http.HandleFunc("/admin/", makeHandler(basicAuth(adminPageHandler)))
	http.HandleFunc("/admin/blog/", makeHandler(basicAuth(blogAdminHandler)))
	http.HandleFunc("/admin/blog/create/", makeHandler(basicAuth(blogCreateHandler)))
	http.HandleFunc("/admin/blog/save/", makeHandler(basicAuth(blogSaveHandler)))
	http.HandleFunc("/admin/blog/edit/", basicAuth(blogEditHandler))
	http.HandleFunc("/admin/blog/delete/", makeHandler(basicAuth(blogDeleteHandler)))

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
