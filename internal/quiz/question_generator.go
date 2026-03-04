package quiz

import (
	"fmt"
	"math/rand"
	"strings"
)

// PromptConfig defines how a question should be phrased
type PromptConfig struct {
	Default     string
	WithContext string
}

// vocabularyPrompts holds the standard templates for each direction and question type
var vocabularyPrompts = map[Direction]map[QuestionType]PromptConfig{
	DirectionSourceToTarget: {
		QuestionTypeTranslate: {
			Default:     "Traduisez ce mot {{source_lang}} en {{target_lang}} : {{source}}",
			WithContext: "Dans le contexte de '{{context}}', comment traduisez-vous '{{source}}' ({{source_lang}}) ?",
		},
		QuestionTypeWriteWord: {
			Default:     "Comment écrit-on ceci en {{target_lang}} : {{source}}",
			WithContext: "Comment écrit-on '{{source}}' (trouvé dans '{{context}}') en {{target_lang}} ?",
		},
		QuestionTypeMCQ: {
			Default:     "Choisissez la bonne traduction pour le mot {{source_lang}} '{{source}}' :",
			WithContext: "Choisissez la bonne traduction pour le mot {{source_lang}} '{{source}}' tel qu'utilisé dans '{{context}}' :",
		},
	},
	DirectionTargetToSource: {
		QuestionTypeTranslate: {
			Default:     "Traduisez ce mot {{target_lang}} en {{source_lang}} : {{target}}",
			WithContext: "Traduisez '{{target}}' ({{target_lang}}) en {{source_lang}} (sachant qu'il signifie '{{context_translation}}' ici) :",
		},
		QuestionTypeWriteWord: {
			Default: "Comment écrit-on ceci en {{source_lang}} : {{target}}",
		},
		QuestionTypeMCQ: {
			Default: "Choisissez la bonne traduction pour le mot {{target_lang}} '{{target}}' :",
		},
	},
	DirectionIdentifyGrammar: {
		QuestionTypeTranslate: {
			Default:     "Quelle est la signification de '{{source}}' ({{class}}) ?",
			WithContext: "Quelle est la signification de '{{source}}' ({{class}}) quand il est attaché à '{{context}}' ?",
		},
		QuestionTypeMCQ: {
			Default:     "Comment traduit-on le {{class}} '{{source}}' ?",
			WithContext: "Traduisez le {{class}} '{{source}}' (contexte: '{{context}}') :",
		},
	},
	DirectionAttachSuffix: {
		QuestionTypeWriteWord: {
			Default:     "Formez '{{context_translation}}' en attachant '{{source}}' ({{target}}) à '{{base_word}}' :",
			WithContext: "Attachez '{{source}}' ({{target}}) à '{{base_word}}' (trouvé dans '{{context}}') :",
		},
		QuestionTypeMCQ: {
			Default: "Quelle est la forme correcte de '{{base_word}}' combiné avec '{{source}}' ({{target}}) ?",
		},
	},
	DirectionConjugate: {
		QuestionTypeWriteWord: {
			Default: "Conjuguez '{{base_word}}' avec '{{source}}' ({{target}}) :",
		},
		QuestionTypeMCQ: {
			Default: "Comment dit-on '{{context_translation}}' avec '{{base_word}}' ?",
		},
	},
}

// QuestionGenerator interface for generating questions from deck entries
type QuestionGenerator interface {
	CanGenerate(questionType QuestionType, direction Direction, deckMetadata DeckMetadata) bool
	Generate(entry DeckEntry, questionType QuestionType, direction Direction, metadata DeckMetadata, getRandomAnswers func(count int, category string) []string) (*GeneratedQuestion, error)
}

// VocabularyQuestionGenerator generates questions for vocabulary decks
type VocabularyQuestionGenerator struct{}

