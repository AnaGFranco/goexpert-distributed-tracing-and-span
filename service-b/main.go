package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	initTracer()

	http.HandleFunc("/weather", handleWeather)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Query().Get("cep")
	if len(cep) != 8 || !isNumeric(cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "handleWeather")
	defer span.End()

	location, err := getLocation(ctx, cep)
	if err != nil {
		log.Printf("Error getting location: %v", err)
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	weather, err := getWeather(ctx, location)
	if err != nil {
		log.Printf("Error getting weather: %v", err)
		http.Error(w, "error fetching weather", http.StatusInternalServerError)
		return
	}

	response := WeatherResponse{
		City:  location,
		TempC: weather,
		TempF: weather*1.8 + 32,
		TempK: weather + 273,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Utility functions remain unchanged as they are functioning correctly
func getLocation(ctx context.Context, cep string) (string, error) {
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get location")
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	location, ok := data["localidade"].(string)
	if !ok {
		return "", fmt.Errorf("localidade not found")
	}

	return location, nil
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func initTracer() {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("otel-collector:4317"), otlptracehttp.WithInsecure())
	if err != nil {
		log.Fatalf("failed to initialize trace exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("service-b"))),
	)
	otel.SetTracerProvider(tp)
}

func getWeather(ctx context.Context, location string) (float64, error) {
	// Configure o client HTTP para usar a instrumentação OpenTelemetry
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Substitua YOUR_API_KEY pela sua chave de API real
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=e5bd00e528e346ff8a840254213009&q&q=%s", location)

	// Criação da requisição HTTP com o contexto adequado
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating weather request: %v", err)
	}

	// Executando a requisição HTTP
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error executando a requisição de clima: %v", err)
	}
	defer resp.Body.Close()

	// Verificação do código de status da resposta
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get weather, status code: %d", resp.StatusCode)
	}

	// Decodificação da resposta JSON
	var result struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("error parsing weather response: %v", err)
	}

	return result.Current.TempC, nil
}
