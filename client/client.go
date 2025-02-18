package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// estrutura de retorno do server
type Cotacao struct {
	Dolar string `json:"dolar"`
}

const serverApiTimeout = 300 * time.Millisecond

func main() {

	cotacao, err := getCotacao()
	if err != nil {
		log.Fatalln("Erro ao buscar a cotação do server", err)
	}

	err = saveCotacao(cotacao)
	if err != nil {
		log.Fatalln("Erro ao gravar a cotação no arquivo", err)
	}

	fmt.Printf("Cotação do dolar é: %s\n", cotacao.Dolar)

}

func getCotacao() (Cotacao, error) {
	ctx, cancel := context.WithTimeout(context.Background(), serverApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return Cotacao{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Cotacao{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Cotacao{}, err
	}
	var cotacao Cotacao
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		return Cotacao{}, err
	}
	return cotacao, nil
}

func saveCotacao(cotacao Cotacao) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	tmp := template.New("Cotacao")
	tmp, err = tmp.Parse("Dólar: {{.Dolar}}")
	if err != nil {
		return err
	}
	err = tmp.Execute(file, cotacao)
	if err != nil {
		return err
	}

	return nil

}
