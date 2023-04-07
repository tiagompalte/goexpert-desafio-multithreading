package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ViaCepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type CdnResponse struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	Status     int    `json:"status"`
	Ok         bool   `json:"ok"`
	StatusText string `json:"statusText"`
}

func requestViaCep(ctx context.Context, cep string, response chan<- ViaCepResponse) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep),
		nil,
	)
	if err != nil {
		fmt.Printf("ViaCepError: %s\n", err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ViaCepError: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("ViaCepError: Status code %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ViaCepError: %s\n", err.Error())
		return
	}

	var bodyResponse ViaCepResponse
	err = json.Unmarshal(body, &bodyResponse)
	if err != nil {
		fmt.Printf("ViaCepError: %s\n", err.Error())
		return
	}

	response <- bodyResponse
}

func requestCdn(ctx context.Context, cep string, response chan<- CdnResponse) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", cep),
		nil,
	)
	if err != nil {
		fmt.Printf("CdnError: %s\n", err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("CdnError: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("CdnError: Status code %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("CdnError: %s\n", err.Error())
		return
	}

	var bodyResponse CdnResponse
	err = json.Unmarshal(body, &bodyResponse)
	if err != nil {
		fmt.Printf("CdnError: %s\n", err.Error())
		return
	}

	response <- bodyResponse
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print("CEP not informed")
		return
	}
	cep := os.Args[1]

	ctx := context.Background()
	ctxReq, cancelReq := context.WithCancel(ctx)

	viaCepResponse := make(chan ViaCepResponse)
	go requestViaCep(ctxReq, cep, viaCepResponse)

	cdnResponse := make(chan CdnResponse)
	go requestCdn(ctx, cep, cdnResponse)

	select {
	case msg := <-viaCepResponse:
		fmt.Printf("Received from ViaCep: %+v", msg)
		ctxReq.Done()

	case msg := <-cdnResponse:
		fmt.Printf("Received from CdnApi: %+v", msg)
		ctxReq.Done()

	case <-time.After(time.Second * 1):
		fmt.Println("TIMEOUT")
		cancelReq()
	}

}
