package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// estrutura de retorno de https://economia.awesomeapi.com.br/json/last/USD-BRL
type AwesomeEconomiaUSDBRL struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

// estrutura de retorno para o client
type Cotacao struct {
	Dolar string `json:"dolar"`
}

const cotacoesDSN = "cotacoes.db"
const awesomeApiTimeout = 200 * time.Millisecond
const registerCotacaoTimeout = 10 * time.Millisecond

// variavel global para permitir a inclusão de cotação acessar o banco
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", cotacoesDSN)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	setubDatabase()

	http.HandleFunc("/cotacao", requestCotacao)
	http.ListenAndServe(":8080", nil)
}

func setubDatabase() {
	// Cria a tabela de cotações se não existir
	statement, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY,
			timestamp VARCHAR(20),
			dolar VARCHAR(10))`)
	if err != nil {
		log.Panic(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Panic(err)
	}
}

func requestCotacao(w http.ResponseWriter, r *http.Request) {

	var data AwesomeEconomiaUSDBRL
	err := buscaAwesomeCotacaoUSDBRL(&data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = gravaCotacao(data.USDBRL.Bid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cotacao := Cotacao{Dolar: data.USDBRL.Bid}
	montaResposta(cotacao, w)
}

func buscaAwesomeCotacaoUSDBRL(cotacao *AwesomeEconomiaUSDBRL) error {
	ctx, cancel := context.WithTimeout(context.Background(), awesomeApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println("Problema para criar request", err)
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Problema para executar request", err)
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Problema para ler request", err)
		return err
	}
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Println("Problema no formato da resposta", err)
		return err
	}
	return nil
}

func gravaCotacao(cotacao string) error {
	ctx, cancel := context.WithTimeout(context.Background(), registerCotacaoTimeout)
	defer cancel()
	stmt, err := db.Prepare("insert into cotacoes(id, timestamp, dolar) values(NULL, DATETIME('now'), ?)")
	if err != nil {
		log.Println("Problema na gravacao da cotacao", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao)
	if err != nil {
		log.Println("Problema gravando da cotacao", err)
		return err
	}
	return nil
}

func montaResposta(cotacao Cotacao, w http.ResponseWriter) {
	err := json.NewEncoder(w).Encode(cotacao)
	if err != nil {
		log.Println("Problema para formatar a resposta", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
