package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func main() {
	jsonStr := `[
  {
    "output": "Hello! How can I assist you today?"
  }
]`
	bodyBytes := []byte(jsonStr)

	// Simulate reading body
	fmt.Printf("Body: %s\n", string(bodyBytes))

	// Restore
	bodyReader := io.NopCloser(bytes.NewBuffer(bodyBytes))

	var response []map[string]interface{}
	if err := json.NewDecoder(bodyReader).Decode(&response); err != nil {
		fmt.Printf("Error decoding: %v\n", err)
		return
	}

	if len(response) > 0 {
		if output, ok := response[0]["output"].(string); ok {
			fmt.Printf("Parsed Output: %s\n", output)
		} else {
			fmt.Println("Output field not found or not string")
		}
	} else {
		fmt.Println("Empty array")
	}
}
