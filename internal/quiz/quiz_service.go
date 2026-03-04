package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/akramboussanni/gocode/internal/applog"
	"github.com/akramboussanni/gocode/internal/model"
	"github.com/akramboussanni/gocode/internal/repo"
	"github.com/akramboussanni/gocode/internal/utils"
)

// QuizService handles quiz generation, execution, and progression
type QuizService struct {
	repos     *repo.Repos
	generator *QuizGenerator
	aiGrader  *AIGrader
}

// NewQuizService creates a new quiz service
func NewQuizService(repos *repo.Repos) *QuizService {
	return &QuizService{
		repos:     repos,
		generator: NewQuizGenerator(),
		aiGrader:  NewAIGrader(),
	}
}

// StartQuiz starts a new quiz attempt, generating questions from templates
func (s *QuizService) StartQuiz(ctx context.Context, userID int64, quizID int64, courseID *string, nodeID *string, assignmentID *int64) (*model.QuizAttempt, []*model.AttemptQuestion, error) {
	// 1. Get the quiz
	quizModel, err := s.repos.Quiz.GetByID(ctx, quizID)
	if err != nil {
		return nil, nil, fmt.Errorf("quiz not found: %w", err)
	}

	// 2. If part of a course, check that user can access this node
	if nodeID != nil && courseID != nil {
		canAccess := true
		if assignmentID != nil {
			// Bypass course tree progression if starting from an assignment
			canAccess = true
		} else {
			canAccess, err = s.CanAccessNode(ctx, userID, *courseID, *nodeID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to check node access: %w", err)
			}
		}

		if !canAccess {
			return nil, nil, fmt.Errorf("node is locked: complete prerequisite nodes first")
		}
	}

	// 2b. If this is an assignment, enforce max_retakes limit
	if assignmentID != nil {
		asgn, err := s.repos.Classroom.GetAssignmentByID(ctx, *assignmentID)
		if err == nil && asgn.MaxRetakes >= 0 {
			completed, _ := s.repos.QuizAttempt.CountCompletedByAssignmentAndUser(ctx, *assignmentID, userID)
			// MaxRetakes=0 means 1 attempt total (no retakes), MaxRetakes=N means N extra attempts
			maxAllowed := asgn.MaxRetakes + 1
			if completed >= maxAllowed {
				return nil, nil, fmt.Errorf("retake limit reached: %d/%d attempts used", completed, maxAllowed)
			}
		}
	}

	// 3. Check for active attempt, resume if exists
	existing, err := s.repos.QuizAttempt.GetActiveByUserAndQuiz(ctx, userID, quizID)
	if err == nil && existing != nil {
		questions, err := s.repos.AttemptQuestion.GetByAttemptID(ctx, existing.ID)
		if err == nil && len(questions) > 0 {
			return existing, questions, nil
		}
	}

	// 4. Create new attempt
	now := time.Now().Unix()
	attempt := &model.QuizAttempt{
		ID:           utils.GenerateSnowflakeID(),
		UserID:       userID,
		QuizID:       quizID,
		CourseID:     courseID,
		NodeID:       nodeID,
		AssignmentID: assignmentID,
		StartedAt:    now,
	}
	if err := s.repos.QuizAttempt.Create(ctx, attempt); err != nil {
		return nil, nil, fmt.Errorf("failed to create attempt: %w", err)
	}

	// 5. Generate questions from templates
	questions, err := s.GenerateQuestionsForAttempt(ctx, quizModel, attempt.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	// 6. Persist generated questions
	for _, q := range questions {
		if err := s.repos.AttemptQuestion.Create(ctx, q); err != nil {
			applog.Errorf("Failed to save question: %v", err)
		}
	}

	return attempt, questions, nil
}

// GenerateQuestionsForAttempt generates questions based on quiz templates
func (s *QuizService) GenerateQuestionsForAttempt(ctx context.Context, quizModel *model.Quiz, attemptID int64) ([]*model.AttemptQuestion, error) {
	templates, err := s.repos.QuestionTemplate.GetByQuizID(ctx, quizModel.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no question templates configured for this quiz")
	}

	var allQuestions []*model.AttemptQuestion
	displayOrder := 0
	now := time.Now().Unix()

	for _, tmpl := range templates {
		switch tmpl.GenerationMode {
		case "random_from_deck":
			questions, err := s.generateFromDeck(ctx, tmpl, quizModel, attemptID, &displayOrder, now)
			if err != nil {
				applog.Errorf("Template %d generation failed: %v", tmpl.ID, err)
				continue
			}
			allQuestions = append(allQuestions, questions...)

		case "manual":
			questions, err := s.generateManual(tmpl, quizModel, attemptID, &displayOrder, now)
			if err != nil {
				applog.Errorf("Manual template %d failed: %v", tmpl.ID, err)
				continue
			}
			allQuestions = append(allQuestions, questions...)

		case "llm":
			// Future: LLM-powered generation
			applog.Warnf("LLM generation not yet implemented for template %d", tmpl.ID)
		}
	}

	// Shuffle if configured
	if quizModel.ShuffleQuestions {
		rand.Shuffle(len(allQuestions), func(i, j int) {
			allQuestions[i], allQuestions[j] = allQuestions[j], allQuestions[i]
		})
		for i := range allQuestions {
			allQuestions[i].DisplayOrder = i
		}
	}

	return allQuestions, nil
}

// generateFromDeck generates questions by pulling random entries from a deck/category
func (s *QuizService) generateFromDeck(ctx context.Context, tmpl *model.QuestionTemplate, quizModel *model.Quiz, attemptID int64, displayOrder *int, now int64) ([]*model.AttemptQuestion, error) {
	if tmpl.DeckID == nil {
		return nil, fmt.Errorf("template has no deck_id")
	}

	// Parse template config
	var questionTypes []string
	if err := json.Unmarshal([]byte(tmpl.QuestionTypes), &questionTypes); err != nil {
		questionTypes = []string{"mcq", "translate"}
	}
	var directions []string
	if err := json.Unmarshal([]byte(tmpl.Directions), &directions); err != nil {
		directions = []string{"source_to_target", "target_to_source"}
	}

	// Temporarily disable French -> Arabic (target_to_source)
	var filtered []string
	for _, d := range directions {
		if d != "target_to_source" {
			filtered = append(filtered, d)
		}
	}
	if len(filtered) == 0 {
		filtered = []string{"source_to_target"}
	}
	directions = filtered

	// Get random entries from the category (or whole deck)
	var entries []*model.DeckEntry
	var err error
	if tmpl.CategoryID != nil {
		entries, err = s.repos.DeckEntry.GetRandomByCategoryID(ctx, *tmpl.CategoryID, tmpl.QuestionCount)
	} else {
		// TODO: get random from entire deck
		return nil, fmt.Errorf("deck-wide random not yet implemented")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Get deck metadata for the generator
	deck, err := s.repos.Deck.GetByID(ctx, *tmpl.DeckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}

	var deckMeta DeckMetadata
	if err := json.Unmarshal([]byte(deck.DeckMetadata), &deckMeta); err != nil {
		applog.Errorf("Failed to parse deck metadata: %v", err)
	}

	generator := s.generator.registry.GetGeneratorForDeck(deckMeta)
	if generator == nil {
		return nil, fmt.Errorf("no generator found for deck type %s", deckMeta.DeckType)
	}

	// Map category keys and IDs for the generator
	dbCats, _ := s.repos.Category.GetByDeckID(ctx, *tmpl.DeckID)
	catIDToKey := make(map[int64]string)
	catKeyToID := make(map[string]int64)
	for _, c := range dbCats {
		catIDToKey[c.ID] = c.CategoryKey
		catKeyToID[c.CategoryKey] = c.ID
	}

	// Helper to get distractor answers
	getRandomAnswers := func(count int, categoryKey string) []string {
		var catID int64
		// 1. Try to resolve the specific category requested by the generator
		if id, ok := catKeyToID[categoryKey]; ok {
			catID = id
		} else if tmpl.CategoryID != nil {
			// 2. Fallback to the template's primary category
			catID = *tmpl.CategoryID
		}

		var ans []string
		if catID > 0 {
			ans, _ = s.repos.DeckEntry.GetRandomAnswersFromCategory(ctx, catID, 30)
		}

		// 3. Fallback to the whole deck if the category is too small
		if len(ans) < 5 && tmpl.DeckID != nil {
			deckAns, _ := s.repos.DeckEntry.GetRandomAnswersFromDeck(ctx, *tmpl.DeckID, 30)
			ans = append(ans, deckAns...)
		}

		if len(ans) == 0 {
			applog.Warnf("Distractor generation failed for category '%s' in deck %d", categoryKey, *tmpl.DeckID)
		}

		return ans
	}

	var questions []*model.AttemptQuestion

	for _, entry := range entries {
		var entryData map[string]interface{}
		if err := json.Unmarshal([]byte(entry.EntryData), &entryData); err != nil {
			continue
		}

		catKey := ""
		if k, ok := catIDToKey[entry.CategoryID]; ok {
			catKey = k
		}

		dEntry := DeckEntry{
			ID:       entry.EntryKey,
			Category: catKey,
			Data:     entryData,
		}

		// Pick random question type and direction
		qt := QuestionType(questionTypes[rand.Intn(len(questionTypes))])
		dir := Direction(directions[rand.Intn(len(directions))])

		// Inject creative directions (20% chance)
		rng := rand.Float32()
		if rng < 0.20 {
			if bw, ok := entryData["base_word"].(string); ok && bw != "" {
				dir = DirectionAttachSuffix
				// Use Conjugate for verb-heavy categories
				if strings.Contains(strings.ToLower(catKey), "verb") {
					dir = DirectionConjugate
				}
			} else {
				dir = DirectionIdentifyGrammar
			}
		}

		genQ, err := generator.Generate(dEntry, qt, dir, deckMeta, getRandomAnswers)
		if err != nil {
			// Fallback to standard direction
			dir = DirectionSourceToTarget
			genQ, err = generator.Generate(dEntry, qt, dir, deckMeta, getRandomAnswers)
			if err != nil {
				continue
			}
		}

		optsJSON, _ := json.Marshal(genQ.Options)
		dirStr := string(genQ.Direction)

		q := &model.AttemptQuestion{
			ID:             utils.GenerateSnowflakeID(),
			AttemptID:      attemptID,
			QuizID:         quizModel.ID,
			QuestionText:   genQ.QuestionText,
			CorrectAnswer:  genQ.CorrectAnswer,
			Options:        string(optsJSON),
			QuestionType:   string(genQ.QuestionType),
			Direction:      &dirStr,
			DisplayOrder:   *displayOrder,
			SourceEntryID:  &entry.ID,
			GenerationMode: "random_from_deck",
			CreatedAt:      now,
		}
		questions = append(questions, q)
		*displayOrder++
	}

	return questions, nil
}

// generateManual generates questions from manually defined data in the template
func (s *QuizService) generateManual(tmpl *model.QuestionTemplate, quizModel *model.Quiz, attemptID int64, displayOrder *int, now int64) ([]*model.AttemptQuestion, error) {
	if tmpl.ManualData == nil || *tmpl.ManualData == "" {
		return nil, fmt.Errorf("no manual data")
	}

	var manualQuestions []struct {
		QuestionText  string   `json:"question_text"`
		CorrectAnswer string   `json:"correct_answer"`
		Options       []string `json:"options"`
		QuestionType  string   `json:"question_type"`
		Direction     string   `json:"direction"`
	}

	if err := json.Unmarshal([]byte(*tmpl.ManualData), &manualQuestions); err != nil {
		return nil, fmt.Errorf("failed to parse manual data: %w", err)
	}

	var questions []*model.AttemptQuestion
	for _, mq := range manualQuestions {
		optsJSON, _ := json.Marshal(mq.Options)
		q := &model.AttemptQuestion{
			ID:             utils.GenerateSnowflakeID(),
			AttemptID:      attemptID,
			QuizID:         quizModel.ID,
			QuestionText:   mq.QuestionText,
			CorrectAnswer:  mq.CorrectAnswer,
			Options:        string(optsJSON),
			QuestionType:   mq.QuestionType,
			Direction:      &mq.Direction,
			DisplayOrder:   *displayOrder,
			GenerationMode: "manual",
			CreatedAt:      now,
		}
		questions = append(questions, q)
		*displayOrder++
	}

	return questions, nil
}

// SubmitAnswer grades a single answer and returns feedback.
// Returns: answer, needsMoreDetail, error
// When needsMoreDetail=true, the answer is NOT persisted — the user should refine their answer.
func (s *QuizService) SubmitAnswer(ctx context.Context, userID int64, attemptID int64, questionID int64, userAnswer string) (*model.UserAnswer, bool, error) {
	// Verify attempt belongs to user
	attempt, err := s.repos.QuizAttempt.GetByID(ctx, attemptID)
	if err != nil {
		return nil, false, fmt.Errorf("attempt not found: %w", err)
	}
	if attempt.UserID != userID {
		return nil, false, fmt.Errorf("unauthorized")
	}
	if attempt.CompletedAt != nil {
		return nil, false, fmt.Errorf("attempt already completed")
	}

	// Get the question
	question, err := s.repos.AttemptQuestion.GetByID(ctx, questionID)
	if err != nil {
		return nil, false, fmt.Errorf("question not found: %w", err)
	}
	if question.AttemptID != attemptID {
		return nil, false, fmt.Errorf("question does not belong to this attempt")
	}

	// Grade the answer
	isCorrect := false
	needsMoreDetail := false
	var aiExplanation *string
	var points float64

	if question.QuestionType == "mcq" {
		// For MCQ, compare the userAnswer with the correct answer text
		isCorrect = strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(question.CorrectAnswer))

		if !isCorrect {
			// Try to resolve as 1-based index
			idx, err := strconv.Atoi(userAnswer)
			if err == nil {
				var options []string
				if err := json.Unmarshal([]byte(question.Options), &options); err == nil {
					if idx > 0 && idx <= len(options) {
						selectedText := options[idx-1]
						isCorrect = strings.EqualFold(strings.TrimSpace(selectedText), strings.TrimSpace(question.CorrectAnswer))
					}
				}
			}
		}
	} else {
		// Check if this is a re-submission after a hint
		prevAnswers, _ := s.repos.UserAnswer.GetByAttemptID(ctx, attemptID)
		hasPrevAttempt := false
		for _, a := range prevAnswers {
			if a.QuestionID == questionID {
				hasPrevAttempt = true
				break
			}
		}

		// For translate/write_word, use AI grading
		correct, needsDetail, explanation, err := s.aiGrader.GradeAnswer(ctx, question.QuestionText, question.CorrectAnswer, userAnswer)
		if err != nil {
			// Fallback to exact match on AI error
			isCorrect = strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(question.CorrectAnswer))
		} else {
			isCorrect = correct
			needsMoreDetail = needsDetail
			aiExplanation = &explanation

			// AI-powered "hint" logic
			if needsMoreDetail {
				if hasPrevAttempt {
					// Second vague attempt: "if u resubmit somthng vague then get 0"
					isCorrect = false
					needsMoreDetail = false // Lock it now
				} else {
					// First vague attempt: "still gives 1 point" and allow 1 redo
					isCorrect = true
					// needsMoreDetail remains true (provided by AI)
				}
			}
		}
	}

	if isCorrect {
		points = 1.0
	} else {
		points = 0.0
	}

	answer := &model.UserAnswer{
		ID:            utils.GenerateSnowflakeID(),
		AttemptID:     attemptID,
		QuestionID:    questionID,
		UserAnswer:    userAnswer,
		IsCorrect:     isCorrect,
		PointsEarned:  points,
		AIExplanation: aiExplanation,
		AnsweredAt:    time.Now().Unix(),
	}

	if err := s.repos.UserAnswer.Create(ctx, answer); err != nil {
		return nil, false, fmt.Errorf("failed to save answer: %w", err)
	}

	// Always return true for needsMoreDetail if the AI flagged it,
	// so the frontend can prompt for a redo, even though we already saved the point.
	return answer, needsMoreDetail, nil
}

