// Command koyomi prints commercial calendars for Japan, Taiwan, South Korea and France, and
// the lunisolar calendar they rest on.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/thibaultbordrie-cloud/koyomi/astro"
	"github.com/thibaultbordrie-cloud/koyomi/events"
	ls "github.com/thibaultbordrie-cloud/koyomi/lunisolar"
)

const usage = `koyomi 暦: commercial calendars computed from astronomy.

  koyomi events  [-market jp,tw,kr,fr] [-year 2027 | -from 2026 -to 2028] [-format text|json|csv|ics]
  koyomi lunar   2027-02-06 [-calendar chinese|japanese|korean|vietnamese]
  koyomi compare 2027          lunar new year and mid-autumn on the four clocks
  koyomi terms   2027 [-calendar japanese]
  koyomi close   [-from 2020 -to 2100]   new moons within ten minutes of midnight
  koyomi build   [-out site] [-from 2024 -to 2035]   write the feeds and the page
`

func iso(jdn int) string {
	y, m, d := astro.Civil(jdn)
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "koyomi:", err)
	os.Exit(1)
}

// parse reads the flags that follow the positional arguments.
func parse(fs *flag.FlagSet, args []string) []string {
	var pos []string
	for len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		pos, args = append(pos, args[0]), args[1:]
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	return append(pos, fs.Args()...)
}

