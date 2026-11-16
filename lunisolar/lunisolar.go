// Package lunisolar builds the East Asian lunisolar calendar from astronomy, with the modern
// rules (the ones of the Chinese standard GB/T 33661-2017):
//
//  1. a day starts at midnight on the reference clock;
//  2. a month starts on the day of the new moon;
//  3. the month that holds the winter solstice is month 11;
//  4. when 13 new moons fall between two month 11s, the first month without a principal
//     solar term (a multiple of 30 degrees) is a leap month and repeats the previous number.
//
// The reference clock is the only thing that changes between countries: UTC+8 for China and
// Taiwan, UTC+9 for Korea and for the Japanese kyureki, UTC+7 for Vietnam. A new moon or a term
// close to midnight then falls on different days, and now and then a festival moves by a day,
// or by a month.
package lunisolar

import (
	"fmt"
	"math"
	"sync"

	"github.com/thibaultbordrie-cloud/koyomi/astro"
)

// Calendar is a lunisolar calendar on one reference clock.
type Calendar struct {
	Name string
	TZ   float64 // hours ahead of UT
	// Earlier clocks, oldest first: TZ applies from the last Before onwards.
	Earlier []Clock

	mu    sync.Mutex
	years map[int][]Month
}

// Clock is a reference clock that was in use before a given day.
type Clock struct {
	Before int     // Julian Day Number
	TZ     float64 // hours ahead of UT
}

// The four clocks in use today. Until the end of 1928 the Chinese calendar was computed for
// the meridian of Beijing (116 deg 23 min east, 7 h 45 min 40 s ahead of UT), not for UTC+8.
var (
	Chinese = &Calendar{Name: "chinese", TZ: 8, // China, Taiwan, Hong Kong
		Earlier: []Clock{{Before: 2425613, TZ: 7 + 45.0/60 + 40.0/3600}}}
	Japanese   = &Calendar{Name: "japanese", TZ: 9}
	Korean     = &Calendar{Name: "korean", TZ: 9}
	Vietnamese = &Calendar{Name: "vietnamese", TZ: 7}
)

// Month is one lunar month.
type Month struct {
	Year   int  // lunar year: the Gregorian year in which its month 1 starts
	Number int  // 1 to 12
	Leap   bool // true for a leap month
	Start  int  // Julian Day Number of its first day
	Days   int  // 29 or 30
	// Uncertain is true when the new moon falls so close to midnight that the first day
	// cannot be relied on (see Uncertain).
	Uncertain bool
}

// Date is a date in the lunisolar calendar.
type Date struct {
	Year, Month, Day int
	Leap             bool
}

func (d Date) String() string {
	leap := ""
	if d.Leap {
		leap = "leap "
	}
	return fmt.Sprintf("%d-%s%02d-%02d", d.Year, leap, d.Month, d.Day)
}

// MinYear and MaxYear bound the Gregorian years the calendar is computed for.
const (
	MinYear = 1912
	MaxYear = 2100
)

// Uncertain tells whether an instant (TT) is too close to local midnight to say on which day
// it falls. The margin is 30 seconds for the model itself, plus 2 seconds for each year after
// 2025: the slowing of the Earth's rotation (Delta T) cannot be predicted better than that.
func (c *Calendar) Uncertain(jde float64) bool {
	local := astro.UT(jde) + 0.5 + c.clock(jde)/24
	seconds := (local - math.Floor(local)) * 86400
	if seconds > 43200 {
		seconds = 86400 - seconds
	}
	margin := 30.0
	if year := 2000 + (jde-2451545)/365.25; year > 2025 {
		margin += 2 * (year - 2025)
	}
	return seconds < margin
}

// clock returns the offset in use at an instant.
func (c *Calendar) clock(jd float64) float64 {
	for _, e := range c.Earlier {
		if jd < float64(e.Before) {
			return e.TZ
		}
	}
	return c.TZ
}

func (c *Calendar) localDay(jde float64) int { return astro.LocalDay(jde, c.clock(jde)) }

func (c *Calendar) newMoonDay(k int) int { return c.localDay(astro.NewMoon(k)) }

// termIndex returns which 30-degree sector the Sun is in when the given local day starts.
func (c *Calendar) termIndex(day int) int {
	jdUT := float64(day) - 0.5 - c.clock(float64(day))/24
	year := 2000 + (jdUT-2451545)/365.25
	return int(math.Floor(astro.SunLongitude(jdUT+astro.DeltaT(year)/86400) / 30))
}

// month11 returns the lunation that holds the winter solstice of a Gregorian year.
func (c *Calendar) month11(year int) int {
	solstice := astro.SolarTerm(year, 270)
	day := c.localDay(solstice)
	k := astro.Lunation(solstice)
	for c.newMoonDay(k) > day {
		k--
	}
	for c.newMoonDay(k+1) <= day {
		k++
	}
	return k
}

