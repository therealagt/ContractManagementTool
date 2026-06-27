package pades

import "testing"

func TestValidateEmptyPDF(t *testing.T) {
	res, err := Validate(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid for empty PDF")
	}
}

func TestValidateUnsignedPDF(t *testing.T) {
	// Minimal PDF without signature
	pdf := []byte("%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n")
	res, err := Validate(pdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Valid {
		t.Fatal("expected invalid for unsigned PDF")
	}
	if res.SHA256 == "" {
		t.Fatal("expected sha256 hash")
	}
}