// CompleteQuiz completes a quiz attempt and calculates score
func (s *QuizService) CompleteQuiz(ctx context.Context, userID int64, attemptID int64) (*QuizResult, error) {
	attempt, err := s.repos.QuizAttempt.GetByID(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("attempt not found: %w", err)
	}
	if attempt.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Get all questions and answers
	questions, err := s.repos.AttemptQuestion.GetByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	answers, err := s.repos.UserAnswer.GetByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	// Calculate score
	var score, maxScore float64
	maxScore = float64(len(questions))
	answerMap := make(map[int64]*model.UserAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	for _, a := range answerMap {
		if a.IsCorrect {
			score++
		}
	}

	percentage := 0.0
	if maxScore > 0 {
		percentage = (score / maxScore) * 100
	}

	// Check pass
	quiz, _ := s.repos.Quiz.GetByID(ctx, attempt.QuizID)
	passPercentage := 70
	if quiz != nil && quiz.PassPercentage != nil {
		passPercentage = *quiz.PassPercentage
	}
	passed := percentage >= float64(passPercentage)

	// Calculate coins
	coinsEarned := 0
	if passed && quiz != nil && quiz.GivesCoins {
		coinsEarned = quiz.CoinReward
	}

	timeTaken := int(time.Now().Unix() - attempt.StartedAt)

	// Update attempt
	if err := s.repos.QuizAttempt.CompleteAttempt(ctx, attemptID, score, maxScore, percentage, passed, timeTaken, coinsEarned); err != nil {
		return nil, fmt.Errorf("failed to complete attempt: %w", err)
	}

	// Record assignment completion regardless of pass/fail — homework is "done" once submitted
	if attempt.AssignmentID != nil {
		ac := &model.AssignmentCompletion{
			ID:           utils.GenerateSnowflakeID(),
			AssignmentID: *attempt.AssignmentID,
			StudentID:    userID,
			AttemptID:    &attemptID,
			Score:        score,
			MaxScore:     maxScore,
			Percentage:   percentage,
			Passed:       passed,
			CompletedAt:  time.Now().Unix(),
		}
		if err := s.repos.AssignmentCompletion.Create(ctx, ac); err != nil {
			applog.Errorf("Failed to record assignment completion: %v", err)
		}
	}

	// Separately: if passed and part of a course, handle course node progression
	if passed && attempt.CourseID != nil && attempt.NodeID != nil {
		if attempt.AssignmentID != nil {
			// Only mark node complete in course if it's actually unlocked in the student's progression
			canAccess, err := s.CanAccessNode(ctx, userID, *attempt.CourseID, *attempt.NodeID)
			if err == nil && canAccess {
				if err := s.CompleteNode(ctx, userID, *attempt.CourseID, *attempt.NodeID); err != nil {
					applog.Errorf("Failed to mark node as completed: %v", err)
				}
			}
			// If the node is locked, we intentionally skip CompleteNode — the assignment is done
			// but the student hasn't reached that point in the course tree yet
		} else {
			// Normal course flow (no assignment) — always complete the node
			if err := s.CompleteNode(ctx, userID, *attempt.CourseID, *attempt.NodeID); err != nil {
				applog.Errorf("Failed to mark node as completed: %v", err)
			}
		}
	}

	// Award coins
	if coinsEarned > 0 {
		coins, _ := s.repos.Coin.GetUserCoins(ctx, userID)
		coins.Balance += coinsEarned
		coins.LifetimeEarned += coinsEarned
		coins.LastUpdated = time.Now().Unix()
		_ = s.repos.Coin.CreateOrUpdateCoins(ctx, coins)

		// Record transaction
		refType := "quiz_attempt"
		descr := fmt.Sprintf("Récompense pour le quiz : %s", quiz.Title)
		_ = s.repos.Coin.AddTransaction(ctx, &model.CoinTransaction{
			ID:              utils.GenerateSnowflakeID(),
			UserID:          userID,
			Amount:          coinsEarned,
			TransactionType: "earn",
			ReferenceType:   &refType,
			ReferenceID:     &attemptID,
			Description:     &descr,
			CreatedAt:       time.Now().Unix(),
		})
	}

	// Build result
	results := make([]AnswerResult, 0, len(questions))
	for _, q := range questions {
		result := AnswerResult{
			QuestionID:    q.ID,
			QuestionText:  q.QuestionText,
			CorrectAnswer: q.CorrectAnswer,
			QuestionType:  q.QuestionType,
		}
		if a, ok := answerMap[q.ID]; ok {
			result.UserAnswer = a.UserAnswer
			result.IsCorrect = a.IsCorrect
			result.PointsEarned = a.PointsEarned
			if a.AIExplanation != nil {
				result.AIExplanation = *a.AIExplanation
			}
		}
		results = append(results, result)
	}

	return &QuizResult{
		AttemptID:   attemptID,
		Score:       score,
		MaxScore:    maxScore,
		Percentage:  percentage,
		Passed:      passed,
		CoinsEarned: coinsEarned,
		TimeTaken:   timeTaken,
		Results:     results,
	}, nil
}

// ============================================================
// COURSE PROGRESSION - TREE ENFORCEMENT
// ============================================================

// CanAccessNode checks whether a user can access a node in a course tree.
// A node is accessible if ALL required parent nodes have been completed.
func (s *QuizService) CanAccessNode(ctx context.Context, userID int64, courseID string, nodeID string) (bool, error) {
	// Teachers can access everything
	user, err := s.repos.User.GetUserByID(ctx, userID)
	if err == nil && user.Role == "teacher" {
		return true, nil
	}

	// Get the node
	node, err := s.repos.CourseNode.GetByID(ctx, nodeID)
	if err != nil {
		return false, fmt.Errorf("node not found: %w", err)
	}

	// Start nodes are always accessible
	if node.NodeType == "start" {
		return true, nil
	}

	// Get all edges pointing TO this node (parents)
	edges, err := s.repos.CourseEdge.GetByCourseID(ctx, courseID)
	if err != nil {
		return false, fmt.Errorf("failed to get edges: %w", err)
	}

	// Find all required parent node IDs
	var requiredParentIDs []string
	for _, edge := range edges {
		if edge.TargetNodeID == nodeID && edge.EdgeType == "required" {
			requiredParentIDs = append(requiredParentIDs, edge.SourceNodeID)
		}
	}

	// If no required parents, it's accessible (e.g., first node after start)
	if len(requiredParentIDs) == 0 {
		return true, nil
	}

	// Get user enrollment and check completed nodes
	enrollment, err := s.repos.Enrollment.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return false, fmt.Errorf("not enrolled: %w", err)
	}

	var completedNodes []string
	if enrollment.CompletedNodes != "" {
		if err := json.Unmarshal([]byte(enrollment.CompletedNodes), &completedNodes); err != nil {
			completedNodes = []string{}
		}
	}

	completedSet := make(map[string]bool)
	for _, cn := range completedNodes {
		completedSet[cn] = true
	}

	// ALL required parents must be completed
	for _, parentID := range requiredParentIDs {
		if !completedSet[parentID] {
			parentNode, err := s.repos.CourseNode.GetByID(ctx, parentID)
			if err != nil || parentNode.NodeType != "start" {
				return false, nil
			}
		}
	}

	return true, nil
}

