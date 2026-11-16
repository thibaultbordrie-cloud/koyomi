// Package astro computes the two things a lunisolar calendar needs: the instant of each new
// moon and the apparent longitude of the Sun. No dependency, no data file.
//
// New moons follow Jean Meeus, "Astronomical Algorithms" (2nd ed.), chapter 49. The Sun follows
// a truncated VSOP87 series for the Earth (same book, appendix III), with nutation and
// aberration. Results are in Julian Ephemeris Days (Terrestrial Time); UT converts them to
// Universal Time with the Espenak-Meeus polynomials for Delta T. Valid for 1900 to 2100.
package astro

import "math"

const (
	// Synodic is the mean length of a lunation in days.
	Synodic = 29.530588861
	// newMoon0 is the mean new moon of 2000 January 6 (lunation k = 0).
	newMoon0 = 2451550.09766
	j2000    = 2451545.0
	rad      = math.Pi / 180
)

func sinDeg(x float64) float64 { return math.Sin(math.Mod(x, 360) * rad) }

// JDN returns the Julian Day Number of a Gregorian calendar date (the day that starts at noon UT).
func JDN(year, month, day int) int {
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + 12*a - 3
	return day + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
}

// Civil returns the Gregorian date of a Julian Day Number.
func Civil(jdn int) (year, month, day int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - 146097*b/4
	d := (4*c + 3) / 1461
	e := c - 1461*d/4
	m := (5*e + 2) / 153
	return 100*b + d - 4800 + m/10, m + 3 - 12*(m/10), e - (153*m+2)/5 + 1
}

// DeltaT returns TT - UT in seconds for a decimal year (Espenak and Meeus, 2006).
func DeltaT(y float64) float64 {
	switch {
	case y < 1920:
		t := y - 1900
		return -2.79 + 1.494119*t - 0.0598939*t*t + 0.0061966*t*t*t - 0.000197*t*t*t*t
	case y < 1941:
		t := y - 1920
		return 21.20 + 0.84493*t - 0.076100*t*t + 0.0020936*t*t*t
	case y < 1961:
		t := y - 1950
		return 29.07 + 0.407*t - t*t/233 + t*t*t/2547
	case y < 1986:
		t := y - 1975
		return 45.45 + 1.067*t - t*t/260 - t*t*t/718
	case y < 2005:
		t := y - 2000
		return 63.86 + 0.3345*t - 0.060374*t*t + 0.0017275*t*t*t + 0.000651814*t*t*t*t + 0.00002373599*t*t*t*t*t
	case y < 2050:
		t := y - 2000
		return 62.92 + 0.32217*t + 0.005589*t*t
	default:
		u := (y - 1820) / 100
		return -20 + 32*u*u - 0.5628*(2150-y)
	}
}

// UT converts a Julian Ephemeris Day (TT) to a Julian Day in Universal Time.
func UT(jde float64) float64 {
	year := 2000 + (jde-j2000)/365.25
	return jde - DeltaT(year)/86400
}

// LocalDay returns the Julian Day Number of the civil day that contains the instant jde (TT),
// for a clock set tz hours ahead of UT.
func LocalDay(jde, tz float64) int {
	return int(math.Floor(UT(jde) + 0.5 + tz/24))
}

// Lunation returns the index k of the new moon closest to the instant jd.
func Lunation(jd float64) int {
	return int(math.Round((jd - newMoon0) / Synodic))
}

