package main

import (
	"fmt"
	"log"
	"net/http"
	"session/session"
)

type handler struct {
	body    string
	session []byte
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("リクエストを処理します %s %s", r.Method, r.RequestURI)
	w.Write([]byte(fmt.Sprintf("%s: %s\n", h.session, h.body)))
}

func (h *handler) GetSession(s []byte) {
	log.Printf("セッションを取得しました: %s", s)
	h.session = s
}

func (h *handler) SetSession() []byte {
	log.Printf("セッションを書き込みます: %s", h.body)
	return []byte(h.body)
}

func main() {
	session.Setup("cookieKey", "sessionNameSpace", session.MemoryDialect, nil)
	defer session.Dispose()

	http.Handle("/a", session.Sugar(&handler{body: "a"}))
	http.Handle("/b", session.Sugar(&handler{body: "b"}))
	http.Handle("/c", session.Sugar(&handler{body: "c"}))

	log.Printf("サーバーを開始します。ポート: %s", ":8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}

}