// GetNodeStatuses returns the status of all nodes in a course for a user
func (s *QuizService) GetNodeStatuses(ctx context.Context, userID int64, courseID string) (map[string]NodeStatus, error) {
	nodes, err := s.repos.CourseNode.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	edges, err := s.repos.CourseEdge.GetByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// Get enrollment
	enrollment, err := s.repos.Enrollment.GetByUserAndCourse(ctx, userID, courseID)
	var completedNodes []string
	if err == nil && enrollment != nil && enrollment.CompletedNodes != "" {
		json.Unmarshal([]byte(enrollment.CompletedNodes), &completedNodes)
	}
	completedSet := make(map[string]bool)
	for _, cn := range completedNodes {
		completedSet[cn] = true
	}

	// Build parent map
	parentMap := make(map[string][]string) // nodeID -> parent nodeIDs (required)
	for _, edge := range edges {
		if edge.EdgeType == "required" {
			parentMap[edge.TargetNodeID] = append(parentMap[edge.TargetNodeID], edge.SourceNodeID)
		}
	}

	// Check if user is teacher for bypass
	user, _ := s.repos.User.GetUserByID(ctx, userID)
	isTeacher := user != nil && user.Role == "teacher"

	statuses := make(map[string]NodeStatus)
	for _, node := range nodes {
		status := NodeStatus{
			NodeID:   node.ID,
			NodeType: node.NodeType,
		}

		if completedSet[node.ID] {
			status.State = "completed"
		} else if node.NodeType == "start" {
			status.State = "completed" // start is always "done"
		} else if isTeacher {
			status.State = "unlocked"
		} else {
			// Check if all required parents are completed
			parents := parentMap[node.ID]
			allParentsDone := true
			for _, parentID := range parents {
				if !completedSet[parentID] {
					// Also allow for start nodes
					parentNode, _ := s.repos.CourseNode.GetByID(ctx, parentID)
					if parentNode == nil || parentNode.NodeType != "start" {
						allParentsDone = false
						break
					}
				}
			}

			if len(parents) == 0 || allParentsDone {
				status.State = "unlocked"
			} else {
				status.State = "locked"
			}
		}

		statuses[node.ID] = status
	}

	return statuses, nil
}

