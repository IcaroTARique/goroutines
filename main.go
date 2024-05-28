package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IcaroTARique/goroutines.git/dto"
	"net/http"
	"sync"
	"time"
)

type ApiResponse struct {
	Name    string
	Content *http.Response
}

func main() {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	escolha := make(chan ApiResponse)

	brasilApiRequest, err := http.NewRequestWithContext(ctx, "GET", "https://brasilapi.com.br/api/cep/v1/58046320", nil)
	if err != nil {
		panic(err)
	}
	viaCepApiRequest, err := http.NewRequestWithContext(ctx, "GET", "https://viacep.com.br/ws/58046320/json/", nil)
	if err != nil {
		panic(err)
	}

	go ViaCepCall(ctx, viaCepApiRequest, escolha, &waitGroup)
	go BrasilApiCall(ctx, brasilApiRequest, escolha, &waitGroup)

	apiResponse := <-escolha

	if apiResponse.Name == "BrasilApi" && apiResponse.Content != nil {
		var brasilApiEndereco dto.BrasilApiEndereco
		err = json.NewDecoder(apiResponse.Content.Body).Decode(&brasilApiEndereco)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %+v ", apiResponse.Name, brasilApiEndereco)
	}
	if apiResponse.Name == "ViaCep" && apiResponse.Content != nil {
		var viaCepEndereco dto.ViaCepEndereco
		err = json.NewDecoder(apiResponse.Content.Body).Decode(&viaCepEndereco)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: %+v ", apiResponse.Name, viaCepEndereco)
	}

	close(escolha)
	waitGroup.Wait()
}

func BrasilApiCall(ctx context.Context, req *http.Request, escolha chan<- ApiResponse, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	select {
	case <-ctx.Done():
		fmt.Println("BrasilApiCall cancelled from user before starting")
		return
	default:
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("BrasilApiCall cancelled after started and runned out of time")
			escolha <- ApiResponse{Name: "BrasilApi", Content: nil}
			return
		}
		panic(err)
	}

	select {
	case <-ctx.Done():
		fmt.Println("BrasilApiCall cancelled after started and runned out of time")
		return
	default:
		escolha <- ApiResponse{Name: "BrasilApi", Content: res}
	}
}

func ViaCepCall(ctx context.Context, req *http.Request, escolha chan<- ApiResponse, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	select {
	case <-ctx.Done():
		fmt.Println("TIMEOUT: ViaCepCall cancelled from user before starting")
		return
	default:
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("TIMEOUT: ViaCepCall cancelled after started and runned out of time")
			escolha <- ApiResponse{Name: "ViaCep", Content: nil}
			return
		}
		panic(err)
	}

	select {
	case <-ctx.Done():
		fmt.Println("ViaCepCall cancelled after started and runned out of time")
		return
	default:
		escolha <- ApiResponse{Name: "ViaCep", Content: res}
	}
}
