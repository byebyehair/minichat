package main

import (
	"embed"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"minichat/config"
	"minichat/conversation"
	"minichat/server"
	"net/http"
)

//go:embed templates/*
var StaticFiles embed.FS

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/ws", server.HandleWs)
	r.HandleFunc("/", HandleFiles)
	//r.HandleFunc("/ws/{room}/{user}/{pwd}/{cmd}", handleConnections)

	go conversation.Manager.Start()

	configVal := config.ParseConfig("config.yaml")

	log.Printf("\n\n********************************\nChat server is running at %d !\n********************************\n\n", configVal.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", configVal.Port), r)
	if err != nil {
		fmt.Printf("Server start fail, error is: [ %+v ]", err)
		return
	}
}

func HandleFiles(w http.ResponseWriter, _ *http.Request) {
	data := struct {
		Url string
	}{
		Url: config.GlobalConfig.ServerUrl,
	}

	tmpl, err := template.ParseFS(StaticFiles, "templates/index.html")
	if err != nil {
		fmt.Printf("failed to parse the template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		fmt.Printf("failed to execute the template: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
