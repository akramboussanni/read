#!/usr/bin/env python3
"""
Combine group part files into the final course template JSON.
Run from any directory:
  python combine_template.py
"""
import json
import os

PARTS_DIR = os.path.join(os.path.dirname(__file__), "parts")
OUTPUT = os.path.join(os.path.dirname(__file__), "course_template_quran_vocab.json")

HEADER = {
    "template_version": 2,
    "course": {
        "title": "80% des Mots du Qour'ân",
        "description": "Maîtrisez les 80% des mots les plus fréquents du Coran avec des leçons interactives et des quiz progressifs.",
        "icon": "book-open",
        "color": "#6C5CE7",
        "is_default": True,
        "deck_key": "80%_des_mots_du_qour'ân_-_part_2"
    }
}

def main():
    part_files = sorted(f for f in os.listdir(PARTS_DIR) if f.startswith("groups_") and f.endswith(".json"))
    all_groups = []
    for fname in part_files:
        path = os.path.join(PARTS_DIR, fname)
        with open(path, "r", encoding="utf-8") as f:
            groups = json.load(f)
        all_groups.extend(groups)
        print(f"  ✓ {fname} — {len(groups)} group(s)")

    template = {**HEADER, "groups": all_groups}
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(template, f, ensure_ascii=False, indent=2)

    print(f"\n✅ Combined {len(all_groups)} groups → {OUTPUT}")

if __name__ == "__main__":
    main()
