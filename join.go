package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/a-know/yukizuri/trace"

	"github.com/stretchr/objx"
)

type joinHandlerStruct struct {
	next http.Handler
}

func (h *joinHandlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("yukizuri"); err == http.ErrNoCookie || cookie.Value == "" {
		// not joined yet
		w.Header().Set("Location", "/join")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// something wrong
		panic(err.Error())
	} else {
		// successful, call next wrapped handler
		h.next.ServeHTTP(w, r)
	}
}

func MustJoin(handler http.Handler) http.Handler {
	return &joinHandlerStruct{next: handler}
}

// Path style : /join/{nickname}
func joinHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	if len(segs) == 3 {
		nickname := segs[2]
		m := md5.New()
		io.WriteString(m, strings.ToLower(time.Now().String()))
		uniqueID := fmt.Sprintf("%x", m.Sum(nil))
		cookieValue := objx.New(map[string]interface{}{
			"userid":      uniqueID,
			"name":        nickname,
			"avatar_url":  "",
			"email":       "",
			"remote_addr": r.RemoteAddr,
		}).MustBase64()
		http.SetCookie(w, &http.Cookie{
			Name:  "yukizuri",
			Value: cookieValue,
			Path:  "/"})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusNotFound)
		tracer := trace.New()
		logContent := tracer.LogContent("system", "-", r.RemoteAddr, "-")
		tracer.TraceError(logContent, fmt.Errorf("Not supported request: %s", r.URL.Path))
	}
}
