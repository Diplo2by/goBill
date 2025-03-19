package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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
	app.Post("/api/analyze-bill", analyzeBillHandler)

	log.Fatal(app.Listen(":8080"))
	log.Printf("Server starting on port 8080...\n")
}

func analyzeBillHandler(c fiber.Ctx) error {
	apikey := os.Getenv("GEMINI_API_KEY")

	// load API Key
	if apikey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(BillAnalysisResponse{
			Success: false,
			Error:   "Gemini API Key error",
		})
	}

	// get uploaded image

	file, err := c.FormFile("image")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BillAnalysisResponse{
			Success: false,
			Error:   "Error retrieving image file",
		})
	}

	fileHandle, err := file.Open()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BillAnalysisResponse{
			Success: false,
			Error:   "Failed to open image file",
		})
	}

	defer fileHandle.Close()

	fileBytes, err := io.ReadAll(fileHandle)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BillAnalysisResponse{
			Success: false,
			Error:   "Failed to read image file",
		})
	}

	// Convert to base64
	encodedImage := base64.StdEncoding.EncodeToString(fileBytes)

	// get image type
	contentType := file.Header["Content-Type"][0]

	// make Gemini API call
	geminiResponse, err := callGeminiApi(encodedImage, apikey, contentType)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BillAnalysisResponse{
			Success: false,
			Error:   fmt.Sprintf("Error calling Gemini API: %v", err),
		})
	}

	return c.JSON(BillAnalysisResponse{
		Success: true,
		Data:    geminiResponse,
	})

}

type BillAnalysisResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func callGeminiApi(imageBytes, apiKey, mimeType string) (interface{}, error) {
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

	var geminiResponse GeminiResponse
	if err := json.Unmarshal(respbody, &geminiResponse); err != nil {
		return nil, fmt.Errorf("error parsing Gemini response: %v", err)
	}
	if len(geminiResponse.Candidates) == 0 || len(geminiResponse.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no valid response from Gemini")
	}
	responseText := geminiResponse.Candidates[0].Content.Parts[0].Text
	jsonData, err = extractJSONFromText(responseText)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON data: %v", err)
	}

	// Parse the extracted JSON
	var parsedData interface{}
	if err := json.Unmarshal(jsonData, &parsedData); err != nil {
		return nil, fmt.Errorf("error parsing calorie data: %v", err)
	}

	return parsedData, nil

}

func extractJSONFromText(text string) ([]byte, error) {
	// Check if text contains markdown code block
	re := regexp.MustCompile("```(?:json)?\\s*((?s:.+?))\\s*```")
	matches := re.FindStringSubmatch(text)

	if len(matches) > 1 {
		// Extract content from code block
		return []byte(strings.TrimSpace(matches[1])), nil
	}

	// If no code block, try to find JSON directly
	// Look for text that starts with { and ends with }
	re = regexp.MustCompile("(?s:\\{.+\\})")
	matches = re.FindStringSubmatch(text)

	if len(matches) > 0 {
		return []byte(matches[0]), nil
	}

	return nil, fmt.Errorf("no valid JSON found in the response")
}
