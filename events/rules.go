// Package events turns the calendar engine into dated commercial moments for each market.
package events

import (
	"github.com/thibaultbordrie-cloud/koyomi/astro"
	ls "github.com/thibaultbordrie-cloud/koyomi/lunisolar"
)

// span is a run of days (Julian Day Numbers, both ends included).
type span struct {
	start, end int
	uncertain  bool
}

// rule returns the occurrences that a rule produces around a Gregorian year. It may return
// dates of the neighbouring years: the caller keeps those that start in the year asked.
type rule func(year int) []span

func one(jdn int) []span { return []span{{start: jdn, end: jdn}} }

// weekday returns 0 for Sunday to 6 for Saturday.
func weekday(jdn int) int { return (jdn + 1) % 7 }

// fixed is the same day every year; days > 1 makes it a run of days.
func fixed(month, day, days int) rule {
	return func(y int) []span {
		start := astro.JDN(y, month, day)
		return []span{{start: start, end: start + days - 1}}
	}
}

// nthWeekday returns the n-th given weekday of a month (n = 1 is the first).
func nthWeekday(y, month, wd, n int) int {
	first := astro.JDN(y, month, 1)
	return first + (wd-weekday(first)+7)%7 + 7*(n-1)
}

// lastWeekday returns the last given weekday of a month.
func lastWeekday(y, month, wd int) int {
	next := astro.JDN(y, month+1, 1) // JDN accepts month 13
	last := next - 1
	return last - (weekday(last)-wd+7)%7
}

func nth(month, wd, n int) rule {
	return func(y int) []span { return one(nthWeekday(y, month, wd, n)) }
}

// lunar is a day of the lunisolar calendar, widened by `before` and `after` days.
func lunar(c *ls.Calendar, month, day, before, after int) rule {
	return func(y int) []span {
		var out []span
		for _, ly := range []int{y - 1, y} {
			if ly < ls.MinYear || ly > ls.MaxYear-1 {
				continue
			}
			for _, m := range c.Months(ly) {
				if m.Number == month && !m.Leap && day <= m.Days {
					d := m.Start + day - 1
					out = append(out, span{start: d - before, end: d + after, uncertain: m.Uncertain})
				}
			}
		}
		return out
	}
}

// beforeNewYear is a run of days that ends the day before lunar new year (New Year's Eve).
func beforeNewYear(c *ls.Calendar, days int) rule {
	return func(y int) []span {
		if y < ls.MinYear || y > ls.MaxYear-1 {
			return nil
		}
		m := c.Months(y)[0]
		return []span{{start: m.Start - days, end: m.Start - 1, uncertain: m.Uncertain}}
	}
}

// term is the day the Sun reaches a longitude, shifted by offset days.
func term(c *ls.Calendar, longitude float64, offset int) rule {
	return func(y int) []span {
		d, unsure := c.TermDay(y, longitude)
		return []span{{start: d + offset, end: d + offset, uncertain: unsure}}
	}
}

// doyoUshi returns the Ox days (丑) of the summer doyo: from the day the Sun reaches 117
// degrees to the day before risshu (135 degrees), on the Japanese clock.
func doyoUshi(y int) []span {
	from, u1 := ls.Japanese.TermDay(y, 117)
	risshu, u2 := ls.Japanese.TermDay(y, 135)
	var out []span
	for d := from; d < risshu; d++ {
		if _, branch := ls.DayCycle(d); branch == 1 {
			out = append(out, span{start: d, end: d, uncertain: (u1 && d == from) || (u2 && d == risshu-1)})
		}
	}
	return out
}

// geng returns the n-th geng day (庚) counted from a day, that day included.
func geng(from, n int) int {
	for d := from; ; d++ {
		if stem, _ := ls.DayCycle(d); stem == 6 {
			if n--; n == 0 {
				return d
			}
		}
	}
}

// sambok returns one of the three "dog days" of Korea: the third and fourth geng days from
// the summer solstice (chobok, jungbok) and the first geng day from ipchu (malbok).
func sambok(which int) rule {
	return func(y int) []span {
		solstice, _ := ls.Korean.TermDay(y, 90)
		ipchu, _ := ls.Korean.TermDay(y, 135)
		switch which {
		case 0:
			return one(geng(solstice, 3))
		case 1:
			return one(geng(solstice, 4))
		default:
			return one(geng(ipchu, 1))
		}
	}
}

// Easter returns the Julian Day Number of Easter Sunday (Gregorian computus).
func Easter(y int) int {
	a, b, c := y%19, y/100, y%100
	d, e := b/4, b%4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i, k := c/4, c%4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return astro.JDN(y, month, day)
}

// feteDesMeres is the last Sunday of May, moved to the first Sunday of June when Pentecost
// falls on that day.
func feteDesMeres(y int) []span {
	d := lastWeekday(y, 5, 0)
	if d == Easter(y)+49 {
		d = nthWeekday(y, 6, 0, 1)
	}
	return one(d)
}

// soldes returns the four weeks of the French winter (month 1) or summer (month 6) sales.
func soldes(month int) rule {
	return func(y int) []span {
		var start int
		if month == 1 {
			start = nthWeekday(y, 1, 3, 2)
			if _, _, day := astro.Civil(start); day > 12 {
				start -= 7
			}
		} else {
			start = lastWeekday(y, 6, 3)
			if _, _, day := astro.Civil(start); day > 28 {
				start -= 7
			}
		}
		return []span{{start: start, end: start + 27}}
	}
}

// blackFriday runs from the day after the fourth Thursday of November to Cyber Monday.
func blackFriday(y int) []span {
	friday := nthWeekday(y, 11, 4, 4) + 1
	return []span{{start: friday, end: friday + 3}}
}

// epiphanie is the first Sunday after January 1.
func epiphanie(y int) []span {
	d := astro.JDN(y, 1, 2)
	return one(d + (7-weekday(d))%7)
}
