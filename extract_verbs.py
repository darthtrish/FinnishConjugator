# extract_verbs_robust.py

verbs = []

with open("nykysuomensanalista2024.txt", encoding="utf-8") as f:
    for line in f:
        parts = line.strip().split()  # split on any whitespace
        if len(parts) >= 2 and parts[1].strip() == "verbi":
            verbs.append(parts[0].strip())

# Save verbs
with open("verbs_only.txt", "w", encoding="utf-8") as f:
    for verb in verbs:
        f.write(verb + "\n")

print(f"Extracted {len(verbs)} verbs and saved to verbs_only.txt")
