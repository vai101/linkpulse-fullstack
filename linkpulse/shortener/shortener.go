// shortener/shortener.go

package shortener

// The character set for our Base62 encoding.
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62Encode takes a number and returns its Base62 encoded string representation.
func Base62Encode(number uint64) string {
	if number == 0 {
		return "0"
	}

	var result []byte
	base := uint64(len(base62Chars)) // This is 62

	// The algorithm is simple:
	// 1. Take the remainder of the number divided by 62.
	// 2. The remainder is the index of the character in our base62Chars string.
	// 3. Prepend that character to our result.
	// 4. Divide the number by 62.
	// 5. Repeat until the number is 0.
	for number > 0 {
		remainder := number % base
		result = append(result, base62Chars[remainder])
		number /= base
	}

	// The result is currently in reverse order, so we need to reverse it.
	// For example, encoding 1000 would produce "8g" but we want "g8".
	// A simple way to reverse is to convert to string and swap characters.
	return reverse(string(result))
}

// A simple helper function to reverse a string.
func reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}