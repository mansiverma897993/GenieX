// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestPackageBuilds(t *testing.T) {}

// Intermediate chunks must serialize finish_reason as null, never "" (#1243).
func TestContentChunkFinishReasonNull(t *testing.T) {
	b, err := json.Marshal(contentChunk("hello"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"finish_reason":null`) {
		t.Errorf("intermediate chunk must carry finish_reason null, got: %s", s)
	}
	if strings.Contains(s, `"finish_reason":""`) {
		t.Errorf("intermediate chunk must not carry empty finish_reason, got: %s", s)
	}
	if !strings.Contains(s, `"content":"hello"`) {
		t.Errorf("content missing from delta, got: %s", s)
	}
	if !strings.Contains(s, `"object":"chat.completion.chunk"`) {
		t.Errorf("object field missing, got: %s", s)
	}
}

// The terminal chunk must carry the mapped finish_reason and an empty delta.
func TestFinishChunkStopWithEmptyDelta(t *testing.T) {
	b, err := json.Marshal(finishChunk("stop"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"finish_reason":"stop"`) {
		t.Errorf("final chunk must carry finish_reason stop, got: %s", s)
	}
	if !strings.Contains(s, `"delta":{}`) {
		t.Errorf("final chunk must carry an empty delta object, got: %s", s)
	}
}

// The usage chunk mirrors OpenAI's stream_options.include_usage shape:
// empty choices array plus a usage object.
func TestUsageChunkShape(t *testing.T) {
	b, err := json.Marshal(usageChunk(openai.CompletionUsage{
		CompletionTokens: 2,
		PromptTokens:     3,
		TotalTokens:      5,
	}))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"choices":[]`) {
		t.Errorf("usage chunk must carry an empty choices array, got: %s", s)
	}
	if !strings.Contains(s, `"usage":`) {
		t.Errorf("usage chunk must carry usage, got: %s", s)
	}
}

func TestMapFinishReason(t *testing.T) {
	cases := map[string]string{
		"length":        "length",
		"eos":           "stop",
		"stop_sequence": "stop",
		"user":          "stop",
		"":              "stop",
		"anything-else": "stop",
	}
	for in, want := range cases {
		if got := mapFinishReason(in); got != want {
			t.Errorf("mapFinishReason(%q) = %q, want %q", in, got, want)
		}
	}
}
