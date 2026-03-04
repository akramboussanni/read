import json
import os

filepath = r'c:\Users\abous\dev\read\internal\quiz\data\templates\course_template_quran_vocab.json'

with open(filepath, 'r', encoding='utf-8') as f:
    data = json.load(f)

groups = data['groups']

# Find the verb groups
verb_groups = []
other_groups = []

for g in groups:
    if g['name'].startswith('Verbes') or 'Trilitères' in g['name']:
        verb_groups.append(g)
    else:
        other_groups.append(g)

# The user wants to:
# 1. give one category from the last group to the one before it
last_verb_group = verb_groups[-1]
prev_verb_group = verb_groups[-2]

if 'augmented_verbs_form_4_c' in last_verb_group['categories']:
    last_verb_group['categories'].remove('augmented_verbs_form_4_c')
    prev_verb_group['categories'].append('augmented_verbs_form_4_c')

# 2. Put stuff in logical order (interleave verbs with grammar)
# Current other_groups: 10 items
# Current verb_groups: 7 items

# Proposed Interleaved Order:
# 1. Pronoms (G1)
# 2. Particules (G2)
# 3. Verbes Trilitères A-C (VG1)
# 4. Prépositions (G3)
# 5. Verbes Trilitères D-E (VG2)
# 6. Préfixes & Particules (G4)
# 7. Verbes Irréguliers I (VG3)
# 8. Noms d'Allah I (G5)
# 9. Verbes Irréguliers II (VG4)
# 10. Noms d'Allah II (G6)
# 11. Verbes Formes 2 & 3 (VG5)
# 12. Création & Vie (G7)
# 13. Verbes Formes 4 A-C (VG6)
# 14. Eschatologie (G8)
# 15. Verbes Formes 4 D & 5-10 (VG7)
# 16. Foi & Morale (G9)
# 17. Prophètes & Société (G10)

new_groups = []
new_groups.append(other_groups[0]) # Pronoms
new_groups.append(other_groups[1]) # Particules
new_groups.append(verb_groups[0])  # Trilitères A-C
new_groups.append(other_groups[2]) # Prépositions
new_groups.append(verb_groups[1])  # Trilitères D-E
new_groups.append(other_groups[3]) # Préfixes
new_groups.append(verb_groups[2])  # Irréguliers I
new_groups.append(other_groups[4]) # Noms d'Allah I
new_groups.append(verb_groups[3])  # Irréguliers II
new_groups.append(other_groups[5]) # Noms d'Allah II
new_groups.append(verb_groups[4])  # Formes 2 & 3
new_groups.append(other_groups[6]) # Création
new_groups.append(verb_groups[5])  # Formes 4 A-C
new_groups.append(other_groups[7]) # Eschatologie
new_groups.append(verb_groups[6])  # Formes 4 D & 5-10
new_groups.append(other_groups[8]) # Foi
new_groups.append(other_groups[9]) # Prophètes

data['groups'] = new_groups

with open(filepath, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=4)

print("Restructuring complete.")
