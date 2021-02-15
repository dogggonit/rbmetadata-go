package metadata

import (
	"math"
	"rbmetadata-go/lib/fixedpoint"
	"rbmetadata-go/tools"
	"strconv"
	"strings"
	"unicode"
)

const (
	FpBits = 12
	FpOne  = 1 << FpBits
	FpMin  = -48 * FpOne
	FpMax  = 17 * FpOne
)

// Set ReplayGain values from integers. Existing values are not overwritten.
//
// album   If true, set album values, otherwise set track values.
// gain    Gain value in dB, multiplied by 512. 0 for no gain.
// peak    Peak volume in Q7.24 format, where 1.0 is full scale. 0 for no
//         peak volume.
// entry   mp3entry struct to update.
func ParseReplayGainInt(album bool, gain, peak int64, entry *Mp3Entry) {
	gain *= FpOne / 512

	if album {
		entry.AlbumLevel = gain
		entry.AlbumGain = ConvertGain(gain)
		entry.AlbumPeak = peak
	} else {
		entry.TrackLevel = gain
		entry.TrackGain = ConvertGain(gain)
		entry.TrackPeak = peak
	}
}

// Parse a ReplayGain tag conforming to the "VorbisGain standard". If a
// valid tag is found, update mp3entry struct accordingly. Existing values
// are not overwritten.
//
// key     Name of the tag.
// value   Value of the tag.
// entry   mp3entry struct to update.
func ParseReplayGain(key string, value string, entry *Mp3Entry) {
	if entry.TrackGain == 0 {
		if tools.Strcasecmp(key, "replaygain_track_gain") || tools.Strcasecmp(key, "rg_radio") {
			entry.TrackLevel = GetReplayGain(value)
			entry.TrackGain = ConvertGain(entry.TrackLevel)
		} else if tools.Strcasecmp(key, "replaygain_album_gain") || tools.Strcasecmp(key, "rg_audiophile") {
			entry.AlbumLevel = GetReplayGain(value)
			entry.AlbumGain = ConvertGain(entry.AlbumLevel)
		} else if tools.Strcasecmp(key, "replaygain_track_peak") || tools.Strcasecmp(key, "rg_peak") {
			entry.TrackPeak = GetReplayPeak(value)
		} else if tools.Strcasecmp(key, "replaygain_album_peak") {
			entry.AlbumPeak = GetReplayPeak(value)
		}
	}
}

// Get the peak volume in Q7.24 format.
//
// str  Peak volume. Full scale is specified as "1.0". Returns 0 for no peak.
func GetReplayPeak(str string) int64 {
	return FPatof(str, 24)
}

// Get the sample scale factor in Q19.12 format from a gain value. Returns 0
// for no gain.
//
// str  Gain in dB as a string. E.g., "-3.45 dB"; the "dB" part is ignored.
func GetReplayGain(str string) int64 {
	return FPatof(str, FpBits)
}

func ConvertGain(gain int64) int64 {
	// Don't allow unreasonably low or high gain changes.
	// Our math code can't handle it properly anyway. :)
	gain = int64(math.Max(float64(gain), float64(FpMin)))
	gain = int64(math.Min(float64(gain), float64(FpMax)))

	return fixedpoint.FpFactor(gain, FpBits) << (24 - FpBits)
}

func FPatof(s string, precision int) int64 {
	var intPart int64 = 0
	var intOne int64 = 1 << precision
	var fracPart int64
	var fracCount int64
	var fracMax = ((int64(precision) * 4) + 12) / 13
	var fracMaxInt int64 = 1
	var sign int64 = 1
	var point = false

	s = strings.TrimSpace(s)

	chars := []rune(s)
	if len(chars) > 0 {
		if chars[0] == '-' {
			sign = -1
			s = s[1:]
		} else if chars[0] == '+' {
			s = s[1:]
		}
	}

	for _, c := range chars {
		if c == '.' {
			if point {
				break
			}
			point = true
		} else if unicode.IsDigit(c) {
			i, _ := strconv.Atoi(string(c)) // Already tested to see if number, no need to check for error
			if point {
				if fracCount < fracMax {
					fracPart = (fracPart * 10) + int64(i)
					fracCount++
					fracMaxInt *= 10
				}
			} else {
				intPart = (intPart * 10) + int64(i)
			}
		} else {
			break
		}
	}

	for fracCount < fracMax {
		fracPart *= 10
		fracCount++
		fracMaxInt *= 10
	}

	return sign * ((intPart * intOne) + ((fracPart * intOne) / fracMaxInt))
}
