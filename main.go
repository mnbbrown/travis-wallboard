package main

import (
	"encoding/json"
	"fmt"
	travis "github.com/Ableton/go-travis"
	"github.com/gorilla/websocket"
	_ "github.com/mnbbrown/travis-wallboard/statik"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Config is a config
type Config struct {
	Repositories      []string `json:"repositories"`
	GithubAccessToken string   `json:"github_access_token"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var notificationChan = make(chan []byte)
var repoStatus = make(map[string]*travis.Repository)

func loadRepository(client *travis.Client, slug string) (*travis.Repository, error) {
	repository, _, err := client.Repositories.GetFromSlug(slug)
	if err != nil {
		return nil, err
	}
	repoStatus[repository.Slug] = repository
	return repository, nil
}

func watchRepositories(client *travis.Client, config *Config) {
	for {

		for _, repo := range config.Repositories {
			repository, err := loadRepository(client, repo)
			if err != nil {
				log.Println(err)
				continue
			}
			de, err := json.Marshal(repository)
			if err != nil {
				log.Println(err)
				continue
			}
			notificationChan <- de
		}

		time.Sleep(5 * time.Second)
	}
}

func writer(ws *websocket.Conn) {
	for {
		select {
		case repo := <-notificationChan:
			if err := ws.WriteMessage(websocket.TextMessage, repo); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	writer(ws)
}

func main() {
	raw, err := ioutil.ReadFile("./config.json")
	var config *Config
	err = json.Unmarshal(raw, &config)
	if err != nil {
		log.Fatal(err)
	}

	client := travis.NewClient(travis.TRAVIS_API_PRO_URL, "")
	_, _, err := client.Authentication.UsingGithubToken(config.GithubAccessToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Loading")
	for _, repo := range config.Repositories {
		loadRepository(client, repo)
	}
	log.Println("Done")

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	go watchRepositories(client, config)
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(statikFS)))
	mux.HandleFunc("/ws", serveWs)
	mux.HandleFunc("/repos", func(rw http.ResponseWriter, req *http.Request) {
		de, err := json.Marshal(repoStatus)
		if err != nil {
			log.Println(err)
			http.Error(rw, "Ooops", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(rw, string(de))
	})

	handler := cors.Default().Handler(mux)
	if err := http.ListenAndServe(":5000", handler); err != nil {
		log.Fatal(err)
	}
}
