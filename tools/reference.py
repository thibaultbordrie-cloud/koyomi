"""Regenerate testdata/ from three independent Python libraries.

    pip install ephem lunardate holidays
    python3 tools/reference.py

Koyomi never reads these files to answer: they are the reference its tests compare against.
"""
from pathlib import Path

import ephem
import holidays
from lunardate import LunarDate

OUT = Path(__file__).resolve().parent.parent / "testdata"


def chinese() -> None:
    rows = ["# Chinese lunisolar calendar 1912-2098: lunar year, new year (MMDD), leap month, month lengths (1 = 30 days).",
            "# Source: the Python package lunardate 0.3.0, a widely used conversion table."]
    for year in range(1912, 2099):
        leap = LunarDate.leap_month_for_year(year) or 0
        months = [(m, False) for m in range(1, 13)]
        if leap:
            months.insert(leap, (leap, True))
        starts = [LunarDate(year, m, 1, is_leap).to_solar_date() for m, is_leap in months]
        starts.append(LunarDate(year + 1, 1, 1).to_solar_date())
        bits = "".join("1" if (b - a).days == 30 else "0" for a, b in zip(starts, starts[1:]))
        rows.append(f"{year} {starts[0]:%m%d} {leap} {bits}")
    (OUT / "chinese.txt").write_text("\n".join(rows) + "\n")


def regional() -> None:
    skip = ("preceding", "second", "before", "after", "lternat", "bserved", "Eve", "day of")

    def first(found) -> str:
        found = sorted(found)
        return f"{found[0]:%m%d}" if found else "-"

    def pick(table, word) -> str:
        return first(d for d, name in table.items() if word in name and not any(s in name for s in skip))

    rows = ["# Holidays 1990-2060, one line per year: kr = Seollal, Chuseok, Buddha's Birthday; vn = Tet; jp = vernal and",
            f"# autumnal equinox days; tw = Tomb-Sweeping Day (when listed). Source: the Python package holidays {holidays.__version__}."]
    for year in range(1990, 2061):
        kr, vn, jp, tw = (cls(years=year, language="en_US") for cls in (holidays.KR, holidays.VN, holidays.JP, holidays.TW))
        rows.append(" ".join([
            str(year),
            "kr:" + ",".join([pick(kr, "Korean New Year"), pick(kr, "Chuseok"), pick(kr, "Buddha")]),
            "vn:" + first(d for d, name in vn.items() if name.strip() == "Lunar New Year"),
            "jp:" + ",".join([pick(jp, "Vernal Equinox"), pick(jp, "Autumnal Equinox")]),
            "tw:" + first(d for d, name in tw.items() if name.strip() in ("Tomb-Sweeping Day", "Tomb Sweeping Day")),
        ]))
    (OUT / "holidays.txt").write_text("\n".join(rows) + "\n")


def ephemeris() -> None:
    def tt(date) -> float:
        return ephem.julian_date(date) + ephem.delta_t(date) / 86400.0

    rows = [f"# Instants from PyEphem {ephem.__version__}, as Julian Ephemeris Days (Terrestrial Time).",
            "# moon = a new moon; sun L = the Sun reaches apparent longitude L (equinoxes and solstices)."]
    date, n = ephem.Date("1912/1/1"), 0
    while date < ephem.Date("2099/12/1"):
        new_moon = ephem.next_new_moon(date)
        if n % 29 == 0:
            rows.append(f"moon {tt(new_moon):.5f}")
        date, n = ephem.Date(new_moon + 1), n + 1
    seasons = ((0, ephem.next_vernal_equinox), (90, ephem.next_summer_solstice),
               (180, ephem.next_autumnal_equinox), (270, ephem.next_winter_solstice))
    for year in range(1915, 2100, 7):
        for longitude, find in seasons:
            rows.append(f"sun {longitude} {year} {tt(find(str(year))):.5f}")
    (OUT / "ephemeris.txt").write_text("\n".join(rows) + "\n")


if __name__ == "__main__":
    chinese()
    regional()
    ephemeris()
    print("testdata/ regenerated")
