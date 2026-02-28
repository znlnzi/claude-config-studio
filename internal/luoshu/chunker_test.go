package luoshu

import "testing"

func TestChunkText_SingleShortSection(t *testing.T) {
	text := "## Title\n\nThis is a sufficiently long text content used for testing whether the chunking logic works correctly, it needs to exceed fifty characters to be retained."
	chunks := ChunkText(text, 2000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Context != "Title" {
		t.Errorf("expected context 'Title', got %q", chunks[0].Context)
	}
}

func TestChunkText_MultipleSections(t *testing.T) {
	text := "## Section One\n\nThis is the content of section one, it needs sufficient length to pass the fifty-character minimum threshold filter.\n\n## Section Two\n\nThis is the content of section two, it also needs sufficient length to pass the fifty-character minimum threshold filter."
	chunks := ChunkText(text, 2000)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Context != "Section One" {
		t.Errorf("chunk 0 context: want 'Section One', got %q", chunks[0].Context)
	}
	if chunks[1].Context != "Section Two" {
		t.Errorf("chunk 1 context: want 'Section Two', got %q", chunks[1].Context)
	}
}

func TestChunkText_ShortContentFiltered(t *testing.T) {
	text := "## Title\n\nToo short"
	chunks := ChunkText(text, 2000)
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks for short content, got %d", len(chunks))
	}
}

func TestChunkText_LargeSection_SplitByParagraphs(t *testing.T) {
	text := "## Large Section\n\n"
	for i := 0; i < 10; i++ {
		text += "This is a sufficiently long paragraph content used for testing whether the paragraph splitting logic works correctly when exceeding the maximum character limit.\n\n"
	}
	chunks := ChunkText(text, 200)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for large section, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.Context != "Large Section" {
			t.Errorf("all chunks should have context 'Large Section', got %q", c.Context)
		}
	}
}

func TestChunkText_DefaultMaxChars(t *testing.T) {
	text := "## Default\n\n" + string(make([]byte, 100)) + "This is a sufficiently long content for testing default parameters."
	chunks := ChunkText(text, 0) // 0 → default 2000
	if len(chunks) == 0 {
		// Content might be too short and get filtered, but should not panic
		return
	}
}

func TestChunkText_NoHeadings(t *testing.T) {
	text := "Plain text content without headings that needs to exceed fifty characters in length to be retained by the chunker and not get filtered out."
	chunks := ChunkText(text, 2000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for plain text, got %d", len(chunks))
	}
	if chunks[0].Context != "" {
		t.Errorf("expected empty context for no-heading text, got %q", chunks[0].Context)
	}
}

func TestSplitBySections_Empty(t *testing.T) {
	sections := splitBySections("")
	if len(sections) != 1 {
		t.Fatalf("expected 1 section for empty text, got %d", len(sections))
	}
}

func TestSplitByParagraphs_SingleParagraph(t *testing.T) {
	result := splitByParagraphs("single paragraph", 100)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}
