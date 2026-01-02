package quiz

import (
	"encoding/json"
	"fmt"
)

// QuranWordsImporter handles the Quranic vocabulary format
type QuranWordsImporter struct{}

func (i *QuranWordsImporter) CanHandle(filename string, rawData []byte) bool {
	// Check if it has the expected structure
	var data struct {
		Deck struct {
			ID string `json:"id"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
	}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return false
	}
	return data.Deck.ID != "" && data.Categories != nil
}

func (i *QuranWordsImporter) Parse(rawData []byte) (*ParsedDeck, error) {
	var data struct {
		Deck struct {
			ID      string `json:"id"`
			Title   string `json:"title"`
			Version int    `json:"version"`
		} `json:"deck"`
		Categories map[string]string `json:"cats"`
		Questions  []struct {
			ID       string `json:"id"`
			Category string `json:"cat"`
			French   string `json:"fr"`
			Arabic   string `json:"ar"`
		} `json:"items"`
	}

	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	parsed := &ParsedDeck{
		DeckKey:    data.Deck.ID,
		Title:      data.Deck.Title,
		Version:    data.Deck.Version,
		Categories: data.Categories,
		Questions:  make([]ParsedQuestion, 0, len(data.Questions)),
	}

	for _, q := range data.Questions {
		parsed.Questions = append(parsed.Questions, ParsedQuestion{
			ID:            q.ID,
			CategoryKey:   q.Category,
			QuestionText:  q.Arabic,
			CorrectAnswer: q.French,
			AdditionalData: map[string]string{
				"arabic": q.Arabic,
				"french": q.French,
			},
		})
	}

	return parsed, nil
}
