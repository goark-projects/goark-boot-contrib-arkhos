package gbcarkhos

import "testing"

func TestParseByteSize(t *testing.T) {
	tests := map[string]int64{
		"1024":  1024,
		"10MB":  10 << 20,
		"10MiB": 10 << 20,
		"2 GiB": 2 << 30,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := parseByteSize(input)
			if err != nil || got != want {
				t.Fatalf("parseByteSize(%q) = %d/%v, want %d/nil", input, got, err, want)
			}
		})
	}
}

func TestParseByteSize_whenShortUnitsUsed_shouldParseCaseInsensitively(t *testing.T) {
	tests := map[string]int64{
		"0":    0,
		"-1":   -1,
		"1k":   1 << 10,
		"1KB":  1 << 10,
		"1KiB": 1 << 10,
		"1m":   1 << 20,
		"1MIB": 1 << 20,
		"1g":   1 << 30,
		"1GiB": 1 << 30,
		"1t":   1 << 40,
		"1TB":  1 << 40,
		"1p":   1 << 50,
		"1PiB": 1 << 50,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := parseByteSize(input)
			if err != nil {
				t.Fatalf("parseByteSize(%q) error = %v", input, err)
			}
			if got != want {
				t.Fatalf("parseByteSize(%q) = %d, want %d", input, got, want)
			}
		})
	}
}

func TestParseByteSizeRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", "-1MiB", "-2", "999999999999999999999GiB"} {
		if _, err := parseByteSize(input); err == nil {
			t.Fatalf("parseByteSize(%q) should fail", input)
		}
	}
}
