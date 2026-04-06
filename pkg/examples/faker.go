package examples

import (
	"math/rand"
	"time"
)

// Faker provides fake data generation for examples
type Faker struct {
	rng *rand.Rand
}

// NewFaker creates a new faker instance
func NewFaker() *Faker {
	return &Faker{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// String generates a random string
func (f *Faker) String() string {
	words := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit"}
	return words[f.rng.Intn(len(words))]
}

// Name generates a random name
func (f *Faker) Name() string {
	firstNames := []string{"John", "Jane", "Alice", "Bob", "Charlie", "Diana", "Edward", "Fiona"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
	return firstNames[f.rng.Intn(len(firstNames))] + " " + lastNames[f.rng.Intn(len(lastNames))]
}

// Email generates a random email
func (f *Faker) Email() string {
	domains := []string{"example.com", "test.com", "demo.org", "sample.net"}
	names := []string{"user", "admin", "contact", "info", "support", "hello"}
	return names[f.rng.Intn(len(names))] + "@" + domains[f.rng.Intn(len(domains))]
}

// Phone generates a random phone number
func (f *Faker) Phone() string {
	return "+1-555-" + f.digits(3) + "-" + f.digits(4)
}

// URL generates a random URL
func (f *Faker) URL() string {
	domains := []string{"example.com", "test.com", "demo.org", "sample.net"}
	paths := []string{"", "/api", "/users", "/products", "/docs"}
	return "https://" + domains[f.rng.Intn(len(domains))] + paths[f.rng.Intn(len(paths))]
}

// UUID generates a random UUID
func (f *Faker) UUID() string {
	return f.hex(8) + "-" + f.hex(4) + "-" + f.hex(4) + "-" + f.hex(4) + "-" + f.hex(12)
}

// Int generates a random integer
func (f *Faker) Int(min, max int) int {
	if min >= max {
		return min
	}
	return min + f.rng.Intn(max-min)
}

// Float generates a random float
func (f *Faker) Float(min, max float64) float64 {
	return min + f.rng.Float64()*(max-min)
}

// Bool generates a random boolean
func (f *Faker) Bool() bool {
	return f.rng.Intn(2) == 1
}

// Date generates a random date string
func (f *Faker) Date() string {
	year := f.Int(2020, 2025)
	month := f.Int(1, 12)
	day := f.Int(1, 28)
	return f.formatDate(year, month, day)
}

// DateTime generates a random datetime string
func (f *Faker) DateTime() string {
	return f.Date() + "T" + f.formatTime(f.Int(0, 23), f.Int(0, 59), f.Int(0, 59)) + "Z"
}

// IPv4 generates a random IPv4 address
func (f *Faker) IPv4() string {
	return f.formatIP(f.Int(1, 255), f.Int(0, 255), f.Int(0, 255), f.Int(1, 254))
}

// Sentence generates a random sentence
func (f *Faker) Sentence() string {
	words := []string{"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
	count := f.Int(5, 10)
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = words[f.rng.Intn(len(words))]
	}
	return result[0] + " " + joinWords(result[1:]) + "."
}

// Paragraph generates a random paragraph
func (f *Faker) Paragraph() string {
	count := f.Int(3, 6)
	sentences := make([]string, count)
	for i := 0; i < count; i++ {
		sentences[i] = f.Sentence()
	}
	return joinWords(sentences)
}

// Helper functions
func (f *Faker) digits(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte('0' + f.rng.Intn(10))
	}
	return string(result)
}

func (f *Faker) hex(n int) string {
	chars := "0123456789abcdef"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[f.rng.Intn(len(chars))]
	}
	return string(result)
}

func (f *Faker) formatDate(year, month, day int) string {
	return padZero(year, 4) + "-" + padZero(month, 2) + "-" + padZero(day, 2)
}

func (f *Faker) formatTime(hour, minute, second int) string {
	return padZero(hour, 2) + ":" + padZero(minute, 2) + ":" + padZero(second, 2)
}

func (f *Faker) formatIP(a, b, c, d int) string {
	return intToStr(a) + "." + intToStr(b) + "." + intToStr(c) + "." + intToStr(d)
}

func padZero(n, width int) string {
	s := intToStr(n)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func joinWords(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := words[0]
	for i := 1; i < len(words); i++ {
		result += " " + words[i]
	}
	return result
}
