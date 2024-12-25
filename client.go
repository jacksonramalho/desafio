package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		log.Fatalf("Erro ao criar request")
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout ao tentar se conectar ao servidor.")
		} else {
			log.Printf("Erro na requisição: %v\n", err)
		}
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Erro do servidor: %s\n", string(body))
		return
	}

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Erro ao decodificar resposta JSON: %v", err)
		return
	}

	bid := response["bid"]

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	if err != nil {
		panic(err)
	}

	file.WriteString("Dólar: " + bid + "\n")

}
