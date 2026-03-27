package moodle

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// QuestionXML is a single Moodle XML question (subset).
type QuestionXML struct {
	ID   int64  `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name>text"`

	QuestionText    MoodleText `xml:"questiontext"`
	GeneralFeedback MoodleText `xml:"generalfeedback"`
	DefaultGrade    float64    `xml:"defaultgrade"`
	Penalty         float64    `xml:"penalty"`
	IDNumber        string     `xml:"idnumber"`

	// RawFragment is the full `<question>...</question>` XML for lossless storage.
	RawFragment string `xml:"-"`
}

// MoodleText is questiontext/generalfeedback with format + text.
type MoodleText struct {
	Format string `xml:"format,attr"`
	Text   string `xml:"text"`
}

// ParseQuestionsXML parses questions.xml and returns questions keyed by legacy question id
// (`id` attribute on `<question>`).
func ParseQuestionsXML(r io.Reader) (map[int64]QuestionXML, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]QuestionXML)
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "question" {
			continue
		}
		var q QuestionXML
		if err := dec.DecodeElement(&q, &se); err != nil {
			return nil, fmt.Errorf("decode question: %w", err)
		}
		if q.ID != 0 {
			pat := `(?s)<question[^>]*\bid\s*=\s*"` + regexp.QuoteMeta(strconv.FormatInt(q.ID, 10)) + `"[^>]*>.*?</question>`
			if re, err := regexp.Compile(pat); err == nil {
				if m := re.Find(data); m != nil {
					q.RawFragment = string(m)
				}
			}
			enrichLegacyQuestionXML(&q)
			out[q.ID] = q
		}
	}
	return out, nil
}

// ParseQuestionsFileCombined tries legacy id-keyed questions, then scans for question_bank_entry
// blocks keyed by entry id (Moodle 4+).
func ParseQuestionsFileCombined(r io.Reader) (byQuestionID map[int64]QuestionXML, byBankEntryID map[int64]QuestionXML, err error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	byQuestionID, err = ParseQuestionsXML(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	byBankEntryID = parseQuestionBankEntries(data)
	return byQuestionID, byBankEntryID, nil
}

func parseQuestionBankEntries(data []byte) map[int64]QuestionXML {
	out := make(map[int64]QuestionXML)
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return out
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "question_bank_entry" {
			continue
		}
		var entry struct {
			ID    int64 `xml:"id,attr"`
			Inner string `xml:",innerxml"`
		}
		if err := dec.DecodeElement(&entry, &se); err != nil {
			continue
		}
		// Find first <question ...> inside innerxml
		idx := strings.Index(entry.Inner, "<question")
		if idx < 0 {
			continue
		}
		sub := entry.Inner[idx:]
		end := strings.Index(sub, "</question>")
		if end < 0 {
			continue
		}
		frag := sub[:end+len("</question>")]
		var q QuestionXML
		if err := xml.Unmarshal([]byte(frag), &q); err != nil {
			continue
		}
		q.RawFragment = frag
		enrichLegacyQuestionXML(&q)
		if entry.ID != 0 {
			out[entry.ID] = q
		}
	}
	return out
}

// enrichLegacyQuestionXML fills fields from Moodle 2.x/3.x backup shape: <qtype>, plain <name>,
// <questiontext> text (not nested <text>), <defaultmark>, etc.
func enrichLegacyQuestionXML(q *QuestionXML) {
	if q.RawFragment == "" {
		return
	}
	raw := q.RawFragment
	if strings.TrimSpace(q.Type) == "" {
		q.Type = strings.TrimSpace(firstTagInnerText(raw, "qtype"))
	}
	if strings.TrimSpace(q.Name) == "" {
		q.Name = strings.TrimSpace(firstTagInnerText(raw, "name"))
	}
	if strings.TrimSpace(q.QuestionText.Text) == "" {
		q.QuestionText.Text = strings.TrimSpace(firstTagInnerText(raw, "questiontext"))
		if q.QuestionText.Text != "" && q.QuestionText.Format == "" {
			q.QuestionText.Format = "html"
		}
	}
	if strings.TrimSpace(q.GeneralFeedback.Text) == "" {
		q.GeneralFeedback.Text = strings.TrimSpace(firstTagInnerText(raw, "generalfeedback"))
	}
	if q.DefaultGrade == 0 {
		if dm := strings.TrimSpace(firstTagInnerText(raw, "defaultmark")); dm != "" {
			q.DefaultGrade = ParseFloat(dm)
		}
	}
}

func firstTagInnerText(raw, local string) string {
	re := regexp.MustCompile(`(?s)<` + regexp.QuoteMeta(local) + `[^>]*>(.*?)</` + regexp.QuoteMeta(local) + `>`)
	m := re.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}
