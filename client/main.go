package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// estrutura de retorno do server
type Cotacao struct {
	Dolar string `json:"dolar"`
}

const serverApiTimeout = 300 * time.Millisecond

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), serverApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalln("Problema para criar a request", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln("Problema para executar a request", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln("Problema para ler o retorno da request", err)
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatalln("Problema no formato do retorno", err)
	}

	fmt.Printf("Cotação do dolar é: %s\n", cotacao.Dolar)

}
