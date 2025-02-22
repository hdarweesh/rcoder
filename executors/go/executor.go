package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/segmentio/kafka-go"
)

type CodeExecutionRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type CodeExecutionResponse struct {
	TaskID string `json:"task_id"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

const (
	kafkaBroker      = "kafka:9092"
	executionTopic   = "code-execution-tasks-go"
	resultTopic      = "code-execution-results"
	consumerGroupID  = "executor-go"
	executionTimeout = 5 * time.Second
)

func executeGoCode(code string) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "go-execution")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %v", err)
	}
	// Do NOT defer os.RemoveAll(tempDir) immediately; remove it after execution

	// Create a Go module in the temp directory
	goModPath := filepath.Join(tempDir, "go.mod")
	goModContent := "module tempmod\n\ngo 1.22\n"
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing go.mod file: %v", err)
	}

	// Create Go source file
	codeFilePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(codeFilePath, []byte(code), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing code to file: %v", err)
	}

	// Run Go code within the temporary directory
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = tempDir // Set working directory

	output, err := cmd.CombinedOutput()
	// Clean up temp files after execution
	os.RemoveAll(tempDir)

	if err != nil {
		return "", fmt.Errorf("execution error: %s\nOutput:\n%s", err, string(output))
	}

	return string(output), nil
}

func consumeAndExecute() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaBroker},
		GroupID:        consumerGroupID,
		Topic:          executionTopic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaBroker},
		Topic:   resultTopic,
	})

	defer reader.Close()
	defer writer.Close()

	fmt.Println("Go Executor is running...")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}

		var task CodeExecutionRequest
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		output, execErr := executeGoCode(task.Code)
		response := CodeExecutionResponse{
			TaskID: string(msg.Key),
			Output: output,
		}
		log.Println("execution result:", output)

		if execErr != nil {
			response.Error = execErr.Error()
		}
		log.Println("execution response: ", response)

		responseJSON, _ := json.Marshal(response)
		err = writer.WriteMessages(context.Background(), kafka.Message{
			Key:   msg.Key,
			Value: responseJSON,
		})

		if err != nil {
			log.Println("Error sending execution result:", err)
		} else {
			log.Println("Execution result sent successfully")
		}
	}
}

func main() {
	consumeAndExecute()
}
