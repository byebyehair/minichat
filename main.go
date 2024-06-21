package main

import (
	"embed"
	"fmt"
	"log"
	"minichat/config"
	"minichat/conversation"
	"minichat/server"
	"net/http"
)

//go:embed static
var DirStatic embed.FS

//go:embed templates/*
var DirTemplate embed.FS

func main() {

	http.HandleFunc("/precheck", server.PreCheck)
	http.HandleFunc("/ws", server.HandleWs)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		server.HandleFiles(writer, request, DirTemplate)
	})
	fs := http.FileServer(http.FS(DirStatic))
	http.Handle("/static/", fs)

	go conversation.Manager.Start()

	configVal := config.ParseConfig("config.yaml")

	log.Printf("\n\n********************************\nChat server is running at %d !\n********************************\n\n", configVal.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", configVal.Port), nil)
	if err != nil {
		fmt.Printf("Server start fail, error is: [ %+v ]", err)
		return
	}
}