// NewMoon returns the instant (JDE) of new moon number k, with k = 0 on 2000 January 6.
func NewMoon(k int) float64 {
	kf := float64(k)
	t := kf / 1236.85
	t2, t3, t4 := t*t, t*t*t, t*t*t*t
	jde := newMoon0 + Synodic*kf + 0.00015437*t2 - 0.000000150*t3 + 0.00000000073*t4
	e := 1 - 0.002516*t - 0.0000074*t2
	m := 2.5534 + 29.10535670*kf - 0.0000014*t2 - 0.00000011*t3                      // Sun's mean anomaly
	mp := 201.5643 + 385.81693528*kf + 0.0107582*t2 + 0.00001238*t3 - 0.000000058*t4 // Moon's mean anomaly
	f := 160.7108 + 390.67050284*kf - 0.0016118*t2 - 0.00000227*t3 + 0.000000011*t4  // argument of latitude
	om := 124.7746 - 1.56375588*kf + 0.0020672*t2 + 0.00000215*t3                    // ascending node
	jde += -0.40720*sinDeg(mp) + 0.17241*e*sinDeg(m) + 0.01608*sinDeg(2*mp) + 0.01039*sinDeg(2*f) +
		0.00739*e*sinDeg(mp-m) - 0.00514*e*sinDeg(mp+m) + 0.00208*e*e*sinDeg(2*m) - 0.00111*sinDeg(mp-2*f) -
		0.00057*sinDeg(mp+2*f) + 0.00056*e*sinDeg(2*mp+m) - 0.00042*sinDeg(3*mp) + 0.00042*e*sinDeg(m+2*f) +
		0.00038*e*sinDeg(m-2*f) - 0.00024*e*sinDeg(2*mp-m) - 0.00017*sinDeg(om) - 0.00007*sinDeg(mp+2*m) +
		0.00004*sinDeg(2*mp-2*f) + 0.00004*sinDeg(3*m) + 0.00003*sinDeg(mp+m-2*f) + 0.00003*sinDeg(2*mp+2*f) -
		0.00003*sinDeg(mp+m+2*f) + 0.00003*sinDeg(mp-m+2*f) - 0.00002*sinDeg(mp-m-2*f) -
		0.00002*sinDeg(3*mp+m) + 0.00002*sinDeg(4*mp)
	// Planetary arguments.
	jde += 0.000325*sinDeg(299.77+0.107408*kf-0.009173*t2) + 0.000165*sinDeg(251.88+0.016321*kf) +
		0.000164*sinDeg(251.83+26.651886*kf) + 0.000126*sinDeg(349.42+36.412478*kf) +
		0.000110*sinDeg(84.66+18.206239*kf) + 0.000062*sinDeg(141.74+53.303771*kf) +
		0.000060*sinDeg(207.14+2.453732*kf) + 0.000056*sinDeg(154.84+7.306860*kf) +
		0.000047*sinDeg(34.52+27.261239*kf) + 0.000042*sinDeg(207.19+0.121824*kf) +
		0.000040*sinDeg(291.34+1.844379*kf) + 0.000037*sinDeg(161.72+24.198154*kf) +
		0.000035*sinDeg(239.56+25.513099*kf) + 0.000023*sinDeg(331.55+3.592518*kf)
	return jde
}

// Earth's heliocentric longitude, VSOP87 truncated: amplitude (1e-8 rad), phase, frequency.
var earthL = [][][3]float64{
	{{175347046, 0, 0}, {3341656, 4.6692568, 6283.07585}, {34894, 4.6261, 12566.1517}, {3497, 2.7441, 5753.3849},
		{3418, 2.8289, 3.5231}, {3136, 3.6277, 77713.7715}, {2676, 4.4181, 7860.4194}, {2343, 6.1352, 3930.2097},
		{1324, 0.7425, 11506.7698}, {1273, 2.0371, 529.691}, {1199, 1.1096, 1577.3435}, {990, 5.233, 5884.927},
		{902, 2.045, 26.298}, {857, 3.508, 398.149}, {780, 1.179, 5223.694}, {753, 2.533, 5507.553},
		{505, 4.583, 18849.228}, {492, 4.205, 775.523}, {357, 2.92, 0.067}, {317, 5.849, 11790.629},
		{284, 1.899, 796.298}, {271, 0.315, 10977.079}, {243, 0.345, 5486.778}, {206, 4.806, 2544.314},
		{205, 1.869, 5573.143}, {202, 2.458, 6069.777}, {156, 0.833, 213.299}, {132, 3.411, 2942.463},
		{126, 1.083, 20.775}, {115, 0.645, 0.98}, {103, 0.636, 4694.003}, {102, 0.976, 15720.839},
		{102, 4.267, 7.114}, {99, 6.21, 2146.17}, {98, 0.68, 155.42}, {86, 5.98, 161000.69}, {85, 1.3, 6275.96},
		{85, 3.67, 71430.7}, {80, 1.81, 17260.15}, {79, 3.04, 12036.46}, {75, 1.76, 5088.63}, {74, 3.5, 3154.69},
		{74, 4.68, 801.82}, {70, 0.83, 9437.76}, {62, 3.98, 8827.39}, {61, 1.82, 7084.9}, {57, 2.78, 6286.6},
		{56, 4.39, 14143.5}, {56, 3.47, 6279.55}, {52, 0.19, 12139.55}, {52, 1.33, 1748.02}, {51, 0.28, 5856.48},
		{49, 0.49, 1194.45}, {41, 5.37, 8429.24}, {41, 2.4, 19651.05}, {39, 6.17, 10447.39}, {37, 6.04, 10213.29},
		{37, 2.57, 1059.38}, {36, 1.71, 2352.87}, {36, 1.78, 6812.77}, {33, 0.59, 17789.85}, {30, 0.44, 83996.85},
		{30, 2.74, 1349.87}, {25, 3.16, 4690.48}},
	{{628331966747, 0, 0}, {206059, 2.678235, 6283.07585}, {4303, 2.6351, 12566.1517}, {425, 1.59, 3.523},
		{119, 5.796, 26.298}, {109, 2.966, 1577.344}, {93, 2.59, 18849.23}, {72, 1.14, 529.69}, {68, 1.87, 398.15},
		{67, 4.41, 5507.55}, {59, 2.89, 5223.69}, {56, 2.17, 155.42}, {45, 0.4, 796.3}, {36, 0.47, 775.52},
		{29, 2.65, 7.11}, {21, 5.34, 0.98}, {19, 1.85, 5486.78}, {19, 4.97, 213.3}, {17, 2.99, 6275.96},
		{16, 0.03, 2544.31}, {16, 1.43, 2146.17}, {15, 1.21, 10977.08}, {12, 2.83, 1748.02}, {12, 3.26, 5088.63},
		{12, 5.27, 1194.45}, {12, 2.08, 4694}, {11, 0.77, 553.57}, {10, 1.3, 6286.6}, {10, 4.24, 1349.87},
		{9, 2.7, 242.73}, {9, 5.64, 951.72}, {8, 5.3, 2352.87}, {6, 2.65, 9437.76}, {6, 4.67, 4690.48}},
	{{52919, 0, 0}, {8720, 1.0721, 6283.0758}, {309, 0.867, 12566.152}, {27, 0.05, 3.52}, {16, 5.19, 26.3},
		{16, 3.68, 155.42}, {10, 0.76, 18849.23}, {9, 2.06, 77713.77}, {7, 0.83, 775.52}, {5, 4.66, 1577.34},
		{4, 1.03, 7.11}, {4, 3.44, 5573.14}, {3, 5.14, 796.3}, {3, 6.05, 5507.55}, {3, 1.19, 242.73},
		{3, 6.12, 529.69}, {3, 0.31, 398.15}, {3, 2.28, 553.57}, {2, 4.38, 5223.69}, {2, 3.75, 0.98}},
	{{289, 5.844, 6283.076}, {35, 0, 0}, {17, 5.49, 12566.15}, {3, 5.2, 155.42}, {1, 4.72, 3.52},
		{1, 5.3, 18849.23}, {1, 5.97, 242.73}},
	{{114, 3.142, 0}, {8, 4.13, 6283.08}, {1, 3.84, 12566.15}},
	{{1, 3.14, 0}},
}

