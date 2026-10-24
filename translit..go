// Package translit реализует транслитерацию между русским и английским
// алфавитами в обоих направлениях.
//
// Базовое использование:
//
//	translit.RuToEn("привет")   // "privet"
//	translit.EnToRu("privet")   // "привет"
//
// Своя карта замен:
//
//	m := translit.RuToEnMap()
//	m['я'] = "a" // переопределяем
//	translit.Transliterate("юля", m) // "ula"
package translit

import (
	"strings"
	"unicode/utf8"
)

// Mode определяет поведение при встрече символов,
// не входящих в исходный алфавит.
type Mode int

const (
	// Strict — транслитерируются только символы исходного алфавита.
	// Прочие символы остаются без изменений.
	// Пример: RuToEn("пkr", Strict) == "pkr"
	Strict Mode = iota

	// NonStrict — если в строке встречены символы целевого алфавита,
	// они сначала нормализуются к исходному алфавиту
	// (исправление смешанного ввода), после чего вся строка
	// транслитерируется целиком.
	// Пример: RuToEn("пq", NonStrict) == "pk"
	NonStrict
)

// ruToEn — таблица транслитерации русских букв в английские.
var ruToEn = map[rune]string{
	// строчные
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d",
	'е': "e", 'ё': "yo", 'ж': "zh", 'з': "z", 'и': "i",
	'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n",
	'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch",
	'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "",
	'э': "e", 'ю': "yu", 'я': "ya",
	// заглавные
	'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D",
	'Е': "E", 'Ё': "Yo", 'Ж': "Zh", 'З': "Z", 'И': "I",
	'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N",
	'О': "O", 'П': "P", 'Р': "R", 'С': "S", 'Т': "T",
	'У': "U", 'Ф': "F", 'Х': "Kh", 'Ц': "Ts", 'Ч': "Ch",
	'Ш': "Sh", 'Щ': "Shch", 'Ъ': "", 'Ы': "Y", 'Ь': "",
	'Э': "E", 'Ю': "Yu", 'Я': "Ya",
}

// enToRuDigraphs — многобуквенные комбинации En→Ru.
// Порядок важен: более длинные должны проверяться раньше.
var enToRuDigraphs = [][2]string{
	{"shch", "щ"}, {"Shch", "Щ"}, {"SHCH", "Щ"},
	{"sch", "щ"}, {"Sch", "Щ"}, {"SCH", "Щ"},
	{"sh", "ш"}, {"Sh", "Ш"}, {"SH", "Ш"},
	{"ch", "ч"}, {"Ch", "Ч"}, {"CH", "Ч"},
	{"zh", "ж"}, {"Zh", "Ж"}, {"ZH", "Ж"},
	{"kh", "х"}, {"Kh", "Х"}, {"KH", "Х"},
	{"ts", "ц"}, {"Ts", "Ц"}, {"TS", "Ц"},
	{"yo", "ё"}, {"Yo", "Ё"}, {"YO", "Ё"},
	{"yu", "ю"}, {"Yu", "Ю"}, {"YU", "Ю"},
	{"ya", "я"}, {"Ya", "Я"}, {"YA", "Я"},
}

// enToRuSingle — одиночные буквы En→Ru.
// Внимание: обратное преобразование принципиально неоднозначно
// (е и э обе дают "e", ы и й дают "y"). Здесь выбраны
// наиболее частотные варианты.
var enToRuSingle = map[rune]string{
	'a': "а", 'b': "б", 'c': "к", 'd': "д", 'e': "е",
	'f': "ф", 'g': "г", 'h': "х", 'i': "и", 'j': "дж",
	'k': "к", 'l': "л", 'm': "м", 'n': "н", 'o': "о",
	'p': "п", 'q': "к", 'r': "р", 's': "с", 't': "т",
	'u': "у", 'v': "в", 'w': "в", 'x': "кс", 'y': "ы",
	'z': "з",
	'A': "А", 'B': "Б", 'C': "К", 'D': "Д", 'E': "Е",
	'F': "Ф", 'G': "Г", 'H': "Х", 'I': "И", 'J': "Дж",
	'K': "К", 'L': "Л", 'M': "М", 'N': "Н", 'O': "О",
	'P': "П", 'Q': "К", 'R': "Р", 'S': "С", 'T': "Т",
	'U': "У", 'V': "В", 'W': "В", 'X': "Кс", 'Y': "Ы",
	'Z': "З",
}

// RuToEn транслитерирует русскую строку в английскую (strict).
func RuToEn(s string) string { return RuToEnMode(s, Strict) }

// EnToRu транслитерирует английскую строку в русскую (strict).
func EnToRu(s string) string { return EnToRuMode(s, Strict) }

// RuToEnMode — как RuToEn, но с явным режимом.
func RuToEnMode(s string, mode Mode) string {
	if mode == NonStrict {
		s = EnToRuMode(s, Strict) // нормализуем «чужие» буквы к русским
	}
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		if tr, ok := ruToEn[r]; ok {
			b.WriteString(tr)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// EnToRuMode — как EnToRu, но с явным режимом.
func EnToRuMode(s string, mode Mode) string {
	if mode == NonStrict {
		s = RuToEnMode(s, Strict) // нормализуем «чужие» буквы к английским
	}
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		// 1. Пробуем многобуквенные комбинации
		matched := false
		for _, d := range enToRuDigraphs {
			if strings.HasPrefix(s[i:], d[0]) {
				b.WriteString(d[1])
				i += len(d[0])
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// 2. Одиночная руна
		r, size := utf8.DecodeRuneInString(s[i:])
		if tr, ok := enToRuSingle[r]; ok {
			b.WriteString(tr)
		} else {
			b.WriteRune(r)
		}
		i += size
	}
	return b.String()
}

// Transliterate применяет пользовательскую карту замен.
// Символы, которых нет в карте, остаются без изменений.
func Transliterate(s string, m map[rune]string) string {
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		if tr, ok := m[r]; ok {
			b.WriteString(tr)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// RuToEnMap возвращает копию стандартной карты Ru→En.
// Удобно для модификации и передачи в Transliterate.
func RuToEnMap() map[rune]string {
	m := make(map[rune]string, len(ruToEn))
	for k, v := range ruToEn {
		m[k] = v
	}
	return m
}