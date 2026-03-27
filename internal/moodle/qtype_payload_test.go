package moodle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPayloadMultichoice(t *testing.T) {
	raw := `<question type="multichoice">
  <single>true</single>
  <shuffleanswers>true</shuffleanswers>
  <answer fraction="1"><text>Yes</text></answer>
  <answer fraction="0"><text>No</text></answer>
</question>`
	j := BuildPayloadJSON("multichoice", []byte(raw))
	require.Contains(t, j, `"qtype":"multichoice"`)
	require.Contains(t, j, "Yes")
}

func TestBuildPayloadTruefalse(t *testing.T) {
	raw := `<question type="truefalse">
  <answer fraction="1"><text>true</text></answer>
  <answer fraction="0"><text>false</text></answer>
</question>`
	j := BuildPayloadJSON("truefalse", []byte(raw))
	require.True(t, strings.Contains(j, "true"))
}
