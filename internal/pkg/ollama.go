package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmbeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingsResponse struct {
	Embedding []float32 `json:"embedding"`
}

func GetVector(text string) ([]float32, error) {

	url := "http://localhost:11434/api/embeddings"
	requestBody, err := json.Marshal(EmbeddingsRequest{
		Model:  "all-minilm",
		Prompt: text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: %d", resp.StatusCode)
	}

	var result EmbeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Embedding, nil
}
