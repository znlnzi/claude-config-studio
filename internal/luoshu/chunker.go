package luoshu

import "strings"

// Chunk represents a text chunk
type Chunk struct {
	Index   int
	Text    string
	Context string // Parent heading
}

// ChunkText splits text into chunks by Markdown sections
// maxChars defaults to 2000
func ChunkText(text string, maxChars int) []Chunk {
	if maxChars <= 0 {
		maxChars = 2000
	}

	sections := splitBySections(text)

	var chunks []Chunk
	idx := 0
	for _, sec := range sections {
		content := strings.TrimSpace(sec.content)
		if len(content) < 50 {
			continue
		}
		if len(content) <= maxChars {
			chunks = append(chunks, Chunk{Index: idx, Text: content, Context: sec.heading})
			idx++
		} else {
			subChunks := splitByParagraphs(content, maxChars)
			for _, sc := range subChunks {
				chunks = append(chunks, Chunk{Index: idx, Text: sc, Context: sec.heading})
				idx++
			}
		}
	}
	return chunks
}

type section struct {
	heading string
	content string
}

func splitBySections(text string) []section {
	lines := strings.Split(text, "\n")
	var sections []section
	currentHeading := ""
	var buf strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if buf.Len() > 0 {
				sections = append(sections, section{heading: currentHeading, content: buf.String()})
				buf.Reset()
			}
			currentHeading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if buf.Len() > 0 {
		sections = append(sections, section{heading: currentHeading, content: buf.String()})
	}
	if len(sections) == 0 {
		sections = append(sections, section{content: text})
	}
	return sections
}

func splitByParagraphs(text string, maxChars int) []string {
	paragraphs := strings.Split(text, "\n\n")
	var result []string
	var current strings.Builder
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if current.Len()+len(p)+2 > maxChars && current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(p)
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
