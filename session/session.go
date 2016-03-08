package session

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/redis.v3"
)

var (
	staticClient    *redis.Client
	staticCookieKey string
	staticNamespace string
)

// Handlee is net/http.Handler and can set session object
type Handlee interface {
	SetSession(*Session)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Handler is handle net/http request with session, and call Next handler
type Handler struct {
	Next Handlee
}

// Session is struct of sessoin data
type Session struct {
	Name string `json:"name"`
}

// Setup sets up redis store for session kvs.
func Setup(cookieKey string, namespace string, options *redis.Options) {
	if staticClient != nil {
		return
	}

	staticCookieKey = cookieKey
	staticNamespace = namespace

	newClient := redis.NewClient(options)
	_, err := newClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	staticClient = newClient
}

// Dispose closes redis.Client
func Dispose() {
	if staticClient != nil {
		return
	}

	err := staticClient.Close()

	if err != nil {
		panic(err)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := GetSessionID(r)

	log.Printf("セッションID: %s", sessionID)

	val, err := staticClient.Get(fmt.Sprintf("%s:%s", staticNamespace, sessionID)).Result()
	var session *Session
	if err == redis.Nil {
		log.Printf("セッションは空でした。新規作成します。\n")
		session = &Session{}
		writeCookie(w, sessionID)
	} else if err != nil {
		panic(err)
	} else {
		log.Printf("セッションを取得しました。 %v", val)
		session = GetSession(val)
	}

	defer writeSession(sessionID, session)

	if h.Next == nil {
		return
	}
	h.Next.SetSession(session)
	h.Next.ServeHTTP(w, r)
}

func writeCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:    staticCookieKey,
		Value:   sessionID,
		Expires: time.Now().Add(5 * time.Minute),
	}

	http.SetCookie(w, &cookie)
}

func writeSession(sessionID string, session *Session) {
	data, err := json.Marshal(session)

	if err != nil {
		log.Println(err)
		return
	}

	staticClient.Set(
		fmt.Sprintf("%s:%s", staticNamespace, sessionID),
		data,
		0,
	)
}

// GetSession returns Session struct made from json data
func GetSession(data string) *Session {
	var session Session
	err := json.Unmarshal([]byte(data), &session)

	if err != nil {
		log.Println(err)
		return &Session{}
	}

	return &session
}

// GetSessionID returns random ID that stored in cookie or generated
func GetSessionID(r *http.Request) string {
	cookie, err := r.Cookie(staticCookieKey)
	var sessionID string
	if cookie == nil || err != nil {
		cryptor := sha1.New()
		key := strconv.Itoa(int(time.Now().UnixNano()))
		cryptor.Write([]byte(key))
		sessionID = hex.EncodeToString(cryptor.Sum(nil))
	} else {
		sessionID = cookie.Value
	}

	return sessionID
}
