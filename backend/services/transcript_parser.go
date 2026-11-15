package services

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	cueArrowLineRegex     = regexp.MustCompile(`^\s*((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)\s*-->\s*((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)(?:\s+.*)?$`)
	vttVoiceTagRegex      = regexp.MustCompile(`^<v(?:\.[^>\s]+)*\s+([^>]+)>(.*)$`)
	htmlTagStripRegex     = regexp.MustCompile(`<[^>]+>`)
	leadingBracketTSRegex = regexp.MustCompile(`^\s*[\[\(]\s*((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)\s*[\]\)]\s*(?:[-–—]\s*)?(.*)$`)
	leadingBareTSRegex    = regexp.MustCompile(`^\s*((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)\s+(?:[-–—]\s*)?(.+)$`)
	trailingBracketTS     = regexp.MustCompile(`^(.*?)\s*[\[\(]\s*((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)\s*[\]\)]\s*$`)
	trailingBareTS        = regexp.MustCompile(`^(.*?)\s+((?:\d{1,2}:)?\d{1,2}:\d{2}(?:[.,]\d{1,3})?)\s*$`)
	digitsOnlyRegex       = regexp.MustCompile(`^\d+$`)
	bulletPrefixRegex     = regexp.MustCompile(`^\s*(?:[-*•]+|\d+[.)])\s+`)
)

// ParseTranscriptText parses a raw transcript string (plain text, speaker-labeled dialogue,
// timestamped lines, WebVTT, SRT, or JSON segment data) into ordered ImportedSegment values.
func ParseTranscriptText(raw string, defaultSpeaker string) []ImportedSegment {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimPrefix(normalized, "\uFEFF")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil
	}

	defSpeaker := strings.TrimSpace(defaultSpeaker)
	if defSpeaker == "" {
		defSpeaker = "Speaker 1"
	}

	// 1. JSON format (segment array, wrapper object, or Speechmatics results JSON)
	if strings.HasPrefix(normalized, "{") || strings.HasPrefix(normalized, "[") {
		if segs := parseTranscriptJSON(normalized, defSpeaker); len(segs) > 0 {
			return finalizeParsedSegments(segs, true)
		}
	}

	// 2. WebVTT / SRT cue format
	lines := strings.Split(normalized, "\n")
	if hasCueArrowLines(lines) {
		if segs := parseCueLines(lines, defSpeaker); len(segs) > 0 {
			return finalizeParsedSegments(segs, false)
		}
	}

	// 3. Speaker-labeled / timestamped / plain text paragraphs
	segs := parseDialogueOrPlainLines(lines, defSpeaker)
	if len(segs) == 0 {
		return nil
	}
	return finalizeParsedSegments(segs, false)
}

func parseTranscriptJSON(trimmed string, defaultSpeaker string) []ImportedSegment {
	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}

	switch v := parsed.(type) {
	case []interface{}:
		return parseGenericSegmentSlice(v, defaultSpeaker)
	case map[string]interface{}:
		if results, ok := v["results"].([]interface{}); ok {
			if segs := parseSpeechmaticsResultsToSegments(results); len(segs) > 0 {
				return segs
			}
		}
		for _, key := range []string{"segments", "transcript", "sentences", "utterances"} {
			if arr, ok := v[key].([]interface{}); ok {
				if segs := parseGenericSegmentSlice(arr, defaultSpeaker); len(segs) > 0 {
					return segs
				}
			}
		}
	}
	return nil
}

func parseGenericSegmentSlice(arr []interface{}, defaultSpeaker string) []ImportedSegment {
	segs := make([]ImportedSegment, 0, len(arr))
	for _, elem := range arr {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		txt := firstNonEmptyString(m, "text", "raw_text", "content")
		if txt == "" {
			continue
		}
		spk := firstNonEmptyString(m, "speaker_name", "speakerName")
		if spk == "" {
			if spObj, ok := m["speaker"].(map[string]interface{}); ok {
				spk = firstNonEmptyString(spObj, "display_name", "name")
			} else if spStr, ok := m["speaker"].(string); ok {
				spk = strings.TrimSpace(spStr)
			}
		}
		if spk == "" {
			spk = defaultSpeaker
		}

		startSec := parseFlexibleSeconds(m["start_time"])
		if startSec == 0 && m["start"] != nil {
			startSec = parseFlexibleSeconds(m["start"])
		}
		if startSec == 0 && m["timestamp"] != nil {
			startSec = parseFlexibleSeconds(m["timestamp"])
		}

		endSec := parseFlexibleSeconds(m["end_time"])
		if endSec == 0 && m["end"] != nil {
			endSec = parseFlexibleSeconds(m["end"])
		}

		segs = append(segs, ImportedSegment{
			StartTime: startSec,
			EndTime:   endSec,
			Speaker:   spk,
			Text:      txt,
		})
	}
	return segs
}