// CompleteNode marks a node as completed and updates course progress
func (s *QuizService) CompleteNode(ctx context.Context, userID int64, courseID string, nodeID string) error {
	enrollment, err := s.repos.Enrollment.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return fmt.Errorf("not enrolled: %w", err)
	}

	var completedNodes []string
	if enrollment.CompletedNodes != "" {
		json.Unmarshal([]byte(enrollment.CompletedNodes), &completedNodes)
	}

	// Check if already completed
	for _, cn := range completedNodes {
		if cn == nodeID {
			return nil // already completed
		}
	}

	completedNodes = append(completedNodes, nodeID)
	completedJSON, _ := json.Marshal(completedNodes)

	// Calculate progress
	allNodes, _ := s.repos.CourseNode.GetByCourseID(ctx, courseID)
	totalPlayableNodes := 0
	for _, n := range allNodes {
		if n.NodeType != "start" {
			totalPlayableNodes++
		}
	}
	progress := 0.0
	if totalPlayableNodes > 0 {
		progress = float64(len(completedNodes)) / float64(totalPlayableNodes)
	}

	// Update enrollment
	return s.repos.Enrollment.UpdateProgress(ctx, enrollment.ID, progress, string(completedJSON), &nodeID)
}

// ============================================================
// ONBOARDING - Auto-enroll users in default courses
// ============================================================

