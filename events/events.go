package events

import (
	"fmt"
	"sort"

	"github.com/thibaultbordrie-cloud/koyomi/astro"
	ls "github.com/thibaultbordrie-cloud/koyomi/lunisolar"
)

// Market is a country with its reference clock.
type Market struct {
	Code, Name, Clock string
}

// Markets lists the markets, in output order.
var Markets = []Market{
	{"jp", "Japan", "UTC+9"}, {"tw", "Taiwan", "UTC+8"}, {"kr", "South Korea", "UTC+9"}, {"fr", "France", "civil calendar"},
}

// Occurrence is one dated event.
type Occurrence struct {
	ID        string `json:"id"`
	Market    string `json:"market"`
	Start     string `json:"start"`
	End       string `json:"end"` // last day, included
	NameLocal string `json:"name_local"`
	NameEN    string `json:"name_en"`
	NameFR    string `json:"name_fr"`
	Kind      string `json:"kind"`  // commercial, festival or public_holiday
	Basis     string `json:"basis"` // gregorian, lunisolar, solar_term, sexagenary or easter
	Rule      string `json:"rule"`
	Sells     string `json:"sells,omitempty"`
	// Window is true when the dates are a trade habit, not a fixed day.
	Window bool `json:"window"`
	// Uncertain is true when a new moon or a solar term falls too close to midnight.
	Uncertain bool `json:"uncertain"`

	start, end int
}

// StartJDN and EndJDN give the Julian Day Numbers of the first and last day.
func (o Occurrence) StartJDN() int { return o.start }
func (o Occurrence) EndJDN() int   { return o.end }

type def struct {
	id, local, en, fr, kind, basis, rule, sells string
	window                                      bool
	when                                        rule
}

const (
	com, fest, hol          = "commercial", "festival", "public_holiday"
	greg, luni, solar, sexa = "gregorian", "lunisolar", "solar_term", "sexagenary"
	sun, mon                = 0, 1
	blackFridayRule         = "from the day after the fourth Thursday of November to the following Monday"
)