// sui returns the months from month 11 of year-1 up to, not including, month 11 of year.
func (c *Calendar) sui(year int) []Month {
	c.mu.Lock()
	defer c.mu.Unlock()
	if m, ok := c.years[year]; ok {
		return m
	}
	first, last := c.month11(year-1), c.month11(year)
	leapAt, hasLeap := 0, false
	if last-first == 13 {
		for k := first + 1; k < last; k++ {
			if c.termIndex(c.newMoonDay(k)) == c.termIndex(c.newMoonDay(k+1)) {
				leapAt, hasLeap = k, true
				break
			}
		}
	}
	months := make([]Month, 0, 13)
	number, lunarYear := 11, year-1
	for k := first; k < last; k++ {
		leap := hasLeap && k == leapAt
		if k > first && !leap {
			number++
			if number > 12 {
				number, lunarYear = 1, year
			}
		}
		start := c.newMoonDay(k)
		months = append(months, Month{Year: lunarYear, Number: number, Leap: leap, Start: start,
			Days: c.newMoonDay(k+1) - start, Uncertain: c.Uncertain(astro.NewMoon(k))})
	}
	if c.years == nil {
		c.years = map[int][]Month{}
	}
	c.years[year] = months
	return months
}

// Months returns the months of a lunar year, in order (12 or 13 of them).
func (c *Calendar) Months(lunarYear int) []Month {
	var out []Month
	for _, y := range []int{lunarYear, lunarYear + 1} {
		for _, m := range c.sui(y) {
			if m.Year == lunarYear {
				out = append(out, m)
			}
		}
	}
	return out
}

// FromJDN converts a Julian Day Number to a lunisolar date.
func (c *Calendar) FromJDN(jdn int) Date {
	year, _, _ := astro.Civil(jdn)
	for _, y := range []int{year + 1, year, year - 1} {
		months := c.sui(y)
		if jdn < months[0].Start {
			continue
		}
		for i := len(months) - 1; i >= 0; i-- {
			if m := months[i]; jdn >= m.Start {
				return Date{Year: m.Year, Month: m.Number, Day: jdn - m.Start + 1, Leap: m.Leap}
			}
		}
	}
	panic("lunisolar: unreachable")
}

// FromGregorian converts a Gregorian date to a lunisolar date.
func (c *Calendar) FromGregorian(year, month, day int) Date {
	return c.FromJDN(astro.JDN(year, month, day))
}

// ToJDN returns the Julian Day Number of a lunisolar date, and false when that date does not
// exist (no such leap month, or day 30 of a 29-day month).
func (c *Calendar) ToJDN(d Date) (int, bool) {
	for _, m := range c.Months(d.Year) {
		if m.Number == d.Month && m.Leap == d.Leap {
			if d.Day < 1 || d.Day > m.Days {
				return 0, false
			}
			return m.Start + d.Day - 1, true
		}
	}
	return 0, false
}

// LeapMonth returns the number of the leap month of a lunar year, or 0.
func (c *Calendar) LeapMonth(lunarYear int) int {
	for _, m := range c.Months(lunarYear) {
		if m.Leap {
			return m.Number
		}
	}
	return 0
}

// The ten stems and twelve branches that name days (and years) in a 60-day cycle.
var (
	Stems    = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	Branches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
)

// DayCycle returns the stem (0 to 9) and branch (0 to 11) of a day in the sexagenary cycle.
func DayCycle(jdn int) (stem, branch int) {
	n := (jdn + 49) % 60
	return n % 10, n % 12
}

// TermDay returns the local day on which the Sun reaches a longitude during a Gregorian year,
// and whether that day is uncertain.
func (c *Calendar) TermDay(year int, longitude float64) (day int, uncertain bool) {
	jde := astro.SolarTerm(year, longitude)
	return c.localDay(jde), c.Uncertain(jde)
}

// Terms names the 24 solar terms in the order of a Gregorian year: the first one, 小寒, is the
// Sun at 285 degrees (early January), and each next one is 15 degrees further.
var Terms = []string{"小寒", "大寒", "立春", "雨水", "驚蟄", "春分", "清明", "穀雨", "立夏", "小滿", "芒種", "夏至",
	"小暑", "大暑", "立秋", "處暑", "白露", "秋分", "寒露", "霜降", "立冬", "小雪", "大雪", "冬至"}

// ByName returns one of the four calendars.
func ByName(name string) (*Calendar, error) {
	for _, c := range []*Calendar{Chinese, Japanese, Korean, Vietnamese} {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("unknown calendar %q (known: chinese, japanese, korean, vietnamese)", name)
}

// LocalTime returns the local date and the time of day (hours, minutes) of an instant (TT).
func (c *Calendar) LocalTime(jde float64) (jdn, hour, minute int) {
	local := astro.UT(jde) + 0.5 + c.clock(jde)/24
	jdn = int(math.Floor(local))
	minutes := int((local - math.Floor(local)) * 1440)
	return jdn, minutes / 60, minutes % 60
}