// CompleteOnboarding enrolls a user in selected courses and sets their active course
func (s *QuizService) CompleteOnboarding(ctx context.Context, userID int64, selectedCourseIDs []string) error {
	now := time.Now().Unix()

	var firstCourseID *string

	for _, courseID := range selectedCourseIDs {
		// Verify course exists
		course, err := s.repos.Course.GetByID(ctx, courseID)
		if err != nil || course == nil {
			continue
		}

		enrollment := &model.UserEnrollment{
			ID:             fmt.Sprintf("enroll_%d_%s", userID, courseID),
			UserID:         userID,
			CourseID:       courseID,
			Status:         "active",
			Progress:       0,
			CompletedNodes: "[]",
			EnrolledAt:     now,
			LastAccessedAt: now,
		}

		if err := s.repos.Enrollment.Create(ctx, enrollment); err != nil {
			applog.Errorf("Failed to enroll user in course %s: %v", courseID, err)
			continue
		}

		if firstCourseID == nil {
			firstCourseID = &courseID
		}
	}

	// Set active course
	if firstCourseID != nil {
		_ = s.repos.User.SetActiveCourse(ctx, userID, firstCourseID)
	}

	// Mark onboarding complete
	return s.repos.User.SetOnboardingCompleted(ctx, userID)
}

// ============================================================
// AUTO COURSE GENERATION from a deck
// ============================================================

