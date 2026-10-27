package main

// Reading the add-on's saved data file. The guild website reads it the same way when a player uploads the file:
// only the code the add-on saved at logout is taken out of the file; nothing else in it is read or sent.

import (
	"regexp"
	"strconv"
)

const savedFile = "FameAndFortune.lua"

type savedCode struct {
	Codes   []string
	At      int64
	Version string
	Reason  string // "", "not-ours", "no-code" or "turned-off"
}

var (
	reOurs    = regexp.MustCompile(`FameAndFortuneDB\s*=`)
	reSingle  = regexp.MustCompile(`\["latestCode"\]\s*=\s*"(FF\d+:[0-9a-fA-F]+:[^"\s]+)"`)
	reParts   = regexp.MustCompile(`\["latestParts"\]\s*=\s*\{([^}]*)\}`)
	reCode    = regexp.MustCompile(`"(FF\d+:[0-9a-fA-F]+:[^"\s]+)"`)
	reAt      = regexp.MustCompile(`\["latestCodeAt"\]\s*=\s*(\d+)`)
	reVersion = regexp.MustCompile(`\["latestCodeVersion"\]\s*=\s*"([^"]{1,20})"`)
	reOff     = regexp.MustCompile(`\["saveCode"\]\s*=\s*false`)
)

func readSavedCode(text string) savedCode {
	if !reOurs.MatchString(text) {
		return savedCode{Reason: "not-ours"}
	}
	var out savedCode
	if m := reSingle.FindStringSubmatch(text); m != nil {
		out.Codes = []string{m[1]}
	} else if m := reParts.FindStringSubmatch(text); m != nil {
		for _, c := range reCode.FindAllStringSubmatch(m[1], -1) {
			out.Codes = append(out.Codes, c[1])
		}
	}
	if len(out.Codes) == 0 {
		if reOff.MatchString(text) {
			return savedCode{Reason: "turned-off"}
		}
		return savedCode{Reason: "no-code"}
	}
	if m := reAt.FindStringSubmatch(text); m != nil {
		out.At, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := reVersion.FindStringSubmatch(text); m != nil {
		out.Version = m[1]
	}
	return out
}
