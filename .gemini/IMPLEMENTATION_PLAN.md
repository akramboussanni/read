# Major Recode: Progression & Quizzing System

## Overview
A complete overhaul of the quiz and progression system with:
1. **LLM-powered question generation** (Gemini API)
2. **Runtime random question generation** with configurable templates
3. **Visual progression tree** (Kahoot-like, kid-friendly UI)
4. **Auto progression route creation** from decks (smart category grouping)
5. **Unified admin/student view** (same component, different behavior)
6. **Complete frontend redesign** - Kahoot-like, colorful, accessible

## Architecture Changes

### Phase 1: Database Schema Reset
Since we're wiping the DB, create a single clean migration.

**New/Modified Tables:**
- `quiz_decks` - unchanged
- `quiz_categories` - unchanged  
- `deck_entries` - unchanged
- `quizzes` - simplified, remove legacy fields
- `quiz_questions` - add `generation_config` JSONB for template-based generation
- `progression_paths` - enhanced with `deck_id` for auto-generation
- `progression_nodes` - enhanced with `node_type` enum, `lesson_content` JSONB
- `progression_edges` - unchanged
- `user_path_enrollments` - add `completed_nodes` JSONB
- NEW `question_templates` - defines HOW to generate questions at runtime
- Removed: `questions`, `question_options` (legacy, replaced by deck_entries + templates)
- Removed: `user_progression` (replaced by path-based progression)

### Phase 2: Backend - Question Template System
Instead of storing pre-generated questions, store **templates** that describe how to generate questions at runtime.

```go
// QuestionTemplate defines how to generate a question
type QuestionTemplate struct {
    ID             int64  `json:"id,string" db:"id"`
    DeckID         int64  `json:"deck_id,string" db:"deck_id"`
    CategoryID     *int64 `json:"category_id,string,omitempty" db:"category_id"`
    QuestionTypes  string `json:"question_types" db:"question_types"` // JSON: ["mcq","translate"]
    Directions     string `json:"directions" db:"directions"`         // JSON: ["source_to_target","target_to_source"]
    GenerationMode string `json:"generation_mode" db:"generation_mode"` // "random_from_deck", "llm", "manual"
    LLMPrompt      string `json:"llm_prompt,omitempty" db:"llm_prompt"` // For LLM mode
    ManualData     string `json:"manual_data,omitempty" db:"manual_data"` // For manual questions
    Count          int    `json:"count" db:"count"`
    CreatedAt      int64  `json:"created_at,string" db:"created_at"`
}
```

### Phase 3: Backend - LLM Question Generation
- Extend `AIGrader` into a more general `LLMService`
- Add question generation endpoint using Gemini
- Support batch generation with caching
- Question types: conjugation, fill-blank, contextual translation, grammar correction

### Phase 4: Backend - Smart Path Auto-Generation
- Input: A deck ID
- Process: 
  1. Load all categories from the deck
  2. Smart ordering: detect numbering patterns ("Pronouns I", "Pronouns II" → ordered)
  3. Group by difficulty if available
  4. Generate a tree: Start → Category groups → Quiz nodes → End
  5. Each quiz node gets question templates from its category
- Output: A complete ProgressionPath with nodes, edges, and quiz templates

### Phase 5: Frontend - Complete Redesign

#### Design System: Kahoot-inspired
- **Colors**: Vibrant purple (#6C5CE7), coral (#FF6B6B), teal (#00CEC9), gold (#FDCB6E)
- **Font**: Google Fonts "Nunito" (rounded, kid-friendly)
- **Cards**: Large rounded corners (20px+), subtle shadows, hover animations
- **Buttons**: Pill-shaped, with micro-animations
- **Icons**: Playful Lucide icons
- **Animations**: Framer Motion throughout

#### Pages:
1. **Home** (`/`) - Dashboard with enrolled paths, stats, quick-start
2. **Paths** (`/paths`) - Browse available learning paths
3. **Path View** (`/paths/[id]`) - Visual tree progression (student + admin dual mode)
4. **Quiz Play** (`/quizzes/[id]`) - Kahoot-style quiz experience
5. **Admin** (`/admin`) - Dashboard with path/quiz management
6. **Admin Path Editor** (`/admin/paths/[id]`) - ReactFlow editor with drag-drop
7. **Admin Path Auto-Create** (`/admin/paths/auto`) - Select deck → auto-generate path

## Implementation Order

### Step 1: New Database Migration
- Drop all old tables, create new schema
- Single migration file

### Step 2: Backend Models & Repos  
- New Go models for all tables
- Repository layer with CRUD operations

### Step 3: Backend Services
- `QuestionGenerationService` - runtime question generation
- `LLMService` - Gemini-powered question generation
- `PathAutoGeneratorService` - smart path creation from decks
- Refactored `QuizService`

### Step 4: Backend API Routes
- Path CRUD + auto-generation
- Quiz execution with template-based generation
- LLM question generation endpoint

### Step 5: Frontend Types & API Client
- New TypeScript types
- API client functions

### Step 6: Frontend Design System
- New globals.css with Kahoot-inspired theme
- Updated UI components

### Step 7: Frontend Pages
- Progression tree view (ReactFlow)
- Quiz play page (Kahoot-style)
- Admin editor
- Home dashboard

### Step 8: Polish & Testing
- Animations
- Error handling
- Loading states
- Responsive design