// AutoGenerateCourse creates a course from a deck, grouping categories into quiz nodes
func (s *QuizService) AutoGenerateCourse(ctx context.Context, deckID int64, title string, createdBy *int64) (*model.Course, error) {
	deck, err := s.repos.Deck.GetByID(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("deck not found: %w", err)
	}

	categories, err := s.repos.Category.GetByDeckID(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	if len(categories) == 0 {
		return nil, fmt.Errorf("deck has no categories")
	}

	// Sort categories by display_order
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DisplayOrder < categories[j].DisplayOrder
	})

	var romanNumRegex = regexp.MustCompile(`\s+(I|II|III|IV|V|VI|VII|VIII|IX|X|[A-E]|\d+)$`)
	var parenthesesRegex = regexp.MustCompile(`\s+\(.*?\)$`)

	extractGroupName := func(title string) string {
		t := strings.ToLower(title)
		if strings.Contains(t, "pronoun") {
			return "Pronouns"
		}
		if strings.Contains(t, "preposition") {
			return "Prepositions"
		}
		if strings.Contains(t, "names of allah") {
			return "Names of Allah"
		}
		if strings.Contains(t, "resurrection") {
			return "Resurrection & Judgment"
		}
		if strings.Contains(t, "augmented verbs") {
			return "Augmented Verbs"
		}
		if strings.Contains(t, "verbs form 1") {
			return "Verbs Form 1"
		}

		t2 := parenthesesRegex.ReplaceAllString(title, "")
		t2 = romanNumRegex.ReplaceAllString(t2, "")
		return strings.TrimSpace(t2)
	}

	type CatGroup struct {
		Name       string
		Categories []*model.Category
	}

	var initialGroups []*CatGroup
	for _, cat := range categories {
		catTitle := cat.Title
		if catTitle == "" {
			catTitle = cat.CategoryKey
		}
		gName := extractGroupName(catTitle)

		val := len(initialGroups)
		if val == 0 || initialGroups[val-1].Name != gName {
			initialGroups = append(initialGroups, &CatGroup{Name: gName, Categories: []*model.Category{cat}})
		} else {
			initialGroups[val-1].Categories = append(initialGroups[val-1].Categories, cat)
		}
	}

	var groups []*CatGroup
	maxPerGroup := 4
	for _, ig := range initialGroups {
		if len(ig.Categories) <= maxPerGroup {
			groups = append(groups, ig)
			continue
		}
		parts := (len(ig.Categories) + maxPerGroup - 1) / maxPerGroup
		for p := 0; p < parts; p++ {
			start := p * maxPerGroup
			end := start + maxPerGroup
			if end > len(ig.Categories) {
				end = len(ig.Categories)
			}
			chunkName := fmt.Sprintf("%s (Partie %d)", ig.Name, p+1)
			groups = append(groups, &CatGroup{
				Name:       chunkName,
				Categories: ig.Categories[start:end],
			})
		}
	}

	now := time.Now().Unix()
	courseID := fmt.Sprintf("course_%d_%d", deckID, now)

	if title == "" {
		title = deck.Title + " - Learning Path"
	}

	course := &model.Course{
		ID:          courseID,
		Title:       title,
		Description: fmt.Sprintf("Auto-generated course from the %s deck", deck.Title),
		Icon:        "graduation-cap",
		Color:       "#6C5CE7",
		DeckID:      &deckID,
		IsDefault:   false,
		IsPublished: true,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repos.Course.Create(ctx, course); err != nil {
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	// Create start node
	startNode := &model.CourseNode{
		ID:        courseID + "_start",
		CourseID:  courseID,
		NodeType:  "start",
		Title:     "Départ",
		PositionX: 250,
		PositionY: 0,
		SortOrder: 0,
		CreatedAt: now,
	}
	if err := s.repos.CourseNode.Create(ctx, startNode); err != nil {
		return nil, fmt.Errorf("failed to create start node: %w", err)
	}

	// Create lesson and quiz nodes for each group
	prevNodeIDs := []string{startNode.ID}
	currentYPos := 0.0

	for gIdx, group := range groups {
		var milestoneID string

		if len(group.Categories) > 1 {
			currentYPos += 140.0
			// 1. Create a Milestone Node for the Group
			milestoneID = fmt.Sprintf("%s_mile_%d", courseID, gIdx)
			milestone := &model.CourseNode{
				ID:          milestoneID,
				CourseID:    courseID,
				NodeType:    "milestone",
				Title:       group.Name,
				Description: fmt.Sprintf("Chapitre : %s", group.Name),
				Icon:        "folder",
				PositionX:   250,
				PositionY:   currentYPos,
				SortOrder:   gIdx * 10,
				CreatedAt:   now,
			}
			_ = s.repos.CourseNode.Create(ctx, milestone)

			// Connect previous nodes to this milestone
			for pIdx, pID := range prevNodeIDs {
				edge := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_em_%d_%d", courseID, gIdx, pIdx),
					CourseID:     courseID,
					SourceNodeID: pID,
					TargetNodeID: milestoneID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, edge)
			}
		}

		var currentGroupEndNodes []string

		currentYPos += 140.0

		for cIdx, cat := range group.Categories {
			catTitle := cat.Title
			if catTitle == "" {
				catTitle = cat.CategoryKey
			}

			// Parallel X Offset Calculation
			offsetX := float64(cIdx) - float64(len(group.Categories)-1)/2.0
			posX := 250.0 + (offsetX * 300.0)

			// Construct the real Quiz ID first
			actualQuizID := utils.GenerateSnowflakeID()

			// Construct the Quiz Node (Child)
			quizNodeID := fmt.Sprintf("%s_q_%d_%d", courseID, gIdx, cIdx)
			quizNodeConfig := map[string]interface{}{
				"deck_id":        deckID,
				"category_id":    cat.ID,
				"quiz_id":        strconv.FormatInt(actualQuizID, 10),
				"question_types": []string{"mcq", "translate"},
				"directions":     []string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"},
				"question_count": 10,
			}
			configJSON, _ := json.Marshal(quizNodeConfig)
			configStr := string(configJSON)

			quizNode := &model.CourseNode{
				ID:          quizNodeID,
				CourseID:    courseID,
				NodeType:    "quiz",
				Title:       catTitle,
				Description: fmt.Sprintf("Maîtrise %s", catTitle),
				Icon:        "joystick",
				PositionX:   posX,
				PositionY:   currentYPos,
				SortOrder:   (gIdx * 10) + cIdx + 1,
				QuizConfig:  &configStr,
				CreatedAt:   now,
			}
			if err := s.repos.CourseNode.Create(ctx, quizNode); err != nil {
				applog.Errorf("Failed to create quiz node %s: %v", quizNodeID, err)
			}

			if milestoneID != "" {
				// Edge from Milestone to Quiz
				e1 := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_e1_%d_%d", courseID, gIdx, cIdx),
					CourseID:     courseID,
					SourceNodeID: milestoneID,
					TargetNodeID: quizNodeID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, e1)
			} else {
				// Connect directly from prevNodeIDs to the single Quiz
				for pIdx, pID := range prevNodeIDs {
					eP := &model.CourseEdge{
						ID:           fmt.Sprintf("%s_ep_%d_%d", courseID, gIdx, pIdx),
						CourseID:     courseID,
						SourceNodeID: pID,
						TargetNodeID: quizNodeID,
						EdgeType:     "required",
						CreatedAt:    now,
					}
					_ = s.repos.CourseEdge.Create(ctx, eP)
				}
			}

			// Dynamic Quiz
			quiz := &model.Quiz{
				ID:               actualQuizID,
				Title:            catTitle,
				Description:      fmt.Sprintf("Quiz: %s", catTitle),
				CourseID:         &courseID,
				NodeID:           &quizNodeID,
				DeckID:           &deckID,
				PassPercentage:   intPtr(70),
				ShuffleQuestions: true,
				QuestionMode:     "mixed",
				GivesCoins:       true,
				CoinReward:       10,
				IsPublic:         false,
				IsSystem:         true,
				IsDynamic:        true,
				CreatedBy:        createdBy,
				CreatedAt:        now,
				IsActive:         true,
			}
			if err := s.repos.Quiz.Create(ctx, quiz); err != nil {
				applog.Errorf("Failed to create generated quiz %d: %v", quiz.ID, err)
			}

			// Template
			questTypes, _ := json.Marshal([]string{"mcq", "translate"})
			dirs, _ := json.Marshal([]string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"})

			catID := cat.ID
			dID := deckID
			template := &model.QuestionTemplate{
				ID:             utils.GenerateSnowflakeID(),
				QuizID:         quiz.ID,
				DeckID:         &dID,
				CategoryID:     &catID,
				QuestionTypes:  string(questTypes),
				Directions:     string(dirs),
				GenerationMode: "random_from_deck",
				QuestionCount:  10,
				CreatedAt:      now,
			}
			if err := s.repos.QuestionTemplate.Create(ctx, template); err != nil {
				applog.Errorf("Failed to create template %d for quiz %d: %v", template.ID, quiz.ID, err)
			}

			currentGroupEndNodes = append(currentGroupEndNodes, quizNodeID)
		}

		if len(group.Categories) > 1 {
			currentYPos += 140.0

			// Synthèse Node
			synthNodeID := fmt.Sprintf("%s_synth_%d", courseID, gIdx)
			actualSynthQuizID := utils.GenerateSnowflakeID()

			synthQuestionCount := 10
			if len(group.Categories)*3 > 10 {
				synthQuestionCount = len(group.Categories) * 3
			}

			synthNodeConfig := map[string]interface{}{
				"deck_id":        deckID,
				"quiz_id":        strconv.FormatInt(actualSynthQuizID, 10),
				"question_types": []string{"mcq", "translate"},
				"directions":     []string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"},
				"question_count": synthQuestionCount,
			}
			synthConfigJSON, _ := json.Marshal(synthNodeConfig)
			synthConfigStr := string(synthConfigJSON)

			synthNode := &model.CourseNode{
				ID:          synthNodeID,
				CourseID:    courseID,
				NodeType:    "quiz",
				Title:       "Synthèse",
				Description: fmt.Sprintf("Validation globale : %s", group.Name),
				Icon:        "sword",
				PositionX:   250,
				PositionY:   currentYPos,
				SortOrder:   (gIdx * 10) + 9,
				QuizConfig:  &synthConfigStr,
				CreatedAt:   now,
			}
			_ = s.repos.CourseNode.Create(ctx, synthNode)

			// Edges from group quizzes to synthèse
			for cIdx, qNodeID := range currentGroupEndNodes {
				e := &model.CourseEdge{
					ID:           fmt.Sprintf("%s_es_%d_%d", courseID, gIdx, cIdx),
					CourseID:     courseID,
					SourceNodeID: qNodeID,
					TargetNodeID: synthNodeID,
					EdgeType:     "required",
					CreatedAt:    now,
				}
				_ = s.repos.CourseEdge.Create(ctx, e)
			}

			// Synthèse Quiz
			synthQuiz := &model.Quiz{
				ID:               actualSynthQuizID,
				Title:            fmt.Sprintf("Synthèse : %s", group.Name),
				Description:      "Mettez vos connaissances à l'épreuve !",
				CourseID:         &courseID,
				NodeID:           &synthNodeID,
				DeckID:           &deckID,
				PassPercentage:   intPtr(80),
				ShuffleQuestions: true,
				QuestionMode:     "mixed",
				GivesCoins:       true,
				CoinReward:       30,
				IsPublic:         false,
				IsSystem:         true,
				IsDynamic:        true,
				CreatedBy:        createdBy,
				CreatedAt:        now,
				IsActive:         true,
			}
			_ = s.repos.Quiz.Create(ctx, synthQuiz)

			// Add Templates for every category in this group
			qPerCat := synthQuestionCount / len(group.Categories)
			if qPerCat < 2 {
				qPerCat = 2
			}
			questTypes, _ := json.Marshal([]string{"mcq", "translate"})
			dirs, _ := json.Marshal([]string{"source_to_target", "target_to_source", "attach_suffix", "conjugate"})

			for _, cat := range group.Categories {
				catID := cat.ID
				dID := deckID
				template := &model.QuestionTemplate{
					ID:             utils.GenerateSnowflakeID(),
					QuizID:         synthQuiz.ID,
					DeckID:         &dID,
					CategoryID:     &catID,
					QuestionTypes:  string(questTypes),
					Directions:     string(dirs),
					GenerationMode: "random_from_deck",
					QuestionCount:  qPerCat,
					CreatedAt:      now,
				}
				_ = s.repos.QuestionTemplate.Create(ctx, template)
			}

			currentGroupEndNodes = []string{synthNodeID}
		}

		prevNodeIDs = currentGroupEndNodes
	}

	return course, nil
}