var catalogue = map[string][]def{
	"jp": {
		{"hatsuuri", "初売り・福袋", "First sales of the year", "Premières ventes de l'année", com, greg, "January 1 to 3", "lucky bags (福袋), New Year sales", false, fixed(1, 1, 3)},
		{"seijin", "成人の日", "Coming of Age Day", "Jour de la majorité", hol, greg, "second Monday of January", "kimono rental, photo studios", false, nth(1, mon, 2)},
		{"setsubun", "節分", "Setsubun", "Setsubun", fest, solar, "day before risshun (Sun at 315 degrees) (UTC+9 clock)", "ehomaki sushi rolls, roasted soybeans", false, term(ls.Japanese, 315, -1)},
		{"valentine", "バレンタインデー", "Valentine's Day", "Saint-Valentin", com, greg, "February 14", "chocolate", false, fixed(2, 14, 1)},
		{"hinamatsuri", "ひな祭り", "Doll Festival", "Fête des poupées", fest, greg, "March 3", "hina dolls, chirashi-zushi, hina-arare", false, fixed(3, 3, 1)},
		{"whiteday", "ホワイトデー", "White Day", "White Day", com, greg, "March 14", "return gifts for Valentine's Day: sweets, accessories", false, fixed(3, 14, 1)},
		{"shunbun", "春分の日", "Vernal Equinox Day", "Équinoxe de printemps", hol, solar, "day the Sun reaches 0 degrees (UTC+9 clock)", "ohagi, flowers for family graves (higan)", false, term(ls.Japanese, 0, 0)},
		{"golden-week", "ゴールデンウィーク", "Golden Week", "Golden Week", com, greg, "April 29 to May 5", "travel, leisure, outdoor goods", false, fixed(4, 29, 7)},
		{"hahanohi", "母の日", "Mother's Day", "Fête des mères", com, greg, "second Sunday of May", "carnations, gifts", false, nth(5, sun, 2)},
		{"chichinohi", "父の日", "Father's Day", "Fête des pères", com, greg, "third Sunday of June", "drinks, clothing, gifts", false, nth(6, sun, 3)},
		{"ochugen", "お中元", "Mid-year gift season", "Saison des cadeaux d'été", com, greg, "July 1 to 15 in Kanto; July 15 to August 15 in Kansai", "gift sets: beer, noodles, sweets, fruit", true, fixed(7, 1, 15)},
		{"tanabata", "七夕", "Tanabata", "Tanabata", fest, greg, "July 7 (some regions keep August 7)", "decorations, somen noodles", false, fixed(7, 7, 1)},
		{"doyo-ushi", "土用の丑の日", "Day of the Ox in midsummer", "Jour du Bœuf de la canicule", com, sexa, "Ox days (丑) between the day the Sun reaches 117 degrees and the day before risshu (135 degrees)", "grilled eel (うなぎ)", false, doyoUshi},
		{"obon", "お盆", "Obon", "Obon", fest, greg, "August 13 to 16", "travel home, offerings, gifts", false, fixed(8, 13, 4)},
		{"keiro", "敬老の日", "Respect for the Aged Day", "Jour du respect des aînés", hol, greg, "third Monday of September", "gifts for grandparents", false, nth(9, mon, 3)},
		{"jugoya", "十五夜（中秋の名月）", "Harvest moon", "Lune des moissons", fest, luni, "day 15 of lunar month 8 (UTC+9 clock)", "tsukimi dango, seasonal 'tsukimi' menus", false, lunar(ls.Japanese, 8, 15, 0, 0)},
		{"shubun", "秋分の日", "Autumnal Equinox Day", "Équinoxe d'automne", hol, solar, "day the Sun reaches 180 degrees (UTC+9 clock)", "ohagi, flowers for family graves (higan)", false, term(ls.Japanese, 180, 0)},
		{"halloween", "ハロウィン", "Halloween", "Halloween", com, greg, "October 31", "costumes, sweets", false, fixed(10, 31, 1)},
		{"shichigosan", "七五三", "Shichi-Go-San", "Shichi-Go-San", fest, greg, "November 15", "kimono rental, photo studios, chitose-ame", false, fixed(11, 15, 1)},
		{"black-friday", "ブラックフライデー", "Black Friday to Cyber Monday", "Black Friday au Cyber Monday", com, greg, blackFridayRule, "online and store promotions", false, blackFriday},
		{"oseibo", "お歳暮", "Year-end gift season", "Saison des cadeaux de fin d'année", com, greg, "December 1 to 20 (from late November in Kanto)", "gift sets: ham, beer, sweets", true, fixed(12, 1, 20)},
		{"toji", "冬至", "Winter solstice", "Solstice d'hiver", fest, solar, "day the Sun reaches 270 degrees (UTC+9 clock)", "yuzu baths, kabocha squash", false, term(ls.Japanese, 270, 0)},
		{"christmas", "クリスマス", "Christmas", "Noël", com, greg, "December 24 and 25", "cake, fried chicken, gifts", false, fixed(12, 24, 2)},
	},
	"tw": {
		{"nianhuo", "年貨採買", "New Year shopping", "Achats du Nouvel An", com, luni, "the two weeks before New Year's Eve", "New Year food, snacks, gifts, cleaning products", true, beforeNewYear(ls.Chinese, 14)},
		{"chuxi", "除夕", "New Year's Eve", "Réveillon du Nouvel An lunaire", hol, luni, "day before day 1 of lunar month 1", "reunion dinner", false, beforeNewYear(ls.Chinese, 1)},
		{"chunjie", "春節", "Lunar New Year", "Nouvel An lunaire", hol, luni, "days 1 to 3 of lunar month 1 (UTC+8 clock)", "red envelopes, travel, gifts", false, lunar(ls.Chinese, 1, 1, 0, 2)},
		{"yuanxiao", "元宵節", "Lantern Festival", "Fête des lanternes", fest, luni, "day 15 of lunar month 1", "tangyuan, lanterns", false, lunar(ls.Chinese, 1, 15, 0, 0)},
		{"valentine", "西洋情人節", "Valentine's Day", "Saint-Valentin", com, greg, "February 14", "flowers, chocolate, dinners", false, fixed(2, 14, 1)},
		{"qingming", "清明節", "Tomb-Sweeping Day", "Fête de Qingming", hol, solar, "day the Sun reaches 15 degrees (UTC+8 clock)", "run bing (潤餅), offerings, flowers", false, term(ls.Chinese, 15, 0)},
		{"muqinjie", "母親節", "Mother's Day", "Fête des mères", com, greg, "second Sunday of May", "cakes, restaurants, appliances", false, nth(5, sun, 2)},
		{"520", "520", "520 (\"I love you\")", "520 (« je t'aime »)", com, greg, "May 20: 五二零 sounds like 我愛你", "couple gifts, flowers", false, fixed(5, 20, 1)},
		{"duanwu", "端午節", "Dragon Boat Festival", "Fête des bateaux-dragons", hol, luni, "day 5 of lunar month 5", "zongzi (粽子)", false, lunar(ls.Chinese, 5, 5, 0, 0)},
		{"618", "618年中慶", "618 mid-year sales", "Promotions de mi-année 618", com, greg, "June 18", "e-commerce promotions", false, fixed(6, 18, 1)},
		{"fuqinjie", "父親節", "Father's Day", "Fête des pères", com, greg, "August 8: 八八 sounds like 爸爸", "gifts, restaurants", false, fixed(8, 8, 1)},
		{"qixi", "七夕情人節", "Qixi", "Qixi", com, luni, "day 7 of lunar month 7", "flowers, chocolate, dinners", false, lunar(ls.Chinese, 7, 7, 0, 0)},
		{"zhongyuan", "中元節", "Ghost Festival", "Fête des fantômes", fest, luni, "day 15 of lunar month 7", "offerings (中元普渡): snacks, drinks, instant noodles", false, lunar(ls.Chinese, 7, 15, 0, 0)},
		{"zhongqiu", "中秋節", "Mid-Autumn Festival", "Fête de la mi-automne", hol, luni, "day 15 of lunar month 8", "mooncakes, pomelos, barbecue", false, lunar(ls.Chinese, 8, 15, 0, 0)},
		{"zhounianqing", "百貨週年慶", "Department store anniversary sales", "Soldes anniversaire des grands magasins", com, greg, "about October 1 to November 15; each store sets its own dates", "cosmetics, appliances, fashion", true, fixed(10, 1, 46)},
		{"double11", "雙11", "Double 11", "Double 11", com, greg, "November 11", "e-commerce promotions", false, fixed(11, 11, 1)},
		{"black-friday", "黑色星期五", "Black Friday to Cyber Monday", "Black Friday au Cyber Monday", com, greg, blackFridayRule, "online and store promotions", false, blackFriday},
		{"double12", "雙12", "Double 12", "Double 12", com, greg, "December 12", "e-commerce promotions", false, fixed(12, 12, 1)},
		{"dongzhi", "冬至", "Winter solstice", "Solstice d'hiver", fest, solar, "day the Sun reaches 270 degrees (UTC+8 clock)", "tangyuan (湯圓)", false, term(ls.Chinese, 270, 0)},
		{"christmas", "聖誕節", "Christmas", "Noël", com, greg, "December 25", "gift exchange (交換禮物)", false, fixed(12, 25, 1)},
		{"weiya", "尾牙", "Year-end banquet", "Banquet de fin d'année", fest, luni, "day 16 of lunar month 12", "company banquets, raffle prizes", false, lunar(ls.Chinese, 12, 16, 0, 0)},
	},
	"kr": {
		{"seollal", "설날", "Lunar New Year", "Nouvel An lunaire", hol, luni, "day 1 of lunar month 1 (UTC+9 clock), with the day before and the day after", "gift sets (선물세트), tteokguk, travel home", false, lunar(ls.Korean, 1, 1, 1, 1)},
		{"daeboreum", "정월대보름", "First full moon", "Première pleine lune", fest, luni, "day 15 of lunar month 1", "nuts (부럼), five-grain rice", false, lunar(ls.Korean, 1, 15, 0, 0)},
		{"valentine", "발렌타인데이", "Valentine's Day", "Saint-Valentin", com, greg, "February 14", "chocolate", false, fixed(2, 14, 1)},
		{"whiteday", "화이트데이", "White Day", "White Day", com, greg, "March 14", "candy, return gifts", false, fixed(3, 14, 1)},
		{"blackday", "블랙데이", "Black Day", "Black Day", com, greg, "April 14", "jajangmyeon (짜장면)", false, fixed(4, 14, 1)},
		{"children", "어린이날", "Children's Day", "Jour des enfants", hol, greg, "May 5", "toys, family outings", false, fixed(5, 5, 1)},
		{"parents", "어버이날", "Parents' Day", "Fête des parents", com, greg, "May 8", "carnations, gifts", false, fixed(5, 8, 1)},
		{"buddha", "부처님오신날", "Buddha's Birthday", "Anniversaire de Bouddha", hol, luni, "day 8 of lunar month 4", "lanterns", false, lunar(ls.Korean, 4, 8, 0, 0)},
		{"chobok", "초복", "Chobok (first dog day)", "Chobok (première canicule)", com, sexa, "third geng day (庚) counted from the summer solstice", "samgyetang (삼계탕), chicken, watermelon", false, sambok(0)},
		{"jungbok", "중복", "Jungbok (middle dog day)", "Jungbok (canicule du milieu)", com, sexa, "fourth geng day (庚) counted from the summer solstice", "samgyetang (삼계탕), chicken, watermelon", false, sambok(1)},
		{"malbok", "말복", "Malbok (last dog day)", "Malbok (dernière canicule)", com, sexa, "first geng day (庚) counted from ipchu (Sun at 135 degrees)", "samgyetang (삼계탕), chicken, watermelon", false, sambok(2)},
		{"chuseok", "추석", "Chuseok", "Chuseok", hol, luni, "day 15 of lunar month 8 (UTC+9 clock), with the day before and the day after", "gift sets, songpyeon, travel home", false, lunar(ls.Korean, 8, 15, 1, 1)},
		{"pepero", "빼빼로데이", "Pepero Day", "Pepero Day", com, greg, "November 11", "Pepero sticks, snacks", false, fixed(11, 11, 1)},
		{"black-friday", "블랙프라이데이", "Black Friday to Cyber Monday", "Black Friday au Cyber Monday", com, greg, blackFridayRule, "online and store promotions", false, blackFriday},
		{"dongji", "동지", "Winter solstice", "Solstice d'hiver", fest, solar, "day the Sun reaches 270 degrees (UTC+9 clock)", "red bean porridge (팥죽)", false, term(ls.Korean, 270, 0)},
		{"christmas", "크리스마스", "Christmas", "Noël", hol, greg, "December 25", "gifts, cakes, dinners", false, fixed(12, 25, 1)},
	},
	"fr": {
		{"epiphanie", "Épiphanie", "Epiphany", "Épiphanie", fest, greg, "first Sunday after January 1", "galette des rois", false, epiphanie},
		{"soldes-hiver", "Soldes d'hiver", "Winter sales", "Soldes d'hiver", com, greg, "four weeks from the second Wednesday of January (the first one if the second falls after the 12th)", "clothing, home, electronics", false, soldes(1)},
		{"chandeleur", "Chandeleur", "Candlemas", "Chandeleur", fest, greg, "February 2", "crêpes", false, fixed(2, 2, 1)},
		{"saint-valentin", "Saint-Valentin", "Valentine's Day", "Saint-Valentin", com, greg, "February 14", "flowers, chocolate, restaurants, jewellery", false, fixed(2, 14, 1)},
		{"grands-meres", "Fête des grands-mères", "Grandmothers' Day", "Fête des grands-mères", com, greg, "first Sunday of March", "flowers, gifts", false, nth(3, sun, 1)},
		{"paques", "Pâques", "Easter", "Pâques", fest, "easter", "Easter Sunday and Easter Monday (public holiday)", "chocolate eggs, lamb", false, func(y int) []span { e := Easter(y); return []span{{start: e, end: e + 1}} }},
		{"fete-des-meres", "Fête des mères", "Mother's Day", "Fête des mères", com, "easter", "last Sunday of May; first Sunday of June when Pentecost falls on that day", "flowers, perfume, gifts", false, feteDesMeres},
		{"fete-des-peres", "Fête des pères", "Father's Day", "Fête des pères", com, greg, "third Sunday of June", "gifts", false, nth(6, sun, 3)},
		{"soldes-ete", "Soldes d'été", "Summer sales", "Soldes d'été", com, greg, "four weeks from the last Wednesday of June (the one before if it falls after the 28th)", "clothing, home, electronics", false, soldes(6)},
		{"rentree", "Rentrée scolaire", "Back to school", "Rentrée scolaire", com, greg, "about August 16 to September 10", "school supplies, clothing", true, fixed(8, 16, 26)},
		{"halloween", "Halloween", "Halloween", "Halloween", com, greg, "October 31", "sweets, costumes", false, fixed(10, 31, 1)},
		{"black-friday", "Black Friday", "Black Friday to Cyber Monday", "Black Friday au Cyber Monday", com, greg, blackFridayRule, "online and store promotions", false, blackFriday},
		{"noel", "Noël", "Christmas", "Noël", hol, greg, "December 25", "toys, gifts, festive food", false, fixed(12, 25, 1)},
	},
}

