# Validation

Koyomi computes its dates; reference tables are only used to check them. The checks below run in `go test ./...` from the files in `testdata/`, which `tools/reference.py` regenerates from three independent Python libraries.

## The astronomy

| Quantity | Reference | Sample | Worst difference |
|---|---|---|---|
| Instant of new moon | PyEphem 4.2.1 | all 2,488 new moons of 1900 to 2100 | 16 s (mean 3.5 s) |
| Apparent longitude of the Sun | PyEphem 4.2.1 | the 24 solar terms of each year, 4,824 instants | 1.0″, which is 24 s of time (mean 0.25″) |

Both are compared in Terrestrial Time, so the result does not depend on Delta T. The test suite keeps a sample: 81 new moons and 108 equinoxes or solstices, with a limit of 60 s.

Two examples from Jean Meeus, *Astronomical Algorithms*, are tests too: the new moon of 1977 February 18 (example 49.a) and the Sun's longitude on 1992 October 13 (example 25.b).

## The calendars

| Calendar | Reference | Sample | Differences |
|---|---|---|---|
| Chinese (UTC+8; Beijing meridian until 1928) | `lunardate` 0.3.0 | new year, leap month and length of 2,312 months, 1912 to 2098 | 3 |
| Korean (UTC+9) | `holidays` | Seollal, Chuseok, Buddha's Birthday, 1990 to 2060 | 0 |
| Vietnamese (UTC+7) | `holidays` | Tết, 1990 to 2060 | 0 |
| Japanese clock (UTC+9) | `holidays` | vernal and autumnal equinox days, 1990 to 2060 | 0 |
| Chinese clock | `holidays` | 清明 (Tomb-Sweeping Day) where listed | 0 |

The three differences with `lunardate` are months whose first day was printed a day off in the calendars of the time:

| Month start | New moon (UTC+8) | Koyomi | Table |
|---|---|---|---|
| lunar month 6 of 1933 | July 23, 00:03 | July 23 | July 22 |
| lunar month 11 of 1954 | November 25, 20:30 | November 25 | November 26 |
| lunar month 8 of 1978 | September 3, 00:09 | September 3 | September 2 |

Two other libraries, `lunar_python` and `LunarCalendar`, agree with Koyomi on these three. During development they were compared over 1902 to 2099 as well: the only other differences were 1906 (before 1912, see below) and the uncertain months listed next.

## What cannot be known

A month starts on the day of the new moon. When the new moon is a few seconds from midnight, the day depends on Delta T, the slow and irregular drift between atomic time and the Earth's rotation. It is measured for the past and only estimated for the future.

Koyomi marks an instant as uncertain when it is closer to midnight than 30 seconds, plus 2 seconds for each year after 2025. Between 2026 and 2098, on the Chinese clock, three months are flagged:

| Month start | Distance to midnight (this model) |
|---|---|
| 2057-09-28 | 2 s before |
| 2089-09-04 | 144 s before |
| 2097-08-07 | 20 s before |

The libraries used for comparison disagree with each other on exactly these three months. Events that depend on such a month carry `uncertain: true`.

## Known limits

- Delta T follows the polynomials of Espenak and Meeus (2006). For 2026 they give 75 s where about 69 s is observed: a 6 s error, far below the margins above.
- Before 1912 the Chinese calendar was computed by the Qing court with its own tables. Koyomi starts in 1912.
- Korea used UTC+8:30 from 1954 to 1961 and Vietnam changed clocks before 1975. The Korean and Vietnamese calendars are checked from 1990 only.
- The 38 "known dates" of `events/events_test.go` come from published calendars of each country.