func parseSpeechmaticsResultsToSegments(results []interface{}) []ImportedSegment {
	var segs []ImportedSegment
	var currentSpeaker *string
	var currentText []string
	var segmentStart float64
	var segmentEnd float64

	for _, rawItem := range results {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		tokenType, _ := item["type"].(string)
		alternatives, _ := item["alternatives"].([]interface{})
		if len(alternatives) == 0 {
			continue
		}
		alt, ok := alternatives[0].(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := alt["content"].(string)
		speaker, _ := alt["speaker"].(string)
		if content == "" {
			continue
		}

		startTime, _ := item["start_time"].(float64)
		var endTime float64
		if endVal, ok := item["end_time"].(float64); ok {
			endTime = endVal
		} else {
			dur, _ := item["duration"].(float64)
			endTime = startTime + dur
		}

		if tokenType == "word" {
			spk := "Unknown"
			if strings.TrimSpace(speaker) != "" {
				spk = strings.TrimSpace(speaker)
			}
			if currentSpeaker == nil {
				currentSpeaker = &spk
				segmentStart = startTime
				segmentEnd = endTime
				currentText = []string{content}
			} else if *currentSpeaker == spk {
				currentText = append(currentText, content)
				segmentEnd = endTime
			} else {
				segs = append(segs, ImportedSegment{
					StartTime: segmentStart,
					EndTime:   segmentEnd,
					Speaker:   *currentSpeaker,
					Text:      strings.Join(currentText, " "),
				})
				currentSpeaker = &spk
				segmentStart = startTime
				segmentEnd = endTime
				currentText = []string{content}
			}
		} else if tokenType == "punctuation" && len(currentText) > 0 {
			if content == "." || content == "," || content == "!" || content == "?" || content == ";" || content == ":" {
				currentText[len(currentText)-1] = currentText[len(currentText)-1] + content
			} else {
				currentText = append(currentText, content)
			}
			segmentEnd = endTime
		}
	}

	if len(currentText) > 0 && currentSpeaker != nil {
		segs = append(segs, ImportedSegment{
			StartTime: segmentStart,
			EndTime:   segmentEnd,
			Speaker:   *currentSpeaker,
			Text:      strings.Join(currentText, " "),
		})
	}
	return segs
}

func hasCueArrowLines(lines []string) bool {
	for _, line := range lines {
		if cueArrowLineRegex.MatchString(line) {
			return true
		}
	}
	return false
}

func parseCueLines(lines []string, defaultSpeaker string) []ImportedSegment {
	var segs []ImportedSegment
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "WEBVTT") {
			i++
			continue
		}
		if strings.HasPrefix(line, "NOTE") || strings.HasPrefix(line, "STYLE") || strings.HasPrefix(line, "REGION") {
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				i++
			}
			continue
		}

		matches := cueArrowLineRegex.FindStringSubmatch(line)
		if matches == nil {
			i++
			continue
		}

		startSec, okStart := parseTimestampToken(matches[1])
		endSec, okEnd := parseTimestampToken(matches[2])
		if !okStart || !okEnd {
			i++
			continue
		}

		i++
		var cueLines []string
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
			if cueArrowLineRegex.MatchString(lines[i]) {
				break
			}
			cueLines = append(cueLines, strings.TrimSpace(lines[i]))
			i++
		}

		if len(cueLines) == 0 {
			continue
		}

		speaker := defaultSpeaker
		cleanedLines := make([]string, 0, len(cueLines))
		for idx, rawCueLine := range cueLines {
			l := rawCueLine
			if vm := vttVoiceTagRegex.FindStringSubmatch(l); vm != nil {
				if idx == 0 && strings.TrimSpace(vm[1]) != "" {
					speaker = strings.TrimSpace(vm[1])
				}
				l = vm[2]
			}
			l = strings.TrimSpace(htmlTagStripRegex.ReplaceAllString(l, ""))
			if l == "" {
				continue
			}
			if idx == 0 && speaker == defaultSpeaker {
				if spk, rest, ok := extractSpeakerColon(l); ok {
					speaker = spk
					l = rest
				}
			}
			if strings.TrimSpace(l) != "" {
				cleanedLines = append(cleanedLines, strings.TrimSpace(l))
			}
		}

		text := strings.TrimSpace(strings.Join(cleanedLines, " "))
		if text == "" {
			continue
		}

		segs = append(segs, ImportedSegment{
			StartTime: startSec,
			EndTime:   endSec,
			Speaker:   speaker,
			Text:      text,
		})
	}
	return segs
}

