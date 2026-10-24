package translit

import "testing"

func TestRuToEn(t *testing.T) {
	cases := map[string]string{
		"привет":     "privet",
		"Привет":     "Privet",
		"ёж":         "yozh",
		"щука":       "shchuka",
		"молоко":     "moloko",
		"съезд":      "sezd",
		"объявление": "obyavlenie",
	}
	for in, want := range cases {
		if got := RuToEn(in); got != want {
			t.Errorf("RuToEn(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnToRu(t *testing.T) {
	cases := map[string]string{
		"privet":  "привет",
		"Privet":  "Привет",
		"yozh":    "ёж",
		"shchuka": "щука",
		"moloko":  "молоко",
	}
	for in, want := range cases {
		if got := EnToRu(in); got != want {
			t.Errorf("EnToRu(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDigraphPriority(t *testing.T) {
	// "shch" должен матчиться раньше, чем "sh"
	if got := EnToRu("shchuka"); got != "щука" {
		t.Errorf("EnToRu(shchuka) = %q, want %q", got, "щука")
	}
}

func TestStrictMixed(t *testing.T) {
	if got := RuToEnMode("пkr", Strict); got != "pkr" {
		t.Errorf("strict: RuToEn(пkr) = %q, want pkr", got)
	}
}

func TestNonStrictMixed(t *testing.T) {
	// q — английская, в non-strict нормализуется к русской «к»
	if got := RuToEnMode("пq", NonStrict); got != "pk" {
		t.Errorf("non-strict: RuToEn(пq) = %q, want pk", got)
	}
}

func TestCustomMap(t *testing.T) {
	m := map[rune]string{
		'ю': "u",
		'я': "a",
		'л': "l",
	}
	if got := Transliterate("юля", m); got != "ula" {
		t.Errorf("Transliterate = %q, want ula", got)
	}
}

func TestRuToEnMapCopy(t *testing.T) {
	m := RuToEnMap()
	m['я'] = "a"
	if got := Transliterate("я", m); got != "a" {
		t.Errorf("custom я = %q, want a", got)
	}
	// Проверяем, что оригинал не изменился
	if got := RuToEn("я"); got != "ya" {
		t.Errorf("original RuToEn(я) = %q, want ya", got)
	}
}