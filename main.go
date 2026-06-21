package main

import "net/http"

func main() {
	servMux := http.NewServeMux()
	fileServ := http.FileServer(http.Dir(""))
	servMux.Handle("/", fileServ)
	var server http.Server
	server.Handler = servMux
	server.Addr = ":8080"
	server.ListenAndServe()

}
