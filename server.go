package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// struct para recuperar a cotacao
type Cotacao struct {
	ID     string
	USDBRL USDBRL
}

type USDBRL struct {
	Bid string
}

// struct para persistir as cotações no sqlite
type CotacaoBanco struct {
	ID    string `gorm:"primaryKey;autoIncrement"`
	Valor string
	gorm.Model
}

func main() {

	http.HandleFunc("/cotacao", GetCotacaoHandler)
	http.ListenAndServe(":8080", nil)

}

func GetCotacaoHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Passou 1")

	//criando e setando time out no contexto para a operação de GET de cotação
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	//criando requisicao e passando o contexto
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	//fazendo a chamada para pegar as cotações
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("Erro ao buscar cotações")
		panic(err)
	}

	defer resp.Body.Close()

	//parseando a resposta
	var cotacao Cotacao
	err = json.NewDecoder(resp.Body).Decode(&cotacao)

	if err != nil {
		panic(err)
	}

	//Salvando cotacao
	SalvandoCotacaoNoBanco(cotacao)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)

}

func SalvandoCotacaoNoBanco(cotacao Cotacao) {

	db := AbirConexao()

	ctx2, _ := context.WithTimeout(context.Background(), time.Millisecond*10)

	cotacaoBanco := &CotacaoBanco{Valor: cotacao.USDBRL.Bid}

	if err := db.WithContext(ctx2).Create(&cotacaoBanco).Error; err != nil {
		log.Fatalf("Erro ao criar cotação no banco: %v", err)
	}

}

func AbirConexao() *gorm.DB {

	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})

	if err != nil {
		log.Fatalf("Erro ao abrir a conexao")
		return nil
	}

	fmt.Println("Conexão aberta com sucesso")

	db.AutoMigrate(&CotacaoBanco{})

	return db
}