// ============================================================
// Response types
// ============================================================

type QuizResult struct {
	AttemptID   int64          `json:"attempt_id,string"`
	Score       float64        `json:"score"`
	MaxScore    float64        `json:"max_score"`
	Percentage  float64        `json:"percentage"`
	Passed      bool           `json:"passed"`
	CoinsEarned int            `json:"coins_earned"`
	TimeTaken   int            `json:"time_taken"`
	Results     []AnswerResult `json:"results"`
}

type AnswerResult struct {
	QuestionID    int64   `json:"question_id,string"`
	QuestionText  string  `json:"question_text"`
	UserAnswer    string  `json:"user_answer"`
	CorrectAnswer string  `json:"correct_answer"`
	IsCorrect     bool    `json:"is_correct"`
	PointsEarned  float64 `json:"points_earned"`
	QuestionType  string  `json:"question_type"`
	AIExplanation string  `json:"ai_explanation,omitempty"`
}

type NodeStatus struct {
	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
	State    string `json:"state"` // "locked", "unlocked", "completed"
}

// ============================================================
// PreviewDeckSelection represents the request payload for previewing decks
type PreviewDeckSelection struct {
	DeckID     string   `json:"deck_id"`
	Categories []string `json:"categories"`
}

// PreviewQuestion represents a generated question for preview
type PreviewQuestion struct {
	QuestionText  string   `json:"question_text"`
	CorrectAnswer string   `json:"correct_answer"`
	Options       []string `json:"options,omitempty"`
	QuestionType  string   `json:"question_type"`
	Direction     string   `json:"direction"`
}

