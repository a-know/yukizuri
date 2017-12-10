package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/a-know/yukizuri/trace"

	"github.com/stretchr/objx"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// Handling HTTP Request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		f, err := Assets.Open(filepath.Join("/templates", t.filename))
		if err != nil {
			tracer := trace.New()
			logContent := tracer.LogContent("system", "-", "-", "TemplateHandler")
			tracer.TraceError(logContent, err)
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			tracer := trace.New()
			logContent := tracer.LogContent("system", "-", "-", "TemplateHandler")
			tracer.TraceError(logContent, err)
		}

		var ns = template.New("template")
		t.templ, _ = ns.Parse(string(data))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if cookie, err := r.Cookie("yukizuri"); err == nil {
		data["UserData"] = objx.MustFromBase64(cookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "port number")
	var logging = flag.Bool("logging", true, "logging with stdout")
	flag.Parse()
	r := newRoom(*logging)
	http.Handle("/chat", MustJoin(&templateHandler{filename: "yukizuri.html"}))
	http.Handle("/join", &templateHandler{filename: "join.html"})
	http.Handle("/", &templateHandler{filename: "join.html"})
	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "yukizuri",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.HandleFunc("/join/", joinHandler)
	http.Handle("/room", r) // for WebSocket connection endpoint
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("./fonts"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))
	http.Handle("/plugins/", http.StripPrefix("/plugins/", http.FileServer(http.Dir("./plugins"))))
	// Starting chatroom
	go r.run()
	// Starting web server
	tracer := trace.New()
	logContent := tracer.LogContent("system", "-", "-", "Starting Web server...")
	tracer.TraceInfo(logContent)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		logContent := tracer.LogContent("system", "-", "-", "ListenAndServe")
		tracer.TraceError(logContent, err)
	}
}