// CanGenerate checks if this generator can handle the given type and direction
func (v *VocabularyQuestionGenerator) CanGenerate(questionType QuestionType, direction Direction, deckMetadata DeckMetadata) bool {
	if deckMetadata.DeckType != "vocabulary" {
		return false
	}

	supported := false
	for _, t := range deckMetadata.SupportedQuestionTypes {
		if t == string(questionType) {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	if len(deckMetadata.LanguagePair) < 2 {
		return false
	}

	switch direction {
	case DirectionSourceToTarget, DirectionTargetToSource, DirectionIdentifyGrammar, DirectionAttachSuffix, DirectionConjugate:
		return true
	default:
		return false
	}
}

// Generate generates a question from a deck entry
func (v *VocabularyQuestionGenerator) Generate(entry DeckEntry, questionType QuestionType, direction Direction, metadata DeckMetadata, getRandomAnswers func(count int, category string) []string) (*GeneratedQuestion, error) {
	sourceText, _ := entry.Data["source_text"].(string)
	targetText, _ := entry.Data["target_text"].(string)

	if sourceText == "" || targetText == "" {
		return nil, fmt.Errorf("entry missing source or target text")
	}

	// 1. Prepare Placeholders
	sourceLang, targetLang := "arabe", "français"
	if len(metadata.LanguagePair) >= 2 {
		sourceLang = translateLanguageName(metadata.LanguagePair[0])
		targetLang = translateLanguageName(metadata.LanguagePair[1])
	}

	placeholders := map[string]string{
		"source":      sourceText,
		"target":      targetText,
		"source_lang": sourceLang,
		"target_lang": targetLang,
		"class":       "mot",
	}

	if cl, ok := entry.Data["class"].(string); ok {
		placeholders["class"] = cl
	}
	if ctx, ok := entry.Data["context"].(string); ok {
		placeholders["context"] = ctx
	}
	if ctxTr, ok := entry.Data["context_translation"].(string); ok {
		placeholders["context_translation"] = ctxTr
	}
	if bw, ok := entry.Data["base_word"].(string); ok {
		placeholders["base_word"] = bw
	}

	// 2. Select Template
	template := v.selectTemplate(entry, questionType, direction)
	if template == "" {
		return nil, fmt.Errorf("could not find template for %s / %s", direction, questionType)
	}

	// 3. Format Question & Determine Answer
	questionText := replacePlaceholders(template, placeholders)
	correctAnswer := targetText
	if direction == DirectionTargetToSource {
		correctAnswer = sourceText
	} else if direction == DirectionAttachSuffix || direction == DirectionConjugate {
		// For construction/conjugation, the answer is the result word
		if ctx, ok := entry.Data["context"].(string); ok {
			correctAnswer = ctx
		} else {
			return nil, fmt.Errorf("direction %s requires 'context' field representing the combined word", direction)
		}
	}

	generated := &GeneratedQuestion{
		QuestionText:  questionText,
		CorrectAnswer: correctAnswer,
		QuestionType:  questionType,
		Direction:     direction,
	}

	// 4. Handle MCQ Options
	if questionType == QuestionTypeMCQ {
		wrongAnswers := getRandomAnswers(3, entry.Category)
		var distractors []string
		for _, ans := range wrongAnswers {
			if ans != correctAnswer {
				distractors = append(distractors, ans)
			}
		}

		// Ensure we have 3 distractors
		for len(distractors) < 3 {
			distractors = append(distractors, fmt.Sprintf("Alternative %d", len(distractors)+1))
		}
		if len(distractors) > 3 {
			distractors = distractors[:3]
		}

		generated.Options = append([]string{correctAnswer}, distractors...)
		rand.Shuffle(len(generated.Options), func(i, j int) {
			generated.Options[i], generated.Options[j] = generated.Options[j], generated.Options[i]
		})
	}

	return generated, nil
}

func (v *VocabularyQuestionGenerator) selectTemplate(entry DeckEntry, qType QuestionType, direction Direction) string {
	// Priority 1: Per-entry prompt override in JSON
	if customData, ok := entry.Data["prompts"].(map[string]interface{}); ok {
		if dirPrompts, ok := customData[string(direction)].(map[string]interface{}); ok {
			if tmpl, ok := dirPrompts[string(qType)].(string); ok {
				return tmpl
			}
		}
	}

	// Priority 2: Standard templates
	config, ok := vocabularyPrompts[direction][qType]
	if !ok {
		// Fallback to basic direction if type is missing? (not ideal but safe)
		return ""
	}

	contextVal, _ := entry.Data["context"].(string)
	if contextVal != "" && config.WithContext != "" {
		return config.WithContext
	}

	return config.Default
}

func translateLanguageName(lang string) string {
	switch strings.ToLower(lang) {
	case "arabic":
		return "arabe"
	case "french":
		return "français"
	case "english":
		return "anglais"
	default:
		return lang
	}
}

func replacePlaceholders(tmpl string, placeholders map[string]string) string {
	res := tmpl
	for k, v := range placeholders {
		res = strings.ReplaceAll(res, "{{"+k+"}}", v)
	}
	return res
}

// QuestionGeneratorRegistry manages question generators
type QuestionGeneratorRegistry struct {
	generators []QuestionGenerator
}

// NewQuestionGeneratorRegistry creates a new registry
func NewQuestionGeneratorRegistry() *QuestionGeneratorRegistry {
	return &QuestionGeneratorRegistry{
		generators: []QuestionGenerator{
			&VocabularyQuestionGenerator{},
		},
	}
}

// GetGeneratorForDeck returns the appropriate generator for a deck
func (r *QuestionGeneratorRegistry) GetGeneratorForDeck(deckMetadata DeckMetadata) QuestionGenerator {
	for _, gen := range r.generators {
		for _, qt := range deckMetadata.SupportedQuestionTypes {
			if gen.CanGenerate(QuestionType(qt), DirectionSourceToTarget, deckMetadata) {
				return gen
			}
		}
	}
	return nil
}
