package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func GenerateOfferPDF(htmlTemplate string, data map[string]interface{}, outputPath string) error {
	tmpl, err := template.New("offer").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	fontPath := findChineseFont()
	if fontPath != "" {
		pdf.AddUTF8Font("zh", "", fontPath)
		pdf.SetFont("zh", "", 12)
	} else {
		pdf.SetFont("Helvetica", "", 12)
	}

	html := body.String()
	renderHTMLToPDF(pdf, html)

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("save pdf: %w", err)
	}

	return nil
}

func findChineseFont() string {
	if runtime.GOOS == "windows" {
		fonts := []string{
			`C:\Windows\Fonts\msyh.ttc`,
			`C:\Windows\Fonts\msyh.ttf`,
			`C:\Windows\Fonts\simhei.ttf`,
			`C:\Windows\Fonts\simsun.ttc`,
		}
		for _, f := range fonts {
			if _, err := os.Stat(f); err == nil {
				return f
			}
		}
	}
	fontCandidates := []string{
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/System/Library/Fonts/PingFang.ttc",
		"/System/Library/Fonts/STHeiti Medium.ttc",
	}
	for _, f := range fontCandidates {
		if _, err := os.Stat(f); err == nil {
			return f
		}
	}
	return ""
}

type pdfRenderer struct {
	pdf        *gofpdf.Fpdf
	bold       bool
	italic     bool
	fontSize   float64
	listLevel  int
	inList     bool
	listType   string
	paragraphs []string
}

func renderHTMLToPDF(pdf *gofpdf.Fpdf, html string) {
	r := &pdfRenderer{
		pdf:      pdf,
		fontSize: 12,
	}

	html = strings.ReplaceAll(html, "\r\n", "\n")
	html = strings.ReplaceAll(html, "\n", " ")
	html = strings.ReplaceAll(html, "\t", " ")
	html = collapseSpaces(html)

	tokens := tokenizeHTML(html)
	for _, token := range tokens {
		r.processToken(token)
	}

	r.flushParagraph()
}

func collapseSpaces(s string) string {
	var result strings.Builder
	prevSpace := false
	for _, ch := range s {
		if ch == ' ' {
			if !prevSpace {
				result.WriteRune(ch)
			}
			prevSpace = true
		} else {
			prevSpace = false
			result.WriteRune(ch)
		}
	}
	return result.String()
}

type htmlToken struct {
	tag     string
	closing bool
	content string
	style   string
}

func tokenizeHTML(html string) []htmlToken {
	var tokens []htmlToken
	i := 0
	for i < len(html) {
		if html[i] == '<' {
			end := strings.IndexByte(html[i:], '>')
			if end == -1 {
				break
			}
			tagContent := html[i+1 : i+end]
			i += end + 1

			closing := false
			tag := tagContent
			style := ""

			if strings.HasPrefix(tagContent, "/") {
				closing = true
				tag = strings.TrimPrefix(tagContent, "/")
			}

			if idx := strings.Index(tag, " "); idx != -1 {
				attrs := tag[idx+1:]
				tag = tag[:idx]
				if si := strings.Index(attrs, "style="); si != -1 {
					style = attrs[si+6:]
					if len(style) > 0 && style[0] == '"' {
						style = style[1:]
					}
					if ei := strings.Index(style, "\""); ei != -1 {
						style = style[:ei]
					}
				}
			}

			tag = strings.ToLower(strings.TrimSpace(tag))

			if tag != "" {
				tokens = append(tokens, htmlToken{tag: tag, closing: closing, style: style})
			}
		} else {
			end := strings.IndexByte(html[i:], '<')
			content := ""
			if end == -1 {
				content = strings.TrimSpace(html[i:])
				i = len(html)
			} else {
				content = strings.TrimSpace(html[i : i+end])
				i += end
			}
			if content != "" {
				tokens = append(tokens, htmlToken{tag: "text", content: content})
			}
		}
	}
	return tokens
}