func iso(jdn int) string {
	y, m, d := astro.Civil(jdn)
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}

// Year returns the events of one market that start in a Gregorian year, in date order.
func Year(market string, year int) ([]Occurrence, error) {
	defs, ok := catalogue[market]
	if !ok {
		return nil, fmt.Errorf("unknown market %q (known: jp, tw, kr, fr)", market)
	}
	if year < ls.MinYear+1 || year > ls.MaxYear-1 {
		return nil, fmt.Errorf("year %d is outside %d to %d", year, ls.MinYear+1, ls.MaxYear-1)
	}
	var out []Occurrence
	for _, d := range defs {
		for _, s := range d.when(year) {
			if y, _, _ := astro.Civil(s.start); y != year {
				continue
			}
			out = append(out, Occurrence{
				ID: market + "-" + d.id, Market: market, Start: iso(s.start), End: iso(s.end),
				NameLocal: d.local, NameEN: d.en, NameFR: d.fr, Kind: d.kind, Basis: d.basis, Rule: d.rule,
				Sells: d.sells, Window: d.window, Uncertain: s.uncertain, start: s.start, end: s.end})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].start < out[j].start })
	return out, nil
}

// Range returns the events of several markets over a span of years.
func Range(markets []string, from, to int) ([]Occurrence, error) {
	var out []Occurrence
	for _, m := range markets {
		for y := from; y <= to; y++ {
			occ, err := Year(m, y)
			if err != nil {
				return nil, err
			}
			out = append(out, occ...)
		}
	}
	return out, nil
}
