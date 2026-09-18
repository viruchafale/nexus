package doctor

import (
	"fmt"
	"strings"
)

// Summary counts results by status.
type Summary struct {
	Passed   int
	Warnings int
	Failed   int
}

// Summarize tallies results.
func Summarize(results []Result) Summary {
	var s Summary
	for _, r := range results {
		switch r.Status {
		case Warn:
			s.Warnings++
		case Fail:
			s.Failed++
		default:
			s.Passed++
		}
	}
	return s
}

// Format renders results in the NEXUS DOCTOR layout: per-check rows,
// a SUMMARY tally, then Warnings and Recommendations for non-Pass rows.
func Format(results []Result) string {
	var b strings.Builder
	b.WriteString("NEXUS DOCTOR\n")
	b.WriteString("════════════════════════════════\n\n")
	for _, r := range results {
		fmt.Fprintf(&b, "%s %-20s %s\n", symbol(r.Status), r.Name, r.Status)
	}
	sum := Summarize(results)
	b.WriteString("\nSUMMARY\n")
	b.WriteString("────────────────────────────────\n")
	fmt.Fprintf(&b, "Passed: %d\nWarnings: %d\nFailed: %d\n", sum.Passed, sum.Warnings, sum.Failed)

	var warnings, recs []string
	for _, r := range results {
		if r.Status == Pass {
			continue
		}
		detail := r.Detail
		if detail == "" {
			detail = "no detail"
		}
		warnings = append(warnings, fmt.Sprintf("- %s: %s", r.Name, detail))
		if r.Recommendation != "" {
			recs = append(recs, "- "+r.Recommendation)
		}
	}
	if len(warnings) > 0 {
		b.WriteString("\nWarnings:\n")
		b.WriteString(strings.Join(warnings, "\n") + "\n")
	}
	if len(recs) > 0 {
		b.WriteString("\nRecommendations:\n")
		b.WriteString(strings.Join(recs, "\n") + "\n")
	}
	return b.String()
}

func symbol(s Status) string {
	switch s {
	case Warn:
		return "⚠"
	case Fail:
		return "✗"
	default:
		return "✓"
	}
}
