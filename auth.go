package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stretchr/gomniauth"
	gomniauthcommon "github.com/stretchr/gomniauth/common"
	"github.com/stretchr/objx"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	gomniauthcommon.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("auth"); err == http.ErrNoCookie || cookie.Value == "" {
		// not authenticated yet
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else if err != nil {
		// something wrong
		panic(err.Error())
	} else {
		// successful, call next wrapped handler
		h.next.ServeHTTP(w, r)
	}
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// Path style : /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	if len(segs) == 4 {
		action := segs[2]
		provider := segs[3]
		switch action {
		case "login":
			provider, err := gomniauth.Provider(provider)
			if err != nil {
				log.Fatalln("Failed to get auth provider:", provider, "-", err)
			}
			loginURL, err := provider.GetBeginAuthURL(nil, nil)
			if err != nil {
				log.Fatalln("Error occurs in calling GetBeginAuthURL:", provider, "-", err)
			}
			w.Header().Set("Location", loginURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		case "callback":
			provider, err := gomniauth.Provider(provider)
			if err != nil {
				log.Fatalln("Failed to get provider:", provider, "-", err)
			}
			creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
			if err != nil {
				log.Fatalln("Could not finish authentication:", provider, "-", err)
			}
			user, err := provider.GetUser(creds)
			if err != nil {
				log.Fatalln("Failed to get user data:", provider, "-", err)
			}
			chatUser := &chatUser{User: user}
			m := md5.New()
			io.WriteString(m, strings.ToLower(user.Email()))
			chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
			avatarURL, err := avatars.GetAvatarURL(chatUser)
			if err != nil {
				log.Fatalln("Failed to GetAvatarURL.", "-", err)
			}
			authCookieValue := objx.New(map[string]interface{}{
				"userid":     chatUser.uniqueID,
				"name":       user.Name(),
				"avatar_url": avatarURL,
				"email":      user.Email(),
			}).MustBase64()
			http.SetCookie(w, &http.Cookie{
				Name:  "auth",
				Value: authCookieValue,
				Path:  "/"})
			w.Header()["Location"] = []string{"/chat"}
			w.WriteHeader(http.StatusTemporaryRedirect)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not supported action: %s", action)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not supported request: %s", r.URL.Path)
	}
}
