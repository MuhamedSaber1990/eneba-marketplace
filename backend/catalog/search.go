package catalog

import (
	"strings"
	"unicode"
)

// Search by name, tolerant of typos and partial words.
//
// A search text is split into words, and a product matches when EVERY word
// matches its name. One word matches in either of two ways:
//
//  1. The name contains it. "forge" is inside "Mythforge", so partial words
//     and beginnings of words work.
//  2. It is similar enough to part of the name. PostgreSQL's pg_trgm
//     extension splits text into trigrams (three-letter slices) and
//     word_similarity() says how much of the word's trigrams appear together
//     in the name, from 0 to 1. "circut" scores 0.71 against "Thunder Circuit
//     Kart", so a typo still matches.
//
// Checking word by word is what lets "kart xbox" find "Thunder Circuit Kart
// Xbox Key GLOBAL", where the two words are not next to each other.

const (
	// similarityThreshold is how similar a word must be to count as a typo of
	// something in the name. It was chosen by measuring the seed catalog:
	// real typos scored 0.50 or more ("legnds" 0.50, "obsidain" 0.56,
	// "circut" 0.71) and unrelated words 0.40 or less ("kart" against
	// "Garage" 0.40, "tied" against "Tide" 0.40). 0.45 sits in the gap.
	similarityThreshold = 0.45

	// minFuzzyLength: words shorter than this must be contained exactly. A
	// three-letter word has only four trigrams, so one wrong letter changes
	// most of them and similarity stops meaning anything. It is text and not
	// a number because it is written into the SQL below. It is a constant of
	// this program, never input, so that is safe.
	minFuzzyLength = "4"

	maxSearchWords = 5  // more words than this are ignored
	maxWordLength  = 40 // longer words are cut
)

// SuggestLimit is how many products the search box dropdown shows.
const SuggestLimit = 6

// MinSuggestLength is the fewest letters a search needs before suggestions
// are offered.
const MinSuggestLength = 2

// SearchWords turns what a visitor typed into the words to search for:
// lowercase, split wherever there is something that is not a letter or a
// digit, at most maxSearchWords.
//
// Only letters and digits survive, so characters that mean something in a
// LIKE pattern (% and _) never reach the query, and "%" alone becomes zero
// words.
func SearchWords(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	words := []string{}
	for _, word := range fields {
		if len(words) == maxSearchWords {
			break
		}
		// A word is cut by characters (runes), not bytes, so a letter that
		// takes several bytes is never split in the middle.
		if runes := []rune(word); len(runes) > maxWordLength {
			word = string(runes[:maxWordLength])
		}
		words = append(words, word)
	}
	return words
}

// searchFilter is the line of the listing's WHERE clause that applies the
// search. $3 is the array of words. With no words the array is empty, the
// inner SELECT has no rows, and NOT EXISTS is true: every product passes.
//
// Read it as: "there is no word that fails to match". $6 is the threshold.
const searchFilter = `
   AND NOT EXISTS (
         SELECT 1
           FROM unnest($3::text[]) AS word
          WHERE NOT (
                name ILIKE '%' || word || '%'
             OR (length(word) >= ` + minFuzzyLength + ` AND word_similarity(word, name) >= $6::real)
          ))`

// relevanceOrder ranks search results, best first. $9 is the whole search
// text (the words joined by spaces).
//
//	group 0: the name starts with the whole text
//	group 1: the name contains the whole text
//	group 2: every word is contained in the name, in any order
//	group 3: at least one word matched only by similarity (a typo)
//
// Inside a group: products that can be bought come before sold-out ones (a
// sold-out product has no price, so from_price_cents is NULL); then the
// closer the whole text is to the name, the higher; then the product that
// sold more; then slug, so the order is always the same.
const relevanceOrder = `
       CASE
         WHEN name ILIKE $9::text || '%' THEN 0
         WHEN name ILIKE '%' || $9::text || '%' THEN 1
         WHEN NOT EXISTS (SELECT 1 FROM unnest($3::text[]) AS word
                           WHERE name NOT ILIKE '%' || word || '%') THEN 2
         ELSE 3
       END,
       (from_price_cents IS NULL),
       word_similarity($9::text, name) DESC,
       COALESCE((SELECT units_sold FROM product_sales sales
                  WHERE sales.product_id = priced.id), 0) DESC,
       slug`
