import json

# Pronouns for conjugation
pronouns = ["minä", "sinä", "hän/se", "me", "te", "Te", "he", "passiivi"]

# Common irregular verbs
irregular_verbs = {
    "olla": {"pres": "ol", "past": "ol", "perf": "ollut"},
    "mennä": {"pres": "men", "past": "meni", "perf": "mennyt"},
    "tehdä": {"pres": "tee", "past": "teki", "perf": "tehnyt"},
    "nähdä": {"pres": "näe", "past": "näki", "perf": "nähnyt"},
    "tulla": {"pres": "tul", "past": "tuli", "perf": "tullut"},
    "syödä": {"pres": "syö", "past": "söi", "perf": "syönyt"},
    "ottaa": {"pres": "ota", "past": "otti", "perf": "ottanut"},
    "juosta": {"pres": "juokse", "past": "juoksi", "perf": "juossut"},
}

# Consonant gradation
grad_map = {
    "kk": "k",
    "pp": "p",
    "tt": "t",
    "k": "",
    "p": "v",
    "t": "d",
}


def gradate(stem):
    for strong, weak in grad_map.items():
        if stem.endswith(strong):
            return stem[: -len(strong)] + weak
    return stem


def verb_type(verb):
    if verb.endswith(("ata", "ätä", "ota", "ötä")):
        return 2
    if verb.endswith(("ita", "itä")):
        return 3
    if verb.endswith(("lla", "llä")):
        return 4
    if verb.endswith(("nna", "nnä")):
        return 5
    if verb.endswith(("ra", "rä", "sta", "stä")):
        return 6
    return 1


def base_stems(verb):
    if verb in irregular_verbs:
        return irregular_verbs[verb]
    typ = verb_type(verb)
    if typ == 1:
        stem = verb[:-2]
        return {"pres": stem, "past": gradate(stem) + "i", "perf": stem + "nut"}
    if typ in [2, 3, 5]:
        stem = verb[:-2] + "e"
        return {"pres": stem, "past": gradate(stem) + "i", "perf": stem + "nyt"}
    if typ == 4:
        stem = verb[:-2] + "a"
        return {"pres": stem, "past": gradate(stem) + "i", "perf": stem + "nut"}
    if typ == 6:
        stem = verb[:-2]
        return {"pres": stem, "past": gradate(stem) + "i", "perf": stem + "nyt"}
    return {"pres": verb[:-2], "past": verb[:-2], "perf": verb[:-2] + "nut"}


def fill(lst):
    return lst + [""] * (len(pronouns) - len(lst))


def conjugate(verb):
    stems = base_stems(verb)
    pres, past, perf = stems["pres"], stems["past"], stems["perf"]

    return {
        "preesens": fill(
            [
                pres + "n",
                pres + "t",
                pres,
                pres + "mme",
                pres + "tte",
                pres + "vat",
                pres + "vat",
                pres + "aan",
            ]
        ),
        "preesens_neg": fill(
            [
                "en " + pres,
                "et " + pres,
                "ei " + pres,
                "emme " + pres,
                "ette " + pres,
                "eivät " + pres,
                "eivät " + pres,
                "ei " + pres,
            ]
        ),
        "imperfekti": fill(
            [
                past + "n",
                past + "t",
                past,
                past + "mme",
                past + "tte",
                past + "vat",
                past + "vat",
                past + "tiin",
            ]
        ),
        "imperfekti_neg": fill(
            [
                "en " + past,
                "et " + past,
                "ei " + past,
                "emme " + past,
                "ette " + past,
                "eivät " + past,
                "eivät " + past,
                "ei " + past,
            ]
        ),
        "perfekti": fill(
            [
                "on " + perf,
                "on " + perf,
                "on " + perf,
                "olemme " + perf,
                "olette " + perf,
                "ovat " + perf,
                "ovat " + perf,
                "on " + perf,
            ]
        ),
        "perfekti_neg": fill(["ei ole " + perf] * 8),
        "pluskvamperfekti": fill(["oli " + perf] * 8),
        "pluskvamperfekti_neg": fill(["ei ollut " + perf] * 8),
        "konditionaali": fill([pres + "isin"] * 8),
        "konditionaali_neg": fill(["ei " + pres + "isin"] * 8),
        "potentiaali": fill([pres + "nee"] * 8),
        "potentiaali_neg": fill(["ei " + pres + "nee"] * 8),
        "imperatiivi": fill([pres] * 8),
        "imperatiivi_neg": fill(["älä " + pres] * 8),
    }


# Load infinitive verbs
with open("verbs_only.txt", "r", encoding="utf-8") as f:
    verbs = [v.strip() for v in f if v.strip()]

all_verbs = {v: conjugate(v) for v in verbs}

# Save JSON
with open("verbs_full.json", "w", encoding="utf-8") as f:
    json.dump(all_verbs, f, ensure_ascii=False, indent=2)

print(f"Generated conjugations for {len(verbs)} verbs in verbs_full.json")
