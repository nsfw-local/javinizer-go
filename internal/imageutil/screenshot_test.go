package imageutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDMMHost(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected bool
	}{
		{"pics.dmm.co.jp", "pics.dmm.co.jp", true},
		{"awsimgsrc.dmm.com", "awsimgsrc.dmm.com", true},
		{"dmm.co.jp", "dmm.co.jp", true},
		{"dmm.com", "dmm.com", true},
		{"example.com", "example.com", false},
		{"dmm.com.evil.com", "dmm.com.evil.com", false},
		{"empty", "", false},
		{"PICS.DMM.CO.JP uppercase", "PICS.DMM.CO.JP", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDMMHost(tt.host))
		})
	}
}

func TestNormalizeDMMScreenshotURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "AVSA-432 without jp (issue #23)",
			input:    "https://awsimgsrc.dmm.com/dig/digital/video/avsa00432/avsa00432-1.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/digital/video/avsa00432/avsa00432jp-1.jpg",
		},
		{
			name:     "Already has jp suffix",
			input:    "https://awsimgsrc.dmm.com/dig/digital/video/avsa00432/avsa00432jp-1.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/digital/video/avsa00432/avsa00432jp-1.jpg",
		},
		{
			name:     "pics.dmm.co.jp without jp",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "pics.dmm.co.jp already has jp",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "awsimgsrc CDN rewritten to pics.dmm.co.jp",
			input:    "https://awsimgsrc.dmm.co.jp/pics_dig/video/ipx00535/ipx00535-2.jpg",
			expected: "https://pics.dmm.co.jp/video/ipx00535/ipx00535jp-2.jpg",
		},
		{
			name:     "DMM prefix content ID without jp (1-digit prefix)",
			input:    "https://awsimgsrc.dmm.com/dig/digital/video/1sdmm00132/1sdmm00132-1.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/digital/video/1sdmm132/1sdmm132jp-1.jpg",
		},
		{
			name:     "Cover image pl.jpg unchanged",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
		},
		{
			name:     "Poster image ps.jpg unchanged",
			input:    "https://awsimgsrc.dmm.com/dig/video/ipx00535/ipx00535ps.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/video/ipx00535/ipx00535ps.jpg",
		},
		{
			name:     "Protocol-relative URL upgraded to https",
			input:    "//pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "Query parameters stripped",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpg?x=1",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "Fragment stripped",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpg#anchor",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "awsimgsrc rewrite + query strip + jp insertion combined",
			input:    "https://awsimgsrc.dmm.co.jp/pics_dig/video/test/test-3.jpg?v=1",
			expected: "https://pics.dmm.co.jp/video/test/testjp-3.jpg",
		},
		{
			name:     "Content ID canonicalization 118abp00880",
			input:    "https://pics.dmm.co.jp/digital/video/118abp00880/118abp00880-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/118abp880/118abp880jp-1.jpg",
		},
		{
			name:     "Non-DMM URL unchanged",
			input:    "https://example.com/images/screenshot-1.jpg",
			expected: "https://example.com/images/screenshot-1.jpg",
		},
		{
			name:     "Empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "Whitespace trimmed",
			input:    "  https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpg  ",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "jpeg extension also handled",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535-1.jpeg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpeg",
		},
		{
			name:     "Non-image DMM URL with dash unchanged",
			input:    "https://pics.dmm.co.jp/digital/video/ipx00535/something-else.png",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/something-else.png",
		},
		{
			name:     "Amateur video path lowercased",
			input:    "https://pics.dmm.co.jp/digital/amateur/ORECO183/ORECO183-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183jp-1.jpg",
		},
		{
			name:     "Amateur video path already lowercase",
			input:    "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183jp-1.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeDMMScreenshotURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpgradeCoverResolution(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "ps.jpg to pl.jpg",
			url:      "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535ps.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
		},
		{
			name:     "jp.jpg to pl.jpg for non-amateur",
			url:      "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
		},
		{
			name:     "amateur jp.jpg unchanged",
			url:      "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183jp.jpg",
			expected: "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183jp.jpg",
		},
		{
			name:     "amateur ps.jpg upgraded to pl.jpg",
			url:      "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183ps.jpg",
			expected: "https://pics.dmm.co.jp/digital/amateur/oreco183/oreco183pl.jpg",
		},
		{
			name:     "pl.jpg unchanged",
			url:      "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535pl.jpg",
		},
		{
			name:     "ps.jpg upgraded to pl.jpg",
			url:      "https://pics.dmm.co.jp/digital/video/test/testps.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/test/testpl.jpg",
		},
		{
			name:     "non-cover URL unchanged",
			url:      "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
			expected: "https://pics.dmm.co.jp/digital/video/ipx00535/ipx00535jp-1.jpg",
		},
		{
			name:     "empty string unchanged",
			url:      "",
			expected: "",
		},
		{
			name:     "awsimgsrc amateur jp.jpg unchanged",
			url:      "https://awsimgsrc.dmm.com/dig/amateur/oreco183/oreco183jp.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/amateur/oreco183/oreco183jp.jpg",
		},
		{
			name:     "awsimgsrc amateur ps.jpg upgraded to pl.jpg",
			url:      "https://awsimgsrc.dmm.com/dig/amateur/oreco183/oreco183ps.jpg",
			expected: "https://awsimgsrc.dmm.com/dig/amateur/oreco183/oreco183pl.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpgradeCoverResolution(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}
