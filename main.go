package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hey go!")

	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading env file")
	}

	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello Fibre")
	})

	log.Fatal(app.Listen(":8080"))

type BillAnalysisResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

func callGeminiApi(imageBytes, apiKey, mimeType string) (map[string]interface{}, error) {
	requestBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: "Extract all food items from this restaurant bill and provide their estimated calorie counts. Format the response as a JSON object with food items as keys and calorie counts as values.",
					},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     imageBytes,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// send api request
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini returned an error : %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(respbody, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return result, nil

}