type rawParsedSegment struct {
	hasExplicitTime bool
	startTime       float64
	speaker         string
	lines           []string
}

func parseDialogueOrPlainLines(lines []string, defaultSpeaker string) []ImportedSegment {
	var parsed []rawParsedSegment
	var current *rawParsedSegment
	lastSpeaker := defaultSpeaker
	pendingHeaderSpeaker := ""
	pendingHeaderHasTime := false
	pendingHeaderTime := 0.0
	sawBlankSinceCurrent := false

	flushCurrent := func() {
		if current != nil {
			txt := strings.TrimSpace(strings.Join(current.lines, "\n"))
			if txt != "" {
				parsed = append(parsed, *current)
			}
			current = nil
		}
	}

	for idx, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" {
			sawBlankSinceCurrent = true
			continue
		}

		// Case 1: Two-line speaker+timestamp header (e.g., "Alice Smith  0:15" or "0:15 Alice Smith")
		if spk, ts, ok := matchSpeakerTimestampHeader(trimmed, lines, idx); ok {
			flushCurrent()
			pendingHeaderSpeaker = spk
			pendingHeaderHasTime = true
			pendingHeaderTime = ts
			lastSpeaker = spk
			sawBlankSinceCurrent = false
			continue
		}

		// Case 2: Line starts with optional timestamp + "Speaker: text" (or "Speaker:" alone)
		if ts, hasTS, spk, rest, ok := matchInlineSpeakerOrTimestamp(trimmed); ok {
			flushCurrent()
			if spk != "" {
				lastSpeaker = spk
			} else if pendingHeaderSpeaker != "" {
				spk = pendingHeaderSpeaker
			} else {
				spk = defaultSpeaker
			}
			pendingHeaderSpeaker = ""
			pendingHeaderHasTime = false

			current = &rawParsedSegment{
				hasExplicitTime: hasTS,
				startTime:       ts,
				speaker:         spk,
			}
			if strings.TrimSpace(rest) != "" {
				current.lines = append(current.lines, strings.TrimSpace(rest))
			}
			sawBlankSinceCurrent = false
			continue
		}

		// Case 3: Plain line following a pending speaker+timestamp header
		if pendingHeaderSpeaker != "" {
			flushCurrent()
			current = &rawParsedSegment{
				hasExplicitTime: pendingHeaderHasTime,
				startTime:       pendingHeaderTime,
				speaker:         pendingHeaderSpeaker,
				lines:           []string{trimmed},
			}
			pendingHeaderSpeaker = ""
			pendingHeaderHasTime = false
			sawBlankSinceCurrent = false
			continue
		}

		// Case 4: Continuation line or new paragraph
		if current == nil || sawBlankSinceCurrent {
			flushCurrent()
			current = &rawParsedSegment{
				speaker: lastSpeaker,
				lines:   []string{trimmed},
			}
		} else {
			current.lines = append(current.lines, trimmed)
		}
		sawBlankSinceCurrent = false
	}

	flushCurrent()

	segs := make([]ImportedSegment, 0, len(parsed))
	for _, item := range parsed {
		txt := strings.TrimSpace(strings.Join(item.lines, "\n"))
		if txt == "" {
			continue
		}
		start := -1.0
		if item.hasExplicitTime {
			start = item.startTime
		}
		segs = append(segs, ImportedSegment{
			StartTime: start,
			EndTime:   0,
			Speaker:   item.speaker,
			Text:      txt,
		})
	}
	return segs
}

