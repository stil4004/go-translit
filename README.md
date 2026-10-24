# translit

Двунаправленная транслитерация русский ↔ английский для Go.

## Установка

    go get github.com/yourusername/translit

## Использование

```go
import "github.com/yourusername/translit"

translit.RuToEn("привет")       // "privet"
translit.EnToRu("privet")       // "привет"

// Режимы
translit.RuToEnMode("пq", translit.NonStrict) // "pk"

// Своя карта
m := translit.RuToEnMap()
m['я'] = "a"
translit.Transliterate("юля", m) // "ula"