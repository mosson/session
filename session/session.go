package session

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/redis.v3"
)

var (
	staticRegistry  registry
	staticCookieKey string
)

// Handlee is net/http.Handler and can handle session
type Handlee interface {
	GetSession([]byte)
	SetSession() []byte
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Handler is handle net/http request with session, and call Next handler
type Handler struct {
	Next Handlee
}

// Sugar returns net/http handler wrapping session handler
func Sugar(handlee Handlee) *Handler {
	return &Handler{Next: handlee}
}

// Setup sets up redis store for session kvs.
func Setup(cookieKey string, namespace string, dialect int, options *redis.Options) {
	if staticRegistry != nil {
		return
	}

	staticCookieKey = cookieKey
	staticRegistry = newRegistry(namespace, dialect, options)
}

// Dispose closes redis.Client
func Dispose() error {
	if staticRegistry != nil {
		return nil
	}

	return staticRegistry.dispose()
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Next == nil {
		return
	}

	sessionID := getSessionID(r)
	setSessionID(w, sessionID)
	session, err := staticRegistry.get(sessionID)
	if err != nil {
		log.Fatal(err)
	}

	h.Next.GetSession(session)

	h.Next.ServeHTTP(w, r)

	err = staticRegistry.set(sessionID, h.Next.SetSession(), 0)
	if err != nil {
		log.Fatal(err)
	}
}

func setSessionID(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:    staticCookieKey,
		Value:   sessionID,
		Expires: time.Now().Add(5 * time.Minute),
	}

	http.SetCookie(w, &cookie)
}

// getSessionID returns random ID that stored in cookie or generated
func getSessionID(r *http.Request) string {
	cookie, err := r.Cookie(staticCookieKey)
	var sessionID string
	if cookie == nil || err != nil {
		cryptor := sha1.New()
		key := fmt.Sprintf("%d:%d", time.Now().UnixNano(), os.Getpid())
		cryptor.Write([]byte(key))
		sessionID = hex.EncodeToString(cryptor.Sum(nil))
	} else {
		sessionID = cookie.Value
	}

	return sessionID
}
