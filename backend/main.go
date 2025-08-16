package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hdarweesh/rcoder/backend/redis"

	"github.com/rs/cors"
	"github.com/segmentio/kafka-go"
)

// CodeExecutionRequest represents the user-submitted execution request
type CodeExecutionRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// CodeExecutionResponse represents the result returned from the executors
type CodeExecutionResponse struct {
	TaskID string `json:"task_id"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Kafka topics
const (
	kafkaBroker   = "kafka:9092"
	resultTopic   = "code-execution-results"
	requestPrefix = "code-execution-tasks-" // Requests will be sent to language-specific topics
)

// Produce messages to the correct Kafka topic
func produceExecutionRequest(language, code string) (string, error) {
	topic := requestPrefix + language
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	taskID := fmt.Sprintf("%d", time.Now().UnixNano())
	message := CodeExecutionRequest{Language: language, Code: code}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %v", err)
	}

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(taskID),
		Value: messageBytes,
	})
	if err != nil {
		return "", fmt.Errorf("error sending Kafka message: %v", err)
	}

	log.Printf("Sent message to Kafka topic: %s with TaskID: %s", topic, taskID)
	return taskID, nil
}

// Consume results from Kafka
func consumeExecutionResult(taskID string) (CodeExecutionResponse, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    resultTopic,
		GroupID:  "backend-consumer",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		// CommitInterval: time.Second,
		StartOffset: kafka.FirstOffset,
	})

	defer reader.Close()

	for {
		log.Println("Starting to consume results...")

		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			return CodeExecutionResponse{}, fmt.Errorf("error reading Kafka message: %v", err)
		}

		var result CodeExecutionResponse
		if err := json.Unmarshal(msg.Value, &result); err != nil {
			log.Printf("Invalid JSON result: %v", err)
			continue
		}
		log.Printf("consumed result: %v", result)

		if result.TaskID == taskID {
			log.Printf("Successfully consumed result for : %s: %v", taskID, result)
			return result, nil
		}
	}
}

// HTTP handler for code execution
func executeCodeHandler(w http.ResponseWriter, r *http.Request) {
	var request CodeExecutionRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Language == "" || request.Code == "" {
		log.Println("Missing required fields:", request)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	redisCache := redis.GetRedisCache()
	codeHash := redisCache.HashCode(request.Code)

	// Check if code result is already cached
	var cachedResult CodeExecutionResponse
	if err := redisCache.GetCachedResult(codeHash, &cachedResult); err == nil {
		// Cache hit, return result
		responseBytes, _ := json.Marshal(cachedResult)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseBytes)
		log.Printf("Cache hit for code hash: %s, returning cached result", codeHash)
		return
	}

	// Produce execution request
	taskID, err := produceExecutionRequest(request.Language, request.Code)
	if err != nil {
		log.Printf("Failed to enqueue execution: %v", err)
		http.Error(w, "Failed to enqueue execution", http.StatusInternalServerError)
		return
	}

	// Wait for result from Kafka
	result, err := consumeExecutionResult(taskID)
	if err != nil {
		log.Printf("Failed to consume result: %v", err)
		http.Error(w, "Execution timed out", http.StatusGatewayTimeout)
		return
	}

	// Cache the result using code hash as key
	_ = redisCache.CacheResult(codeHash, result)

	responseBytes, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func main() {
	log.Println("Starting HTTP server")
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", executeCodeHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Health OK"))
	})

	// Configure CORS options
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow specific origin
		AllowedMethods:   []string{"GET", "POST"},           // Allow specific methods
		AllowedHeaders:   []string{"Content-Type"},          // Allow specific headers
		AllowCredentials: true,
	})

	// Wrap the router with the CORS middleware
	handler := c.Handler(mux)

	http.ListenAndServe(":8080", handler)
}
