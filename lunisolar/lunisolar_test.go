package lunisolar

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/thibaultbordrie-cloud/koyomi/astro"
)

func mmdd(jdn int) string {
	_, m, d := astro.Civil(jdn)
	return fmt.Sprintf("%02d%02d", m, d)
}

func lines(t *testing.T, name string) []string {
	t.Helper()
	f, err := os.Open("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if !strings.HasPrefix(sc.Text(), "#") {
			out = append(out, sc.Text())
		}
	}
	return out
}

// In three months the reference table follows calendars printed at the time, which were a day
// off: the new moons of 1933-07-23 (00:03) and 1978-09-03 (00:09) fell just after midnight,
// and the new moon of 1954-11-25 was at 20:30. Two other libraries (lunar_python,
// LunarCalendar) agree with the astronomy. Each case shows as a 29/30 swap in one year.
var printedDifferently = map[int]bool{1933: true, 1954: true, 1978: true}

func TestChineseCalendarAgreesWithReferenceTable(t *testing.T) {
	months, differing := 0, 0
	for _, line := range lines(t, "chinese.txt") {
		var year, leap int
		var newYear, bits string
		fmt.Sscanf(line, "%d %s %d %s", &year, &newYear, &leap, &bits)
		got := Chinese.Months(year)
		mine := ""
		for _, m := range got {
			mine += map[bool]string{true: "1", false: "0"}[m.Days == 30]
		}
		months += len(got)
		if mmdd(got[0].Start) != newYear || Chinese.LeapMonth(year) != leap {
			t.Errorf("%d: new year %s leap %d, want %s leap %d", year, mmdd(got[0].Start), Chinese.LeapMonth(year), newYear, leap)
		}
		if mine != bits {
			differing++
			if !printedDifferently[year] {
				t.Errorf("%d: month lengths %s, want %s", year, mine, bits)
			}
		}
	}
	if months < 2300 || differing != len(printedDifferently) {
		t.Fatalf("%d months compared, %d years differ (want %d)", months, differing, len(printedDifferently))
	}
	t.Logf("%d months compared over 1912-2098; %d differ, all explained", months, differing)
}

func TestRegionalClocksAgreeWithPublishedHolidays(t *testing.T) {
	checked := 0
	for _, line := range lines(t, "holidays.txt") {
		fields := strings.Fields(line)
		var year int
		fmt.Sscan(fields[0], &year)
		lunar := func(c *Calendar, month, day int) string {
			jdn, ok := c.ToJDN(Date{Year: year, Month: month, Day: day})
			if !ok {
				t.Fatalf("%d: no such date", year)
			}
			return mmdd(jdn)
		}
		term := func(c *Calendar, lon float64) string {
			jdn, _ := c.TermDay(year, lon)
			return mmdd(jdn)
		}
		got := map[string][]string{
			"kr": {lunar(Korean, 1, 1), lunar(Korean, 8, 15), lunar(Korean, 4, 8)},
			"vn": {lunar(Vietnamese, 1, 1)},
			"jp": {term(Japanese, 0), term(Japanese, 180)},
			"tw": {term(Chinese, 15)},
		}
		for _, f := range fields[1:] {
			key, list, _ := strings.Cut(f, ":")
			for i, want := range strings.Split(list, ",") {
				if want == "-" {
					continue
				}
				checked++
				if got[key][i] != want {
					t.Errorf("%d %s[%d] = %s, want %s", year, key, i, got[key][i], want)
				}
			}
		}
	}
	if checked < 400 {
		t.Fatalf("only %d dates checked", checked)
	}
	t.Logf("%d holiday dates checked", checked)
}

func TestClocksDisagreeWhenTheNewMoonIsNearMidnight(t *testing.T) {
	// 2027: the new moon is at 23:56 on February 6 in Beijing, 00:56 on February 7 in Seoul.
	cn, _ := Chinese.ToJDN(Date{Year: 2027, Month: 1, Day: 1})
	kr, _ := Korean.ToJDN(Date{Year: 2027, Month: 1, Day: 1})
	if mmdd(cn) != "0206" || mmdd(kr) != "0207" {
		t.Fatalf("new year 2027: chinese %s, korean %s", mmdd(cn), mmdd(kr))
	}
	// 1985: Vietnam started the year a month before China (January 21, February 20).
	vn, _ := Vietnamese.ToJDN(Date{Year: 1985, Month: 1, Day: 1})
	cn, _ = Chinese.ToJDN(Date{Year: 1985, Month: 1, Day: 1})
	if mmdd(vn) != "0121" || mmdd(cn) != "0220" {
		t.Fatalf("new year 1985: vietnamese %s, chinese %s", mmdd(vn), mmdd(cn))
	}
}

func TestRoundTripAndLeapMonths(t *testing.T) {
	for jdn := astro.JDN(1950, 1, 1); jdn < astro.JDN(2080, 1, 1); jdn += 17 {
		d := Chinese.FromJDN(jdn)
		back, ok := Chinese.ToJDN(d)
		if !ok || back != jdn {
			t.Fatalf("round trip failed for %s", d)
		}
	}
	for year, leap := range map[int]int{2020: 4, 2023: 2, 2025: 6, 2028: 5, 2033: 11, 2024: 0} {
		if got := Chinese.LeapMonth(year); got != leap {
			t.Errorf("leap month of %d = %d, want %d", year, got, leap)
		}
	}
	if d := Chinese.FromGregorian(2025, 7, 25); d != (Date{Year: 2025, Month: 6, Day: 1, Leap: true}) {
		t.Errorf("2025-07-25 = %s, want the first day of leap month 6", d)
	}
}

func TestBeforeNineteenTwentyNineUsesBeijingTime(t *testing.T) {
	// New moon of 1916-02-04 00:05 UTC+8, which is February 3 at the meridian of Beijing.
	jdn, _ := Chinese.ToJDN(Date{Year: 1916, Month: 1, Day: 1})
	if mmdd(jdn) != "0203" {
		t.Fatalf("new year 1916 = %s, want 0203", mmdd(jdn))
	}
}

func TestUncertainMonths(t *testing.T) {
	// 2057: the new moon that starts month 9 is within seconds of midnight in UTC+8.
	var flagged []string
	for year := 2026; year <= 2098; year++ {
		for _, m := range Chinese.Months(year) {
			if m.Uncertain {
				flagged = append(flagged, fmt.Sprintf("%d-%s", year, mmdd(m.Start)))
			}
		}
	}
	want := "2057-0928 2089-0904 2097-0807"
	if strings.Join(flagged, " ") != want {
		t.Fatalf("uncertain months = %v, want %s", flagged, want)
	}
}

func TestDayCycle(t *testing.T) {
	stem, branch := DayCycle(astro.JDN(2000, 1, 1)) // 戊午
	if Stems[stem]+Branches[branch] != "戊午" {
		t.Fatalf("2000-01-01 = %s%s, want 戊午", Stems[stem], Branches[branch])
	}
}
