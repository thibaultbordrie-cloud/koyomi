package events

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Version is written in the outputs.
const Version = "0.1.0"

// WriteJSON writes the events with a short header that says how they were made.
func WriteJSON(w io.Writer, occ []Occurrence) error {
	doc := struct {
		Generator string       `json:"generator"`
		Notes     string       `json:"notes"`
		Events    []Occurrence `json:"events"`
	}{
		Generator: "koyomi " + Version,
		Notes: "Dates are computed, not copied: new moons and solar terms come from astronomy. " +
			"'end' is the last day, included. 'uncertain' marks dates that depend on a new moon or a " +
			"solar term less than a few minutes from midnight. This is not a public-holiday list.",
		Events: occ,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	return enc.Encode(doc)
}

// WriteCSV writes one row per event.
func WriteCSV(w io.Writer, occ []Occurrence) error {
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "market", "start", "end", "name_local", "name_en", "name_fr", "kind", "basis", "rule", "sells", "window", "uncertain"})
	for _, o := range occ {
		_ = cw.Write([]string{o.ID, o.Market, o.Start, o.End, o.NameLocal, o.NameEN, o.NameFR, o.Kind, o.Basis, o.Rule, o.Sells,
			strconv.FormatBool(o.Window), strconv.FormatBool(o.Uncertain)})
	}
	cw.Flush()
	return cw.Error()
}

// WriteText writes a table for the terminal.
func WriteText(w io.Writer, occ []Occurrence) error {
	for _, o := range occ {
		when := o.Start
		if o.End != o.Start {
			when += " → " + o.End[5:]
		}
		flag := ""
		if o.Uncertain {
			flag = "  (uncertain: close to midnight)"
		}
		if _, err := fmt.Fprintf(w, "%s  %-18s %-10s %s · %s%s\n", o.Market, when, o.Basis, o.NameLocal, o.NameEN, flag); err != nil {
			return err
		}
	}
	return nil
}

func icsEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\n", `\n`).Replace(s)
}

// fold writes one content line, folded at 75 octets without cutting a UTF-8 character (RFC 5545).
func fold(w io.Writer, line string) {
	limit := 75
	for len(line) > limit {
		cut := limit
		for cut > 0 && !utf8.RuneStart(line[cut]) {
			cut--
		}
		fmt.Fprintf(w, "%s\r\n ", line[:cut])
		line = line[cut:]
		limit = 74 // the leading space counts
	}
	fmt.Fprintf(w, "%s\r\n", line)
}

// WriteICS writes an iCalendar feed of all-day events. The output does not depend on the day
// it is generated: the same input gives the same bytes.
func WriteICS(w io.Writer, name string, occ []Occurrence) error {
	fold(w, "BEGIN:VCALENDAR")
	fold(w, "VERSION:2.0")
	fold(w, "PRODID:-//koyomi//commercial calendar "+Version+"//EN")
	fold(w, "CALSCALE:GREGORIAN")
	fold(w, "X-WR-CALNAME:"+icsEscape(name))
	for _, o := range occ {
		compact := func(jdn int) string { return strings.ReplaceAll(iso(jdn), "-", "") }
		summary := o.NameLocal
		if o.NameEN != o.NameLocal {
			summary += " · " + o.NameEN
		}
		desc := "Rule: " + o.Rule + "."
		if o.Sells != "" {
			desc += "\nSells: " + o.Sells + "."
		}
		if o.Window {
			desc += "\nThese dates are a trade habit, not a fixed day."
		}
		if o.Uncertain {
			desc += "\nUncertain: depends on an instant very close to midnight."
		}
		fold(w, "BEGIN:VEVENT")
		fold(w, "UID:"+o.ID+"-"+compact(o.start)+"@koyomi")
		fold(w, "DTSTAMP:"+compact(o.start)+"T000000Z")
		fold(w, "DTSTART;VALUE=DATE:"+compact(o.start))
		fold(w, "DTEND;VALUE=DATE:"+compact(o.end+1)) // DTEND is exclusive
		fold(w, "SUMMARY:"+icsEscape(summary))
		fold(w, "DESCRIPTION:"+icsEscape(desc))
		fold(w, "CATEGORIES:"+icsEscape(o.Kind))
		fold(w, "TRANSP:TRANSPARENT")
		fold(w, "END:VEVENT")
	}
	fold(w, "END:VCALENDAR")
	return nil
}

// WriteHTML writes the page published with the feeds.
func WriteHTML(w io.Writer, occ []Occurrence, from, to int) error {
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Koyomi 暦 · commercial calendar for Japan, Taiwan, Korea and France</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:60rem;margin:2rem auto;padding:0 1rem;color:#1f2328;background:#fff}
h1{margin-bottom:0}table{border-collapse:collapse;width:100%;font-size:.9rem}td,th{border-bottom:1px solid #d0d7de;padding:.3rem .5rem;text-align:left;vertical-align:top}
.past{color:#8c959f}.u{color:#9a6700}code{background:#f6f8fa;padding:.1rem .3rem;border-radius:4px}
@media(prefers-color-scheme:dark){body{background:#0d1117;color:#e6edf3}td,th{border-color:#30363d}code{background:#161b22}.past{color:#6e7681}}</style></head><body>
<h1>Koyomi 暦</h1><p>Commercial calendar for Japan, Taiwan, South Korea and France, `)
	fmt.Fprintf(&b, "%d to %d. Lunar dates and solar terms are computed from astronomy.</p>\n", from, to)
	b.WriteString(`<p>Subscribe in a calendar app: <a href="jp.ics">jp.ics</a> · <a href="tw.ics">tw.ics</a> · <a href="kr.ics">kr.ics</a> · <a href="fr.ics">fr.ics</a> · <a href="all.ics">all.ics</a>. Data: <a href="events.json">events.json</a> · <a href="events.csv">events.csv</a>. Source: <a href="https://github.com/thibaultbordrie-cloud/koyomi">github.com/thibaultbordrie-cloud/koyomi</a>.</p>
<table><thead><tr><th>Date</th><th>Market</th><th>Event</th><th>Rule</th><th>Sells</th></tr></thead><tbody>
`)
	for _, o := range occ {
		when := o.Start
		if o.End != o.Start {
			when += " → " + o.End[5:]
		}
		mark := ""
		if o.Uncertain {
			mark = ` <span class="u" title="close to midnight">(uncertain)</span>`
		}
		fmt.Fprintf(&b, "<tr data-end=\"%s\"><td>%s%s</td><td>%s</td><td>%s<br>%s</td><td>%s</td><td>%s</td></tr>\n",
			o.End, when, mark, strings.ToUpper(o.Market), html.EscapeString(o.NameLocal), html.EscapeString(o.NameEN),
			html.EscapeString(o.Rule), html.EscapeString(o.Sells))
	}
	b.WriteString(`</tbody></table>
<script>const today=new Date().toISOString().slice(0,10);let first=null;
for(const r of document.querySelectorAll("tr[data-end]")){if(r.dataset.end<today)r.className="past";else if(!first)first=r}
if(first)first.scrollIntoView({block:"center"})</script></body></html>
`)
	_, err := io.WriteString(w, b.String())
	return err
}
