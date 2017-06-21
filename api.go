package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

const salt = "salmon rosado"

var redisClient *redis.Client

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	r := mux.NewRouter()
	corsRouter := mux.NewRouter()

	corsHandler := gh.CORS(gh.AllowCredentials(), gh.AllowedHeaders([]string{"x-requested-with", "content-type"}), gh.AllowedOrigins([]string{"*"}))

	r.HandleFunc("/ping", ping).Methods("GET")
	corsRouter.HandleFunc("/share", share).Methods("POST")
	corsRouter.HandleFunc("/shares/{id}", shares).Methods("GET")

	n := negroni.Classic()
	r.PathPrefix("/").Handler(negroni.New(negroni.Wrap(corsHandler(corsRouter))))
	n.UseHandler(r)

	httpServer := http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           n,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := httpServer.ListenAndServe()
	log.Fatal(err)
}

func ping(rw http.ResponseWriter, req *http.Request) {
}

func share(rw http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)

	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	i := id(b)
	err = redisClient.Set(i, b, 0).Err()
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(rw, i)
}

func shares(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	i := vars["id"]

	b, err := redisClient.Get(i).Result()
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprint(rw, b)
}

func id(body []byte) string {
	h := sha1.New()
	io.WriteString(h, salt)
	h.Write(body)
	sum := h.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)
	return string(b)[:10]
}
