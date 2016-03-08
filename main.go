package main

import (
	"fmt"
	"log"
	"net/http"
	"session/session"

	"gopkg.in/redis.v3"
)

type handler struct {
	body    string
	session *session.Session
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("リクエストを処理します %s %s", r.Method, r.RequestURI)
	w.Write([]byte(fmt.Sprintf("%s: %s\n", h.session.Name, h.body)))
	h.session.Name = h.session.Name + h.body
}

func (h *handler) SetSession(s *session.Session) {
	log.Print("セッションをセットします")
	h.session = s
}

func main() {
	session.Setup("example", "example", &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer session.Dispose()

	http.Handle("/", &session.Handler{Next: &handler{body: "root"}})
	http.Handle("/a", &session.Handler{Next: &handler{body: "a"}})
	http.Handle("/b", &session.Handler{Next: &handler{body: "b"}})
	http.Handle("/c", &session.Handler{Next: &handler{body: "c"}})

	log.Printf("サーバーを開始します。ポート: %s", ":8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}

}
