package main

import (
	"context"
	"encoding/json"
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

	http.HandleFunc("/cotacao", GetQuotationHandler)
	http.ListenAndServe(":8080", nil)

}

func GetQuotationHandler(w http.ResponseWriter, r *http.Request) {

	var cotacao Cotacao

	if err := GetQuotationFromAPI(&cotacao); err != nil {
		http.Error(w, "Erro ao buscar a cotação na API de Cotações: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := SavingQuotationInDB(&cotacao); err != nil {
		http.Error(w, "Erro ao salvar a cotação no DB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"bid": cotacao.USDBRL.Bid,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Erro ao formatar a resposta", http.StatusInternalServerError)
		return
	}

}

func GetQuotationFromAPI(cotacao *Cotacao) error {

	//criando e setando time out no contexto para a operação de GET de cotação
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	//criando requisicao e passando o contexto
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		log.Fatalf("Erro ao criar request para API de cotações")
		return err
	}

	//fazendo a chamada para pegar as cotações na API
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout atingido ao buscar cotações na API de cotações.")
			return err
		} else {
			log.Fatalf("Erro ao buscar cotações na API de cotações: %v", err)
			return err
		}

	}

	defer resp.Body.Close()

	//parseando a resposta
	if err = json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Fatalf("Erro no parse da resposta da cotação")
		return err
	}

	return nil
}

func SavingQuotationInDB(cotacao *Cotacao) error {

	db := AbirConexao()

	ctx2, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	//Salvando cotação no banco - com contexto
	cotacaoBanco := &CotacaoBanco{Valor: cotacao.USDBRL.Bid}

	if err := db.WithContext(ctx2).Create(&cotacaoBanco).Error; err != nil {
		if ctx2.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout atingido ao salvar as cotações no DB.")
			return err
		} else {
			log.Fatalf("Erro ao salvar a cotação no DB: %v", err)
			return err
		}
	}

	return nil
}

func AbirConexao() *gorm.DB {

	//Abrindo conexão sqlite
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})

	if err != nil {
		log.Fatalf("Erro ao abrir a conexao")
		return nil
	}

	db.AutoMigrate(&CotacaoBanco{})

	return db
}
