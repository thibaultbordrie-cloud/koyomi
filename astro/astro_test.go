package astro

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

func TestCivilRoundTrip(t *testing.T) {
	if JDN(2000, 1, 1) != 2451545 {
		t.Fatalf("JDN(2000-01-01) = %d, want 2451545", JDN(2000, 1, 1))
	}
	for jdn := JDN(1900, 1, 1); jdn < JDN(2101, 1, 1); jdn += 13 {
		y, m, d := Civil(jdn)
		if JDN(y, m, d) != jdn {
			t.Fatalf("round trip failed for %d-%d-%d", y, m, d)
		}
	}
}

// The instants must agree with an independent ephemeris (PyEphem) within one minute.
func TestInstantsAgreeWithPyEphem(t *testing.T) {
	f, err := os.Open("../testdata/ephemeris.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	moons, suns, worst := 0, 0, 0.0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		var want, got float64
		switch {
		case strings.HasPrefix(line, "moon"):
			fmt.Sscanf(line, "moon %f", &want)
			got = NewMoon(Lunation(want))
			moons++
		case strings.HasPrefix(line, "sun"):
			var lon float64
			var year int
			fmt.Sscanf(line, "sun %f %d %f", &lon, &year, &want)
			got = SolarTerm(year, lon)
			suns++
		}
		diff := math.Abs(got-want) * 86400
		worst = math.Max(worst, diff)
		if diff > 60 {
			t.Errorf("%s: off by %.0f s", line, diff)
		}
	}
	if moons < 80 || suns < 100 {
		t.Fatalf("too few reference instants: %d moons, %d sun", moons, suns)
	}
	t.Logf("%d new moons and %d equinoxes or solstices, worst difference %.0f s", moons, suns, worst)
}

func TestSunLongitudeAtKnownInstants(t *testing.T) {
	// Meeus, example 25.b: 1992 October 13.0 TT, apparent longitude 199 deg 54 min 21.8 s.
	got := SunLongitude(2448908.5)
	want := 199 + 54.0/60 + 21.8/3600
	if math.Abs(got-want)*3600 > 1.5 {
		t.Fatalf("SunLongitude = %.5f, want %.5f", got, want)
	}
}

func TestNewMoonMeeusExample(t *testing.T) {
	// Meeus, example 49.a: new moon of 1977 February 18, JDE 2443192.65118.
	got := NewMoon(Lunation(2443192.65))
	if math.Abs(got-2443192.65118)*86400 > 5 {
		t.Fatalf("NewMoon = %.5f, want 2443192.65118", got)
	}
}
