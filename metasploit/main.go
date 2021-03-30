package main

import (
	"bytes"
	"github.com/vmihailenco/msgpack/v5"
	"fmt"
	"net/http"
	"log"
)

type Metasploit struct {
	host string
	user string
	pass string
	token string
}

type loginReq struct {
	_msgpack struct{} `msgpack:",asArray"`
	Method string
	Username string 
	Password string
}

type loginRes struct {
	Result string `msgpack:"result"`
	Token  string `msgpack:"token"`
	Error  bool  `msgpack:"error"`
	ErrorClass  string  `msgpack:"error_class"`
	ErrorMessage  string  `msgpack:"error_message"`
}

type sessionListReq struct {
	_msgpack struct{} `msgpack:",asArray"`
	Method string 
	Token string
}

type SessionListRes struct {
	ID uint32 `msgpack:",omitempty"`
	Type  string `msgpack:"type"`
	TunnelLocal string `msgpack:"tunnel_local"`
	TuneelPeer string `msgpack:"tunnel_peer"`
	ViaExploit string  `msgpack:"via_exploit"`
	ViaPayload string `msgpack:"via_payload"`
	Description string `msgpack:"desc"`
	Info  string `msgpack:"info"`
	Workspace  string `msgpack:"workspace"`
	SessionHost  string `msgpack:"session_host"`
	SessionPort string `msgpack:"session_port"`
	Username string `msgpack:"username"`
	UUID  string `msgpack:"uuid"`
	ExploitUUID string `msgpack:"exploit_uuid"`	
}

func main(){

	host := "10.0.1.6:55552"
	pass := "s3cr3t"
	user := "msf"

	if host == "" || pass == "" {
		log.Fatalln("Missing required enviroment")
	}

	ctx := &loginReq{
		Method: "auth.login",
		Username: user,
		Password: pass, 
	}

	var res loginRes

	buf := new(bytes.Buffer)
	msgpack.NewEncoder(buf).Encode(ctx)
	dest := fmt.Sprintf("http://%s/api", host)
	r, error := http.Post(dest, "binary/message-pack", buf)

	if error != nil {
		log.Panicln(error)
	}

	defer r.Body.Close()

	if err := msgpack.NewDecoder(r.Body).Decode(&res); err != nil {
		log.Panicln(err)
	}

	fmt.Println(res.Token)

	reqSession := &sessionListReq{
		Method: "session.list",
		Token: res.Token,
	}

	resSession := make(map[uint32]SessionListRes)

	bufSession :=  new(bytes.Buffer)
	msgpack.NewEncoder(bufSession).Encode(&reqSession)
	response, error := http.Post(dest,"binary/message-pack", bufSession)
	if error != nil {
		log.Panicln(error)
	}

	defer response.Body.Close()

	if err := msgpack.NewDecoder(response.Body).Decode(&resSession); err != nil {
		fmt.Println("error fazendo a conversão")
		log.Panicln(err)
	}

	fmt.Println("Sessions:")
	for id, session := range resSession {
		session.ID = id
		resSession[id] = session
	}


	fmt.Println(resSession)
	for _, session := range resSession {
		fmt.Println(session)
	}
}