func marketList(s string) []string {
	if s == "all" {
		s = "jp,tw,kr,fr"
	}
	return strings.Split(s, ",")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	thisYear := time.Now().Year()
	switch os.Args[1] {
	case "events":
		market := fs.String("market", "all", "markets, comma separated")
		year := fs.Int("year", thisYear, "year")
		from := fs.Int("from", 0, "first year")
		to := fs.Int("to", 0, "last year")
		format := fs.String("format", "text", "text, json, csv or ics")
		parse(fs, os.Args[2:])
		if *from == 0 {
			*from, *to = *year, *year
		}
		occ, err := events.Range(marketList(*market), *from, *to)
		if err != nil {
			fail(err)
		}
		sort.SliceStable(occ, func(i, j int) bool { return occ[i].StartJDN() < occ[j].StartJDN() })
		switch *format {
		case "json":
			err = events.WriteJSON(os.Stdout, occ)
		case "csv":
			err = events.WriteCSV(os.Stdout, occ)
		case "ics":
			err = events.WriteICS(os.Stdout, "Koyomi "+strings.ToUpper(*market), occ)
		default:
			err = events.WriteText(os.Stdout, occ)
		}
		if err != nil {
			fail(err)
		}
	case "lunar":
		name := fs.String("calendar", "chinese", "chinese, japanese, korean or vietnamese")
		pos := parse(fs, os.Args[2:])
		if len(pos) != 1 {
			fail(fmt.Errorf("usage: koyomi lunar YYYY-MM-DD"))
		}
		t, err := time.Parse("2006-01-02", pos[0])
		if err != nil {
			fail(err)
		}
		c, err := ls.ByName(*name)
		if err != nil {
			fail(err)
		}
		if t.Year() <= ls.MinYear || t.Year() >= ls.MaxYear {
			fail(fmt.Errorf("the calendar is computed for %d to %d", ls.MinYear+1, ls.MaxYear-1))
		}
		jdn := astro.JDN(t.Year(), int(t.Month()), t.Day())
		stem, branch := ls.DayCycle(jdn)
		fmt.Printf("%s = %s in the %s calendar, day %s%s\n", pos[0], c.FromJDN(jdn), c.Name, ls.Stems[stem], ls.Branches[branch])
	case "compare":
		pos := parse(fs, os.Args[2:])
		year := thisYear
		if len(pos) == 1 {
			fmt.Sscan(pos[0], &year)
		}
		if year <= ls.MinYear || year >= ls.MaxYear {
			fail(fmt.Errorf("the calendar is computed for %d to %d", ls.MinYear+1, ls.MaxYear-1))
		}
		fmt.Printf("%-11s %-7s %-12s %-12s %s\n", "calendar", "clock", "new year", "mid-autumn", "leap month")
		for _, c := range []*ls.Calendar{ls.Chinese, ls.Japanese, ls.Korean, ls.Vietnamese} {
			ny, _ := c.ToJDN(ls.Date{Year: year, Month: 1, Day: 1})
			ma, _ := c.ToJDN(ls.Date{Year: year, Month: 8, Day: 15})
			leap := "-"
			if l := c.LeapMonth(year); l > 0 {
				leap = fmt.Sprint(l)
			}
			fmt.Printf("%-11s UTC+%-3.0f %-12s %-12s %s\n", c.Name, c.TZ, iso(ny), iso(ma), leap)
		}
	case "terms":
		name := fs.String("calendar", "chinese", "clock used for the local time")
		pos := parse(fs, os.Args[2:])
		year := thisYear
		if len(pos) == 1 {
			fmt.Sscan(pos[0], &year)
		}
		c, err := ls.ByName(*name)
		if err != nil {
			fail(err)
		}
		for i, term := range ls.Terms {
			lon := float64((285 + 15*i) % 360)
			jdn, h, m := c.LocalTime(astro.SolarTerm(year, lon))
			fmt.Printf("%s  %3.0f°  %s %02d:%02d (UTC+%.0f)\n", term, lon, iso(jdn), h, m, c.TZ)
		}
	case "close":
		from := fs.Int("from", 2020, "first year")
		to := fs.Int("to", 2100, "last year")
		parse(fs, os.Args[2:])
		first := astro.Lunation(float64(astro.JDN(*from, 1, 1)))
		last := astro.Lunation(float64(astro.JDN(*to+1, 1, 1)))
		for _, c := range []*ls.Calendar{ls.Chinese, ls.Japanese, ls.Vietnamese} {
			for k := first; k < last; k++ {
				jde := astro.NewMoon(k)
				local := astro.UT(jde) + 0.5 + c.TZ/24
				sec := (local - math.Floor(local)) * 86400
				if sec > 43200 {
					sec -= 86400
				}
				if math.Abs(sec) < 600 {
					mark := ""
					if c.Uncertain(jde) {
						mark = "  UNCERTAIN"
					}
					fmt.Printf("UTC+%.0f  new moon %+4.0f s from midnight, month starts %s%s\n", c.TZ, sec, iso(int(math.Floor(local))), mark)
				}
			}
		}
	case "build":
		out := fs.String("out", "site", "output folder")
		from := fs.Int("from", 2024, "first year")
		to := fs.Int("to", 2035, "last year")
		parse(fs, os.Args[2:])
		if err := build(*out, *from, *to); err != nil {
			fail(err)
		}
	default:
		fmt.Print(usage)
		os.Exit(2)
	}
}

func build(out string, from, to int) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	write := func(name string, fn func(f *os.File) error) error {
		f, err := os.Create(filepath.Join(out, name))
		if err != nil {
			return err
		}
		defer f.Close()
		return fn(f)
	}
	var all []events.Occurrence
	for _, m := range events.Markets {
		occ, err := events.Range([]string{m.Code}, from, to)
		if err != nil {
			return err
		}
		all = append(all, occ...)
		if err := write(m.Code+".ics", func(f *os.File) error { return events.WriteICS(f, "Koyomi · "+m.Name, occ) }); err != nil {
			return err
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].StartJDN() < all[j].StartJDN() })
	steps := []struct {
		name string
		fn   func(f *os.File) error
	}{
		{"all.ics", func(f *os.File) error { return events.WriteICS(f, "Koyomi · all markets", all) }},
		{"events.json", func(f *os.File) error { return events.WriteJSON(f, all) }},
		{"events.csv", func(f *os.File) error { return events.WriteCSV(f, all) }},
		{"index.html", func(f *os.File) error { return events.WriteHTML(f, all, from, to) }},
	}
	for _, s := range steps {
		if err := write(s.name, s.fn); err != nil {
			return err
		}
	}
	fmt.Printf("Wrote %d events for %d to %d in %s\n", len(all), from, to, out)
	return nil
}
