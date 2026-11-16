# Koyomi 暦

[![ci](https://github.com/thibaultbordrie-cloud/koyomi/actions/workflows/ci.yml/badge.svg)](https://github.com/thibaultbordrie-cloud/koyomi/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-no%20dependency-00ADD8?logo=go&logoColor=white)
![Validated](https://img.shields.io/badge/validated-2%2C312%20lunar%20months%20%C2%B7%20459%20holidays-2E7D32)
![License: MIT](https://img.shields.io/badge/License-MIT-lightgrey)

**The commercial calendar of Japan, Taiwan, South Korea and France, computed from astronomy and published as data: JSON, CSV and calendar feeds you can subscribe to.**

*Koyomi* means "calendar". When is 春節 in 2031? Which day do Japanese supermarkets sell eel next summer? Why is Korean New Year 2027 one day after the Chinese one? Koyomi answers from first principles: it computes the new moons and the position of the Sun, builds the lunisolar calendar on each country's clock, and derives the dates that drive CRM and retail plans.

[日本語](#日本語での概要) · [繁體中文](#中文簡介繁體) · [Français](#en-bref-français)

## Use the data

No install needed. The feeds are rebuilt by CI and served by GitHub Pages:

| Feed | URL |
|---|---|
| Page | https://thibaultbordrie-cloud.github.io/koyomi/ |
| Calendar feeds (iCalendar) | [`jp.ics`](https://thibaultbordrie-cloud.github.io/koyomi/jp.ics) · [`tw.ics`](https://thibaultbordrie-cloud.github.io/koyomi/tw.ics) · [`kr.ics`](https://thibaultbordrie-cloud.github.io/koyomi/kr.ics) · [`fr.ics`](https://thibaultbordrie-cloud.github.io/koyomi/fr.ics) · [`all.ics`](https://thibaultbordrie-cloud.github.io/koyomi/all.ics) |
| Data | [`events.json`](https://thibaultbordrie-cloud.github.io/koyomi/events.json) · [`events.csv`](https://thibaultbordrie-cloud.github.io/koyomi/events.csv) |

In Google Calendar or Outlook, "add calendar from URL" with one of the `.ics` links. The feeds cover 2024 to 2035: about 73 events per year, 883 in total.

## Or run it

```bash
go install github.com/thibaultbordrie-cloud/koyomi/cmd/koyomi@latest   # or download a binary from Releases
koyomi events -market tw -year 2027
koyomi compare 2027
```

```text
tw  2027-01-23 → 02-05 lunisolar  年貨採買 · New Year shopping
tw  2027-01-23         lunisolar  尾牙 · Year-end banquet
tw  2027-02-05         lunisolar  除夕 · New Year's Eve
tw  2027-02-06 → 02-08 lunisolar  春節 · Lunar New Year
tw  2027-02-20         lunisolar  元宵節 · Lantern Festival
tw  2027-04-05         solar_term 清明節 · Tomb-Sweeping Day
…

calendar    clock   new year     mid-autumn   leap month
chinese     UTC+8   2027-02-06   2027-09-15   -
japanese    UTC+9   2027-02-07   2027-09-15   -
korean      UTC+9   2027-02-07   2027-09-15   -
vietnamese  UTC+7   2027-02-06   2027-09-15   -
```

In 2027 the new moon falls at 23:56 on February 6 in Taipei and at 00:56 on February 7 in Seoul. Same Moon, different day: a campaign planned on "Chinese New Year" for Korea would be a day early. The same happens in 2028, and for Vietnam in 2030.

Other commands: `koyomi lunar 2027-02-06` (a date in the lunisolar calendar), `koyomi terms 2027` (the 24 solar terms with their local time), `koyomi close` (new moons within ten minutes of midnight), `koyomi build` (the feeds).

## What is computed

| Basis | How | Examples |
|---|---|---|
| `lunisolar` | new moons + solar terms → months and leap months, on each country's clock | 春節, 端午節, 中秋節, 尾牙, 설날, 추석, 十五夜 |
| `solar_term` | day the Sun reaches a longitude | 清明節, 冬至, 春分の日 and 秋分の日 (Japanese public holidays set by the equinox), 節分 (the day before 立春: February 2 in 2021 and 2025, not the 3rd) |
| `sexagenary` | solar terms + the 60-day cycle of stems and branches | 土用の丑の日 (eel day in Japan: one or two per summer), 초복 · 중복 · 말복 (samgyetang days in Korea) |
| `easter` | the Gregorian computus | Pâques; fête des mères moves to June when Pentecost takes the last Sunday of May |
| `gregorian` | fixed days and "n-th weekday" rules | 母の日, 雙11, 빼빼로데이, soldes d'hiver and d'été with their legal exceptions, Black Friday |

Each event carries its names (local, English, French), the rule in plain words, what it sells, and two honest flags:

- `window: true` when the dates are a trade habit, not a fixed day (お中元, 百貨週年慶, la rentrée);
- `uncertain: true` when the date hangs on a new moon or a solar term a few seconds from midnight. In 2057 a new moon falls about 2 seconds before midnight in UTC+8: no model can say today on which day that month will start, because the slowing of the Earth's rotation is not predictable that far.

## How it is validated

Nothing is copied from a table; the tables are used to check the result. Details in [docs/VALIDATION.md](docs/VALIDATION.md).

| What | Compared with | Result |
|---|---|---|
| New moon instants, 1900 to 2100 | PyEphem | 2,488 new moons, worst difference 16 s |
| Sun's longitude at the 24 solar terms, 1900 to 2100 | PyEphem | 4,824 instants, worst difference 1.0″ (24 s of time) |
| Chinese months and leap months, 1912 to 2098 | `lunardate` | 2,312 months, 3 differences, all explained: the printed calendars of 1933, 1954 and 1978 were a day off |
| Seollal, Chuseok, Buddha's Birthday, Tết, Japanese equinox days, 清明, 1990 to 2060 | `holidays` | 459 dates, 0 difference |
| Eel days, samgyetang days, 節分, 十五夜, 春節, soldes, fête des mères… | known published dates | 38 dates, 0 difference |

The engine has no dependency and no data file: about 200 lines of astronomy (Jean Meeus, *Astronomical Algorithms*: new moons from chapter 49, the Sun from a truncated VSOP87 series), and the rules of the Chinese standard GB/T 33661-2017 for months and leap months. Until 1928 the Chinese calendar is computed for the meridian of Beijing, as it was at the time.

## Limits

- **Not a public-holiday library.** Some events are public holidays and say so, but substitute days and one-off decrees are out of scope: use [`holidays`](https://github.com/vacanza/holidays) for that.
- **Trade habits are habits.** Sales seasons and gift seasons vary by region and by retailer; the `rule` field says what was assumed.
- **Range.** 1913 to 2099. Before 1912 the Chinese calendar followed the tables of the Qing court, which astronomy does not reproduce.
- **Japan has no official lunisolar calendar today.** Koyomi applies the same rules on Japanese time, the usual convention for 旧暦.
- **French sales** are the national dates; some border and overseas départements have their own.

## Develop

```bash
go test ./...          # engine, calendars, events, feeds
go run ./cmd/koyomi build -out site
python3 tools/reference.py   # regenerates testdata/ from PyEphem, lunardate and holidays
```

## 日本語での概要

Koyomi（暦）は、日本・台湾・韓国・フランスの「商戦カレンダー」を天文計算から作るツールです。新月と太陽黄経を計算し、国ごとの標準時で旧暦を組み立てます。春節、中秋節、설날、十五夜、節分、春分の日、土用の丑の日、韓国の三伏などの日付を求め、JSON・CSV・購読できる iCalendar で公開します。2027年の旧正月は、台湾では2月6日、韓国では2月7日です。新月が真夜中の直前か直後かで日付が変わります。結果は PyEphem や公開されている祝日データと照合しています。Go 製で、外部ライブラリは使っていません。

## 中文簡介（繁體）

Koyomi（暦）用天文計算產生日本、台灣、韓國與法國的「行銷行事曆」。它計算朔日與太陽黃經，依各國的標準時間排出農曆，再推算春節、端午、中秋、尾牙、清明、冬至、韓國的三伏、日本的土用丑日等日期，並以 JSON、CSV 與可訂閱的 iCalendar 發布。2027 年的農曆新年，台灣是 2 月 6 日，韓國是 2 月 7 日：同一個朔，落在午夜的前後，日期就不同。結果已與 PyEphem 及公開的假日資料比對。以 Go 撰寫，沒有任何外部相依套件。

## En bref (français)

Koyomi calcule le calendrier commercial du Japon, de Taïwan, de la Corée du Sud et de la France à partir de l'astronomie : nouvelles lunes, position du Soleil, puis calendrier luni-solaire à l'heure de chaque pays. Il en tire les dates qui comptent pour un plan CRM (Nouvel An lunaire, mi-automne, jours de l'anguille au Japon, soldes et fête des mères en France) et les publie en JSON, CSV et flux iCalendar. Les résultats sont comparés à PyEphem et à des tables publiées. Écrit en Go, sans aucune dépendance.

## License

MIT.
