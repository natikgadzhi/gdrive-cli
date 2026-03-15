package output

import "testing"

func TestParseFormat_JSON(t *testing.T) {
	f, err := ParseFormat("json")
	if err != nil {
		t.Fatalf("ParseFormat(\"json\") error: %v", err)
	}
	if f != FormatJSON {
		t.Errorf("ParseFormat(\"json\") = %v, want FormatJSON", f)
	}
}

func TestParseFormat_Markdown(t *testing.T) {
	f, err := ParseFormat("markdown")
	if err != nil {
		t.Fatalf("ParseFormat(\"markdown\") error: %v", err)
	}
	if f != FormatMarkdown {
		t.Errorf("ParseFormat(\"markdown\") = %v, want FormatMarkdown", f)
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := ParseFormat("xml")
	if err == nil {
		t.Fatal("ParseFormat(\"xml\") expected error, got nil")
	}
}

func TestParseFormat_Empty(t *testing.T) {
	_, err := ParseFormat("")
	if err == nil {
		t.Fatal("ParseFormat(\"\") expected error, got nil")
	}
}

func TestFormatString(t *testing.T) {
	if FormatJSON.String() != "json" {
		t.Errorf("FormatJSON.String() = %q, want \"json\"", FormatJSON.String())
	}
	if FormatMarkdown.String() != "markdown" {
		t.Errorf("FormatMarkdown.String() = %q, want \"markdown\"", FormatMarkdown.String())
	}
}