// SunLongitude returns the apparent geocentric longitude of the Sun in degrees [0, 360),
// referred to the true equinox of the date, at the instant jde (TT).
func SunLongitude(jde float64) float64 {
	tau := (jde - j2000) / 365250
	l, pow := 0.0, 1.0
	for _, series := range earthL {
		sum := 0.0
		for _, c := range series {
			sum += c[0] * math.Cos(c[1]+c[2]*tau)
		}
		l += sum * pow
		pow *= tau
	}
	lon := l/1e8/rad + 180 // heliocentric Earth -> geocentric Sun
	t := (jde - j2000) / 36525
	lon -= 0.09033 / 3600 // FK5
	// Nutation in longitude (main terms) and aberration.
	om := 125.04452 - 1934.136261*t
	ls := 280.4665 + 36000.7698*t
	lm := 218.3165 + 481267.8813*t
	lon += (-17.20*sinDeg(om) - 1.32*sinDeg(2*ls) - 0.23*sinDeg(2*lm) + 0.21*sinDeg(2*om)) / 3600
	m := 357.52911 + 35999.05029*t
	r := 1.000140 - 0.016708*math.Cos(m*rad) - 0.000139*math.Cos(2*m*rad)
	lon -= 20.4898 / 3600 / r
	lon = math.Mod(lon, 360)
	if lon < 0 {
		lon += 360
	}
	return lon
}

// SolarTerm returns the instant (JDE) at which the Sun reaches the given longitude (degrees)
// during the given Gregorian year. Longitudes run from 285 (early January) through 0 (March
// equinox) to 270 (December solstice): each value is reached once per calendar year.
func SolarTerm(year int, longitude float64) float64 {
	// Days after the March equinox (about March 20.5) for this longitude; January and February
	// longitudes (270 < L < 360) belong to the start of the year.
	offset := longitude
	if longitude > 275 {
		offset = longitude - 360
	}
	jd := float64(JDN(year, 3, 20)) + offset/360*365.2422
	for i := 0; i < 8; i++ {
		diff := math.Mod(longitude-SunLongitude(jd)+540, 360) - 180
		jd += diff / 360 * 365.2422
		if math.Abs(diff) < 1e-7 {
			break
		}
	}
	return jd
}