func matchSpeakerTimestampHeader(line string, lines []string, idx int) (string, float64, bool) {
	// Must have a following non-empty line to act as a header
	hasNextContent := false
	for j := idx + 1; j < len(lines); j++ {
		if strings.TrimSpace(lines[j]) != "" {
			hasNextContent = true
			break
		}
	}
	if !hasNextContent {
		return "", 0, false
	}

	// Trailing bracketed timestamp: "Alice Smith (0:15)" or "Alice Smith [00:15]"
	if m := trailingBracketTS.FindStringSubmatch(line); m != nil {
		if ts, ok := parseTimestampToken(m[2]); ok {
			if spk, okSpk := normalizeSpeakerCandidate(m[1]); okSpk {
				return spk, ts, true
			}
		}
	}

	// Trailing bare timestamp: "Alice Smith  0:15"
	if m := trailingBareTS.FindStringSubmatch(line); m != nil {
		if ts, ok := parseTimestampToken(m[2]); ok {
			candidate := strings.TrimSpace(m[1])
			lower := strings.ToLower(candidate)
			// Avoid matching prose sentences ending with a clock time like "Meet at 10:30"
			if !hasTrailingTimePreposition(lower) {
				if spk, okSpk := normalizeSpeakerCandidate(candidate); okSpk {
					return spk, ts, true
				}
			}
		}
	}

	// Leading timestamp + Speaker without colon: "[0:15] Alice Smith" or "0:15 - Alice Smith"
	if m := leadingBracketTSRegex.FindStringSubmatch(line); m != nil && strings.TrimSpace(m[2]) != "" {
		rest := strings.TrimSpace(m[2])
		if !strings.Contains(rest, ":") {
			if ts, ok := parseTimestampToken(m[1]); ok {
				if spk, okSpk := normalizeSpeakerCandidate(rest); okSpk {
					return spk, ts, true
				}
			}
		}
	}
	if m := leadingBareTSRegex.FindStringSubmatch(line); m != nil {
		rest := strings.TrimSpace(m[2])
		if !strings.Contains(rest, ":") {
			if ts, ok := parseTimestampToken(m[1]); ok {
				if spk, okSpk := normalizeSpeakerCandidate(rest); okSpk {
					return spk, ts, true
				}
			}
		}
	}

	return "", 0, false
}

func hasTrailingTimePreposition(lower string) bool {
	for _, prep := range []string{" at", " by", " until", " till", " around", " before", " after", " from", " since", " to"} {
		if strings.HasSuffix(lower, prep) {
			return true
		}
	}
	return false
}

func matchInlineSpeakerOrTimestamp(line string) (ts float64, hasTS bool, speaker string, rest string, ok bool) {
	working := line

	if m := leadingBracketTSRegex.FindStringSubmatch(working); m != nil {
		if parsedTS, valid := parseTimestampToken(m[1]); valid {
			ts = parsedTS
			hasTS = true
			working = strings.TrimSpace(m[2])
		}
	} else if m := leadingBareTSRegex.FindStringSubmatch(working); m != nil {
		if parsedTS, valid := parseTimestampToken(m[1]); valid {
			ts = parsedTS
			hasTS = true
			working = strings.TrimSpace(m[2])
		}
	}

	if working == "" {
		if hasTS {
			return ts, true, "", "", true
		}
		return 0, false, "", "", false
	}

	if spk, textAfter, found := extractSpeakerColon(working); found {
		return ts, hasTS, spk, textAfter, true
	}

	if hasTS {
		return ts, true, "", working, true
	}
	return 0, false, "", "", false
}

