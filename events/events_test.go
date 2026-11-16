package events

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func dates(t *testing.T, market string, year int, id string) string {
	t.Helper()
	occ, err := Year(market, year)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, o := range occ {
		if o.ID == market+"-"+id {
			out = append(out, o.Start[5:])
		}
	}
	return strings.Join(out, " ")
}

// Dates published by national observatories, governments or trade bodies.
func TestKnownDates(t *testing.T) {
	cases := []struct {
		market string
		year   int
		id     string
		want   string
	}{
		{"jp", 2021, "setsubun", "02-02"}, {"jp", 2025, "setsubun", "02-02"}, {"jp", 2026, "setsubun", "02-03"},
		{"jp", 2024, "doyo-ushi", "07-24 08-05"}, {"jp", 2025, "doyo-ushi", "07-19 07-31"}, {"jp", 2026, "doyo-ushi", "07-26"},
		{"jp", 2024, "jugoya", "09-17"}, {"jp", 2025, "jugoya", "10-06"}, {"jp", 2026, "jugoya", "09-25"},
		{"jp", 2026, "shunbun", "03-20"}, {"jp", 2026, "shubun", "09-23"}, {"jp", 2026, "hahanohi", "05-10"},
		{"tw", 2026, "chunjie", "02-17"}, {"tw", 2026, "chuxi", "02-16"}, {"tw", 2026, "nianhuo", "02-03"},
		{"tw", 2026, "duanwu", "06-19"}, {"tw", 2026, "zhongqiu", "09-25"}, {"tw", 2026, "qingming", "04-05"},
		{"tw", 2027, "chunjie", "02-06"}, {"tw", 2026, "weiya", "02-03"}, {"tw", 2026, "qixi", "08-19"},
		{"kr", 2027, "seollal", "02-06"}, {"kr", 2026, "chuseok", "09-24"}, {"kr", 2026, "buddha", "05-24"},
		{"kr", 2025, "chobok", "07-20"}, {"kr", 2025, "jungbok", "07-30"}, {"kr", 2025, "malbok", "08-09"},
		{"kr", 2024, "chobok", "07-15"}, {"kr", 2024, "malbok", "08-14"},
		{"fr", 2025, "soldes-hiver", "01-08"}, {"fr", 2026, "soldes-hiver", "01-07"}, {"fr", 2025, "soldes-ete", "06-25"},
		{"fr", 2023, "fete-des-meres", "06-04"}, {"fr", 2026, "fete-des-meres", "05-31"}, {"fr", 2026, "paques", "04-05"},
		{"fr", 2025, "paques", "04-20"}, {"fr", 2026, "black-friday", "11-27"}, {"fr", 2026, "epiphanie", "01-04"},
	}
	for _, c := range cases {
		if got := dates(t, c.market, c.year, c.id); got != c.want {
			t.Errorf("%s %d %s = %q, want %q", c.market, c.year, c.id, got, c.want)
		}
	}
}

func TestEveryEventHappensEveryYear(t *testing.T) {
	for _, m := range Markets {
		for year := 1950; year <= 2090; year += 7 {
			occ, err := Year(m.Code, year)
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]int{}
			for i, o := range occ {
				seen[o.ID]++
				if o.End < o.Start || !strings.HasPrefix(o.Start, strings.Split(o.Start, "-")[0]) {
					t.Errorf("%s: bad span %s to %s", o.ID, o.Start, o.End)
				}
				if i > 0 && occ[i-1].Start > o.Start {
					t.Errorf("%s %d: not sorted", m.Code, year)
				}
				if o.NameLocal == "" || o.NameEN == "" || o.NameFR == "" || o.Rule == "" {
					t.Errorf("%s: a name or the rule is empty", o.ID)
				}
			}
			for _, d := range catalogue[m.Code] {
				n := seen[m.Code+"-"+d.id]
				if n < 1 || (n > 1 && d.id != "doyo-ushi") {
					t.Errorf("%s %d: %s happens %d times", m.Code, year, d.id, n)
				}
			}
		}
	}
}

func TestIdentifiersAreUnique(t *testing.T) {
	for market, defs := range catalogue {
		seen := map[string]bool{}
		for _, d := range defs {
			if seen[d.id] {
				t.Errorf("%s: duplicate id %s", market, d.id)
			}
			seen[d.id] = true
		}
	}
}

func TestUncertainDatesAreFlagged(t *testing.T) {
	// Lunar month 9 of 2057 starts within seconds of midnight in UTC+8: nothing in month 8 or
	// before depends on it, so mid-autumn 2057 is safe, but the mid-autumn of 2089 is not.
	for year, want := range map[int]bool{2026: false, 2057: false, 2089: true} {
		occ, _ := Year("tw", year)
		for _, o := range occ {
			if o.ID == "tw-zhongqiu" && o.Uncertain != want {
				t.Errorf("%d mid-autumn: uncertain = %v, want %v", year, o.Uncertain, want)
			}
		}
	}
}

func TestOutputsAreValidAndRepeatable(t *testing.T) {
	occ, err := Range([]string{"jp", "tw", "kr", "fr"}, 2026, 2027)
	if err != nil {
		t.Fatal(err)
	}
	var a, b, js, csv bytes.Buffer
	if err := WriteICS(&a, "Koyomi", occ); err != nil {
		t.Fatal(err)
	}
	_ = WriteICS(&b, "Koyomi", occ)
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatal("the ICS output changes between two runs")
	}
	events := 0
	for _, line := range strings.Split(strings.TrimSuffix(a.String(), "\r\n"), "\r\n") {
		if len(line) > 75 {
			t.Fatalf("ICS line longer than 75 octets: %q", line)
		}
		if strings.ContainsAny(line, "\r\n") {
			t.Fatalf("bare line break in %q", line)
		}
		if line == "BEGIN:VEVENT" {
			events++
		}
	}
	if events != len(occ) {
		t.Fatalf("%d VEVENT for %d events", events, len(occ))
	}
	if !strings.Contains(a.String(), "DTSTART;VALUE=DATE:20261225\r\nDTEND;VALUE=DATE:20261226") {
		t.Fatal("DTEND must be the day after the last day")
	}
	if err := WriteJSON(&js, occ); err != nil {
		t.Fatal(err)
	}
	var doc struct{ Events []Occurrence }
	if err := json.Unmarshal(js.Bytes(), &doc); err != nil || len(doc.Events) != len(occ) {
		t.Fatalf("JSON does not round-trip: %v", err)
	}
	_ = WriteCSV(&csv, occ)
	if rows := strings.Count(csv.String(), "\n"); rows != len(occ)+1 {
		t.Fatalf("%d CSV rows for %d events", rows, len(occ))
	}
}

func TestUnknownMarketAndYear(t *testing.T) {
	if _, err := Year("de", 2026); err == nil {
		t.Fatal("an unknown market must be an error")
	}
	if _, err := Year("jp", 1850); err == nil {
		t.Fatal("a year outside the range must be an error")
	}
}
