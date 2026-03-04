import json
import os

PARTS_DIR = os.path.join(os.path.dirname(__file__), "parts")

def update_parts():
    part_files = [f for f in os.listdir(PARTS_DIR) if f.startswith("groups_") and f.endswith(".json")]
    
    for fname in part_files:
        path = os.path.join(PARTS_DIR, fname)
        with open(path, "r", encoding="utf-8") as f:
            groups = json.load(f)
        
        updated = False
        for group in groups:
            # Default rewards
            quiz_reward = 20
            synth_reward = 50
            
            # If it's a "Verbes" group, or something harder, maybe more?
            # For now, let's just apply these defaults if they don't exist
            if "quiz_reward" not in group:
                group["quiz_reward"] = quiz_reward
                updated = True
            if "synth_reward" not in group:
                group["synth_reward"] = synth_reward
                updated = True
                
        if updated:
            with open(path, "w", encoding="utf-8") as f:
                json.dump(groups, f, ensure_ascii=False, indent=2)
            print(f"Updated {fname}")
        else:
            print(f"No changes for {fname}")

if __name__ == "__main__":
    update_parts()
    # Also run the combine script to update the main template
    os.system("python " + os.path.join(os.path.dirname(__file__), "combine_template.py"))