func extractSpeakerColon(line string) (string, string, bool) {
	if bulletPrefixRegex.MatchString(line) {
		return "", "", false
	}

	// Support "**Speaker:** text" and "**Speaker**: text" and "Speaker (00:12): text"
	colonIdx := strings.Index(line, ":")
	// If colon is inside a trailing timestamp like "Alice (0:12): text", find the colon after ")" or "]"
	if colonIdx != -1 {
		closeParen := strings.IndexAny(line, ")]")
		if closeParen != -1 && closeParen > colonIdx {
			nextColon := strings.Index(line[closeParen:], ":")
			if nextColon != -1 {
				colonIdx = closeParen + nextColon
			}
		}
	}
	if colonIdx <= 0 || colonIdx > 65 {
		return "", "", false
	}

	after := line[colonIdx+1:]
	// Colon must be at end of line or followed by whitespace (prevents matching URLs, ratios, times like 10:30)
	if after != "" && after[0] != ' ' && after[0] != '\t' && after[0] != '*' {
		return "", "", false
	}

	rawSpeaker := strings.TrimSpace(line[:colonIdx])
	// Strip trailing markdown bold if written as "**Alice:**"
	rawSpeaker = strings.TrimSuffix(rawSpeaker, "**")
	rawSpeaker = strings.TrimPrefix(rawSpeaker, "**")
	rawSpeaker = strings.TrimSpace(rawSpeaker)

	// Strip trailing timestamp in speaker label like "Alice (00:12)" or "Alice [00:12]"
	if m := trailingBracketTS.FindStringSubmatch(rawSpeaker); m != nil {
		if _, okTS := parseTimestampToken(m[2]); okTS {
			rawSpeaker = strings.TrimSpace(m[1])
		}
	}

	spk, ok := normalizeSpeakerCandidate(rawSpeaker)
	if !ok {
		return "", "", false
	}

	rest := strings.TrimSpace(after)
	rest = strings.TrimPrefix(rest, "**")
	rest = strings.TrimSpace(rest)
	return spk, rest, true
}

func normalizeSpeakerCandidate(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "*_`")
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 50 {
		return "", false
	}
	if bulletPrefixRegex.MatchString(s) {
		return "", false
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") || strings.Contains(lower, "mailto:") {
		return "", false
	}
	if digitsOnlyRegex.MatchString(s) {
		return "", false
	}
	if strings.HasSuffix(s, "?") || strings.HasSuffix(s, "!") {
		return "", false
	}
	if strings.HasSuffix(s, ".") && !isAllowedAbbreviation(lower) {
		return "", false
	}

	words := strings.Fields(s)
	if len(words) == 0 || len(words) > 6 {
		return "", false
	}

	hasLetter := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127 {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return "", false
	}

	return s, true
}

func isAllowedAbbreviation(lower string) bool {
	for _, abbr := range []string{"dr.", "mr.", "mrs.", "ms.", "prof.", "jr.", "sr."} {
		if strings.HasSuffix(lower, abbr) {
			return true
		}
	}
	// Single-letter initial like "John D."
	if len(lower) >= 3 && lower[len(lower)-1] == '.' && lower[len(lower)-3] == ' ' {
		return true
	}
	return false
}

func parseTimestampToken(token string) (float64, bool) {
	t := strings.TrimSpace(token)
	t = strings.ReplaceAll(t, ",", ".")
	parts := strings.Split(t, ":")
	switch len(parts) {
	case 2:
		mins, err1 := strconv.ParseFloat(parts[0], 64)
		secs, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil || mins < 0 || secs < 0 || secs >= 60 {
			return 0, false
		}
		return mins*60 + secs, true
	case 3:
		hrs, err0 := strconv.ParseFloat(parts[0], 64)
		mins, err1 := strconv.ParseFloat(parts[1], 64)
		secs, err2 := strconv.ParseFloat(parts[2], 64)
		if err0 != nil || err1 != nil || err2 != nil || hrs < 0 || mins < 0 || mins >= 60 || secs < 0 || secs >= 60 {
			return 0, false
		}
		return hrs*3600 + mins*60 + secs, true
	default:
		return 0, false
	}
}

func finalizeParsedSegments(segs []ImportedSegment, fromJSON bool) []ImportedSegment {
	if len(segs) == 0 {
		return nil
	}

	// Ensure deterministic, monotonically non-decreasing start times so DB/UI sorting
	// preserves exact input order even when no timestamps were supplied.
	lastStart := -0.001
	for i := range segs {
		hasTime := segs[i].StartTime >= 0
		if fromJSON && i > 0 && segs[i].StartTime == 0 && segs[i-1].StartTime == 0 {
			hasTime = false
		}
		if !hasTime {
			nextStart := math.Max(0, lastStart+0.001)
			segs[i].StartTime = math.Round(nextStart*1000) / 1000
			if segs[i].EndTime < segs[i].StartTime {
				segs[i].EndTime = segs[i].StartTime
			}
		} else {
			if segs[i].StartTime < lastStart {
				segs[i].StartTime = math.Round((lastStart+0.001)*1000) / 1000
			}
			if segs[i].EndTime < segs[i].StartTime {
				segs[i].EndTime = segs[i].StartTime
			}
		}
		lastStart = segs[i].StartTime
	}

	return segs
}
