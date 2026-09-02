package gbcarkhos

import "testing"

func TestParseByteSize(t *testing.T) {
	tests := map[string]int64{
		"1024":  1024,
		"10MB":  10_000_000,
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

func TestParseByteSizeRejectsInvalidValues(t *testing.T) {
	for _, input := range []string{"", "0", "-1MiB", "1TiB", "999999999999999999999GiB"} {
		if _, err := parseByteSize(input); err == nil {
			t.Fatalf("parseByteSize(%q) should fail", input)
		}
	}
}