func (r *pdfRenderer) processToken(t htmlToken) {
	style := ""
	if !t.closing {
		style = r.parseStyle(t.style)
	}

	switch t.tag {
	case "text":
		r.paragraphs = append(r.paragraphs, t.content)

	case "br":
		r.paragraphs = append(r.paragraphs, "\n")

	case "h1", "h2", "h3", "h4", "h5", "h6":
		r.flushParagraph()
		r.applyHeadingStyle(t.tag)
		if !t.closing {
			r.paragraphs = append(r.paragraphs, "")
		} else {
			r.flushParagraph()
			r.resetStyle()
			r.pdf.Ln(4)
		}

	case "p", "div":
		if t.closing {
			r.flushParagraph()
			r.pdf.Ln(3)
		}

	case "strong", "b":
		if !t.closing {
			r.bold = true
			r.applyStyle()
		} else {
			r.bold = false
			r.applyStyle()
		}

	case "em", "i":
		if !t.closing {
			r.italic = true
			r.applyStyle()
		} else {
			r.italic = false
			r.applyStyle()
		}

	case "ul", "ol":
		if !t.closing {
			r.flushParagraph()
			r.inList = true
			r.listType = t.tag
			r.listLevel++
		} else {
			r.flushParagraph()
			r.inList = false
			r.listLevel--
			r.pdf.Ln(2)
		}

	case "li":
		if !t.closing {
			r.flushParagraph()
			prefix := "• "
			if r.listType == "ol" {
				prefix = fmt.Sprintf("%d. ", r.pdf.GetX()/10)
			}
			r.pdf.SetX(r.pdf.GetX() + float64(r.listLevel)*5)
			r.paragraphs = append(r.paragraphs, prefix)
		} else {
			r.flushParagraph()
			r.pdf.Ln(2)
		}

	case "hr":
		r.flushParagraph()
		r.pdf.Ln(2)
		r.pdf.Line(r.pdf.GetX(), r.pdf.GetY(), 190, r.pdf.GetY())
		r.pdf.Ln(4)

	case "span":
		if style != "" && !t.closing {
			r.applyInlineStyle(style)
		}

	case "style", "script":
	}
}

func (r *pdfRenderer) parseStyle(style string) string {
	return style
}

func (r *pdfRenderer) applyHeadingStyle(tag string) {
	sizes := map[string]float64{
		"h1": 22, "h2": 18, "h3": 16, "h4": 14, "h5": 13, "h6": 12,
	}
	if s, ok := sizes[tag]; ok {
		r.fontSize = s
		r.bold = true
		r.applyStyle()
	}
}

func (r *pdfRenderer) applyInlineStyle(style string) {
	parts := strings.Split(style, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "font-size":
			val = strings.ReplaceAll(val, "px", "")
			val = strings.ReplaceAll(val, "pt", "")
			val = strings.TrimSpace(val)
			if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 {
				r.fontSize = f
			}
		case "font-weight":
			if val == "bold" {
				r.bold = true
			}
		case "font-style":
			if val == "italic" {
				r.italic = true
			}
		case "color":
			r.pdf.SetTextColor(0, 0, 0)
		}
	}
	r.applyStyle()
}

func (r *pdfRenderer) resetStyle() {
	r.bold = false
	r.italic = false
	r.fontSize = 12
	r.applyStyle()
}

func (r *pdfRenderer) applyStyle() {
	style := ""
	if r.bold {
		style += "B"
	}
	if r.italic {
		style += "I"
	}
	r.pdf.SetFont("zh", style, r.fontSize)
}

func (r *pdfRenderer) flushParagraph() {
	if len(r.paragraphs) == 0 {
		return
	}

	text := strings.Join(r.paragraphs, "")
	text = strings.TrimSpace(text)
	r.paragraphs = nil

	if text == "" {
		return
	}

	x := r.pdf.GetX()
	if r.listLevel > 0 {
		x = 10 + float64(r.listLevel)*5
	}

	r.pdf.SetX(x)
	r.pdf.MultiCell(190-x+10, r.fontSize/2+2, text, "", "L", false)
}
