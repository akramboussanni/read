# Quiz Data

This folder contains embedded JSON files that are compiled into the Go binary.

## Structure

Each JSON file should follow this format:

```json
{
  "deck": {
    "id": "unique_deck_id",
    "title": "Deck Title",
    "version": 1
  },
  "cats": {
    "category_key": "Category Display Name",
    "another_category": "Another Category"
  },
  "items": [
    {
      "id": "unique_item_id",
      "cat": "category_key",
      "ar": "Arabic text",
      "fr": "French text"
    }
  ]
}
```

## Files

- `words.json` - Main vocabulary deck (80% Quranic Words Part 2)

## Adding New Decks

1. Create a new JSON file in this directory following the format above
2. Add the filename to the `deckFiles` array in `internal/quiz/loader.go`
3. Rebuild the application - data will be automatically embedded and seeded on first run

## Notes

- Files are embedded at compile time using Go's `//go:embed` directive
- Data is automatically imported to database on first application startup
- Changes to JSON files require recompiling the application