// PreviewQuestions generates sample questions from deck entries for review
func (s *QuizService) PreviewQuestions(ctx context.Context, deckSelections []PreviewDeckSelection, questionTypes []string, directions []string, count int) ([]PreviewQuestion, error) {
	if len(deckSelections) == 0 {
		return nil, fmt.Errorf("at least one deck selection is required")
	}

	if len(questionTypes) == 0 {
		questionTypes = []string{"mcq", "translate"}
	}
	if len(directions) == 0 {
		directions = []string{"source_to_target", "target_to_source"}
	}

	// Build parsed decks
	var parsedDecks []*ParsedDeck
	var config QuizConfig

	for _, sel := range deckSelections {
		deckID, err := strconv.ParseInt(sel.DeckID, 10, 64)
		if err != nil {
			continue
		}

		deck, err := s.repos.Deck.GetByID(ctx, deckID)
		if err != nil {
			continue
		}

		// Get deck cache for parsed data
		cache, err := s.repos.DeckCache.GetByDeckID(ctx, deckID)
		if err != nil {
			continue
		}

		var parsed ParsedDeck
		if err := json.Unmarshal([]byte(cache.CachedData), &parsed); err != nil {
			continue
		}
		parsed.DeckID = deckID

		// Parse deck metadata
		var meta DeckMetadata
		json.Unmarshal([]byte(deck.DeckMetadata), &meta)
		parsed.Metadata = meta

		parsedDecks = append(parsedDecks, &parsed)
		config.DeckSelections = append(config.DeckSelections, DeckSelection{
			DeckID:     deckID,
			Categories: sel.Categories,
		})
	}

	if len(parsedDecks) == 0 {
		return nil, fmt.Errorf("no valid decks found")
	}

	// Build question types and directions
	for _, qt := range questionTypes {
		config.QuestionTypes = append(config.QuestionTypes, QuestionType(qt))
	}
	for _, d := range directions {
		config.Directions = append(config.Directions, Direction(d))
	}
	config.QuestionCount = count

	// Generate using the existing generator
	quiz, err := s.generator.GenerateQuiz(ctx, config, parsedDecks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	// Convert to preview format
	var result []PreviewQuestion
	for _, q := range quiz.Questions {
		result = append(result, PreviewQuestion{
			QuestionText:  q.QuestionText,
			CorrectAnswer: q.CorrectAnswer,
			Options:       q.Options,
			QuestionType:  string(q.QuestionType),
			Direction:     string(q.Direction),
		})
	}

	return result, nil
}

func intPtr(i int) *int {
	return &i
}
