import json
import re
from pathlib import Path

# ===== CONFIG =====
INPUT_FILE = "old.json"     # your current JSON
OUTPUT_FILE = "deck.json"   # optimized output

DECK_ID = "quran_80_p2"
DECK_TITLE = "80% des Mots du Qour'ân - Part 2"
VERSION = 1


# ===== HELPERS =====
def slug(text: str) -> str:
    text = text.lower()
    text = re.sub(r"[^a-z0-9]+", "_", text)
    return text.strip("_")


def pad(n: int) -> str:
    return str(n).zfill(3)


# ===== LOAD =====
with open(INPUT_FILE, "r", encoding="utf-8") as f:
    raw = json.load(f)


# ===== BUILD CATEGORIES =====
cats = {}
for section in raw["sections"]:
    cat_id = slug(section["category"])
    cats.setdefault(cat_id, section["category"])


# ===== BUILD ITEMS =====
counters = {}
items = []

for section in raw["sections"]:
    cat_id = slug(section["category"])
    counters.setdefault(cat_id, 0)

    for word in section["words"]:
        counters[cat_id] += 1

        items.append({
            "id": f"{cat_id}_{pad(counters[cat_id])}",
            "cat": cat_id,
            "fr": word["french"].strip(),
            "ar": word["arabic"].strip()
        })


# ===== FINAL OUTPUT =====
output = {
    "deck": {
        "id": DECK_ID,
        "title": DECK_TITLE,
        "version": VERSION
    },
    "cats": cats,
    "items": items
}


# ===== SAVE =====
Path(OUTPUT_FILE).write_text(
    json.dumps(output, ensure_ascii=False, indent=2),
    encoding="utf-8"
)

print(f"✅ Converted {len(items)} items → {OUTPUT_FILE}")
