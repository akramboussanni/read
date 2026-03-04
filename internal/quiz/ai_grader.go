package quiz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
)

// AIGrader handles AI-powered grading of quiz answers
type AIGrader struct {
	apiKey string
	client *http.Client
}

// NewAIGrader creates a new AI grader
func NewAIGrader() *AIGrader {
	return &AIGrader{
		apiKey: os.Getenv("GOOGLE_API_KEY"),
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig struct {
		CandidateCount int     `json:"candidateCount"`
		Temperature    float64 `json:"temperature"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type GradingResult struct {
	IsCorrect       bool   `json:"is_correct"`
	NeedsMoreDetail bool   `json:"needs_more_detail"`
	Explanation     string `json:"explanation"`
}

// verdict encodes the three-way outcome as an int for consensus logic:
// 2=correct, 1=needs_more_detail, 0=wrong
func (g GradingResult) verdict() int {
	if g.IsCorrect {
		return 2
	}
	if g.NeedsMoreDetail {
		return 1
	}
	return 0
}

// splitVariants splits a multi-value target_text into individual cleaned variants.
// Handles formats like:
//
//	"Notre, nous"              → ["notre", "nous"]
//	"Son, le, lui (masc. sg.)" → ["son", "le", "lui"]
//	"Où / où que"              → ["où", "où que"]
func splitVariants(s string) []string {
	// Split on commas and slashes
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '/'
	})

	var variants []string
	for _, v := range raw {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			variants = append(variants, v)
		}
	}
	return variants
}

// buildPrompt builds the structured grading prompt with concrete examples.
func buildPrompt(question, correctAnswer, userAnswer string) string {
	return fmt.Sprintf(`Tu es un correcteur expert en linguistique Arabe-Français pour un quiz de vocabulaire coranique.

Ta mission : décider si la réponse de l'étudiant capture PRÉCISÉMENT le sens du mot attendu.

═══════════════════════════════
CONTEXTE
═══════════════════════════════
Question posée  : %s
Réponse attendue: %s   ← Liste de VARIANTES (synonymes ou formes différentes).
Réponse étudiant: %s

═══════════════════════════════
RÈGLES D'OR DE BIENVEILLANCE
═══════════════════════════════

1. SYNONYMES & PRONOMS : 
   - Soyez très indulgent sur les nuances "Tu / Toi / Toi-même". Si le pronom est correct (ex: 2ème pers. masc.), c'est CORRECT.
   - Les synonymes évidents (ex: "ouvrir" vs "déverrouiller") sont acceptés.

2. ABRÉVIATIONS & GENRE :
   - "masc", "masc.", "m" = Masculin.
   - "fem", "fém.", "f" = Féminin.
   - Si l'étudiant a écrit "(masc)" ou "(fem)", il a DÉJÀ précisé le genre. NE PAS demander de précision sémantique s'il l'a déjà fournie.

3. RÈGLES DE DÉCISION :

   ✅ CORRECT (is_correct=true, needs_more_detail=false) :
   - Correspondance sémantique forte avec l'une des variantes.
   - Fautes de frappe mineures acceptées.
   - Exemple : "toi (masc)" ✓ pour "Tu (masc.)" 

   ⚠️ PRÉCISION MANQUANTE (is_correct=true, needs_more_detail=true) :
   - Le sens est bon mais une distinction CRUCIALE manque (ex: Duel en arabe confondu avec Pluriel).
   - NE DEMANDEZ PAS de précision si l'étudiant l'a déjà écrite dans sa réponse (même abrégée).
   - Explication : "Genre requis" ou "Précisez nombre".

   ❌ FAUX (is_correct=false, needs_more_detail=false) :
   - Mauvaise racine, contresens total ou hors sujet.

═══════════════════════════════
RÉPONSE (JSON pur, sans markdown)
═══════════════════════════════
{
  "is_correct": boolean,
  "needs_more_detail": boolean,
  "explanation": "Pédagogique, 6 mots max. Ex: 'Correct !' | 'Précisez le genre' | 'C'est un duel'."
}`, question, correctAnswer, userAnswer)
}

// parseCandidate extracts and parses a GradingResult from a raw candidate text.
func parseCandidate(text string) (GradingResult, error) {
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var result GradingResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return GradingResult{}, fmt.Errorf("parse error: %w (raw: %s)", err, text)
	}
	return result, nil
}

// GradeAnswer grades an answer using Gemini AI with candidateCount=2 self-consistency.
// A single API call returns two independent completions; consensus logic picks the winner.
// Returns: isCorrect, needsMoreDetail, explanation, error
func (g *AIGrader) GradeAnswer(ctx context.Context, question, correctAnswer, userAnswer string) (isCorrect bool, needsMoreDetail bool, explanation string, err error) {
	if g.apiKey == "" {
		return false, false, "", fmt.Errorf("GOOGLE_API_KEY not set")
	}

	if strings.TrimSpace(userAnswer) == "" {
		return false, false, "Réponse vide", nil
	}

	// Fast path: the correct answer can list multiple valid forms separated by
	// commas, slashes, or parentheses (e.g. "Notre, nous" or "Son, le, lui (masc. sg.)").
	// Check if the user's answer matches ANY of the listed variants.
	userTrimmed := strings.ToLower(strings.TrimSpace(userAnswer))
	for _, variant := range splitVariants(correctAnswer) {
		if strings.EqualFold(userTrimmed, variant) {
			return true, false, "Correspondance exacte ✓", nil
		}
	}

	prompt := buildPrompt(question, correctAnswer, userAnswer)

	reqBody := geminiRequest{}
	reqBody.Contents = append(reqBody.Contents, struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}{
		// Send the prompt twice — repeating the instruction in the same call
		// boosts model accuracy and consistency of the JSON output.
		Parts: []struct {
			Text string `json:"text"`
		}{{Text: prompt}, {Text: prompt}},
	})
	// Request 2 independent completions in one call — self-consistency / accuracy boost
	reqBody.GenerationConfig.CandidateCount = 2
	reqBody.GenerationConfig.Temperature = 0.2 // low temperature for reliable JSON output

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, false, "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-lite:generateContent?key=%s", g.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, false, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return false, false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, false, "", fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return false, false, "", err
	}

	if len(geminiResp.Candidates) == 0 {
		return false, false, "", fmt.Errorf("empty response from gemini")
	}

	// Parse all candidates we got back
	var results []GradingResult
	for i, candidate := range geminiResp.Candidates {
		if len(candidate.Content.Parts) == 0 {
			continue
		}
		r, err := parseCandidate(candidate.Content.Parts[0].Text)
		if err != nil {
			applog.Errorf("AI grading candidate %d parse error: %v", i, err)
			continue
		}
		results = append(results, r)
	}

	if len(results) == 0 {
		return false, false, "", fmt.Errorf("no valid candidates parsed")
	}

	// Only one candidate parsed successfully — use it directly
	if len(results) == 1 {
		r := results[0]
		applog.Infof("AI Grading (1 candidate) '%s': correct=%v needsDetail=%v expl=%s", userAnswer, r.IsCorrect, r.NeedsMoreDetail, r.Explanation)
		return r.IsCorrect, r.NeedsMoreDetail, r.Explanation, nil
	}

	// Two candidates — apply consensus:
	// Verdicts: 2=correct, 1=needs_more_detail, 0=wrong
	v0, v1 := results[0].verdict(), results[1].verdict()
	applog.Infof("AI Grading (2 candidates) '%s': c1=%d(%s) c2=%d(%s)", userAnswer, v0, results[0].Explanation, v1, results[1].Explanation)

	if v0 == v1 {
		// Full agreement → high-confidence, use candidate 0's explanation
		return results[0].IsCorrect, results[0].NeedsMoreDetail, results[0].Explanation, nil
	}

	// Disagreement: pick the more cautious (lower) verdict
	// correct(2) > needs_more_detail(1) > wrong(0)
	// Rationale: when the model is unsure, better to prompt for more detail
	// than to give undeserved points or unfairly mark wrong.
	chosen := results[0]
	if results[1].verdict() < results[0].verdict() {
		chosen = results[1]
	}

	applog.Infof("AI Grading consensus (disagreement→cautious): correct=%v needsDetail=%v expl=%s", chosen.IsCorrect, chosen.NeedsMoreDetail, chosen.Explanation)
	return chosen.IsCorrect, chosen.NeedsMoreDetail, chosen.Explanation, nil
}
