package moodle

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	"strings"
)

// BuildPayloadJSON converts parsed question XML into a small JSON payload for v1 types.
func BuildPayloadJSON(qType string, xmlBytes []byte) string {
	t := strings.TrimSpace(strings.ToLower(qType))
	switch t {
	case "multichoice":
		var mc struct {
			Single          string `xml:"single"`
			ShuffleAnswers  string `xml:"shuffleanswers"`
			AnswerNumbering string `xml:"answernumbering"`
			Answers         []struct {
				Fraction string `xml:"fraction,attr"`
				Text     string `xml:"text"`
			} `xml:"answer"`
		}
		_ = xml.Unmarshal(xmlBytes, &mc)
		type ans struct {
			Text     string  `json:"text"`
			Fraction float64 `json:"fraction"`
		}
		var answers []ans
		for _, a := range mc.Answers {
			f, _ := strconv.ParseFloat(strings.TrimSpace(a.Fraction), 64)
			if f > 1.001 {
				f /= 100
			}
			answers = append(answers, ans{Text: strings.TrimSpace(a.Text), Fraction: f})
		}
		out := map[string]any{
			"qtype":           "multichoice",
			"single":          strings.EqualFold(strings.TrimSpace(mc.Single), "true") || mc.Single == "1",
			"shuffleanswers":  strings.EqualFold(strings.TrimSpace(mc.ShuffleAnswers), "true") || mc.ShuffleAnswers == "1",
			"answernumbering": strings.TrimSpace(mc.AnswerNumbering),
			"answers":         answers,
		}
		b, _ := json.Marshal(out)
		return string(b)
	case "truefalse":
		var tf struct {
			Answers []struct {
				Fraction string `xml:"fraction,attr"`
				Text     string `xml:"text"`
			} `xml:"answer"`
		}
		_ = xml.Unmarshal(xmlBytes, &tf)
		out := map[string]any{"qtype": "truefalse"}
		for _, a := range tf.Answers {
			f, _ := strconv.ParseFloat(strings.TrimSpace(a.Fraction), 64)
			if f > 1.001 {
				f /= 100
			}
			lab := strings.ToLower(strings.TrimSpace(a.Text))
			if f > 0.99 {
				out["correct"] = lab
			}
		}
		if _, ok := out["correct"]; !ok {
			// Moodle 3.x backup: answers use <answertext> under plugin_qtype_truefalse_question.
			var leg struct {
				Plugin struct {
					Answers struct {
						List []struct {
							AnsText  string `xml:"answertext"`
							Fraction string `xml:"fraction"`
						} `xml:"answer"`
					} `xml:"answers"`
				} `xml:"plugin_qtype_truefalse_question"`
			}
			_ = xml.Unmarshal(xmlBytes, &leg)
			for _, a := range leg.Plugin.Answers.List {
				f, _ := strconv.ParseFloat(strings.TrimSpace(a.Fraction), 64)
				if f > 1.001 {
					f /= 100
				}
				if f > 0.99 {
					out["correct"] = strings.ToLower(strings.TrimSpace(a.AnsText))
					break
				}
			}
		}
		b, _ := json.Marshal(out)
		return string(b)
	case "shortanswer":
		var sa struct {
			Answers []struct {
				Fraction string `xml:"fraction,attr"`
				Text     string `xml:"text"`
			} `xml:"answer"`
		}
		_ = xml.Unmarshal(xmlBytes, &sa)
		var answers []map[string]any
		for _, a := range sa.Answers {
			f, _ := strconv.ParseFloat(strings.TrimSpace(a.Fraction), 64)
			if f > 1.001 {
				f /= 100
			}
			answers = append(answers, map[string]any{"text": strings.TrimSpace(a.Text), "fraction": f})
		}
		b, _ := json.Marshal(map[string]any{"qtype": "shortanswer", "answers": answers})
		return string(b)
	case "essay":
		var es struct {
			ResponseFormat string `xml:"responseformat"`
			Attachments    string `xml:"attachments"`
		}
		_ = xml.Unmarshal(xmlBytes, &es)
		b, _ := json.Marshal(map[string]any{
			"qtype":          "essay",
			"responseformat": strings.TrimSpace(es.ResponseFormat),
			"attachments":    strings.TrimSpace(es.Attachments),
		})
		return string(b)
	case "match":
		var m struct {
			ShuffleAnswers string `xml:"shuffleanswers"`
			Subquestions   []struct {
				QuestionText struct {
					Text struct {
						Text string `xml:"text"`
					} `xml:"text"`
				} `xml:"questiontext"`
				AnswerText struct {
					Text struct {
						Text string `xml:"text"`
					} `xml:"text"`
				} `xml:"answertext"`
			} `xml:"subquestion"`
		}
		_ = xml.Unmarshal(xmlBytes, &m)
		type pair struct {
			Stem   string `json:"stem"`
			Answer string `json:"answer"`
		}
		var pairs []pair
		for _, s := range m.Subquestions {
			pairs = append(pairs, pair{
				Stem:   strings.TrimSpace(s.QuestionText.Text.Text),
				Answer: strings.TrimSpace(s.AnswerText.Text.Text),
			})
		}
		b, _ := json.Marshal(map[string]any{
			"qtype":          "match",
			"shuffleanswers": strings.EqualFold(strings.TrimSpace(m.ShuffleAnswers), "true"),
			"pairs":          pairs,
		})
		return string(b)
	case "description":
		b, _ := json.Marshal(map[string]any{"qtype": "description"})
		return string(b)
	default:
		b, _ := json.Marshal(map[string]any{"qtype": t, "unsupported": true})
		return string(b)
	}
}
