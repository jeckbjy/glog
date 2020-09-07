package glog

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	df_text  dfType = iota // 普通文本
	df_d1                  // d
	df_d2                  // dd
	df_d3                  // ddd
	df_d4                  // dddd
	df_h1                  // h
	df_h2                  // hh
	df_H1                  // H
	df_H2                  // HH
	df_m1                  // m
	df_m2                  // mm
	df_M1                  // M
	df_M2                  // MM
	df_M3                  // MMM
	df_M4                  // MMMM
	df_s1                  // s
	df_s2                  // ss
	df_f                   // f
	df_ff                  // ff
	df_fff                 // fff
	df_ffff                // ffff
	df_fffff               // fffff
	df_t1                  // t
	df_t2                  // tt
	df_y1                  // y
	df_y2                  // yy
	df_y3                  // yyy
	df_y4                  // yyyy
	df_z1                  // z
	df_z2                  // zz
	df_z3                  // zzz
)

type dfType int

var keyToType = map[string]dfType{
	"d":     df_d1,
	"dd":    df_d2,
	"ddd":   df_d3,
	"dddd":  df_d4,
	"h":     df_h1,
	"hh":    df_h2,
	"H":     df_H1,
	"HH":    df_H2,
	"m":     df_m1,
	"mm":    df_m2,
	"M":     df_M1,
	"MM":    df_M2,
	"MMM":   df_M3,
	"MMMM":  df_M4,
	"s":     df_s1,
	"ss":    df_s2,
	"t":     df_t1,
	"tt":    df_t2,
	"y":     df_y1,
	"yy":    df_y2,
	"yyy":   df_y3,
	"yyyy":  df_y4,
	"z":     df_z1,
	"zz":    df_z2,
	"zzz":   df_z3,
	"f":     df_f,
	"ff":    df_ff,
	"fff":   df_fff,
	"ffff":  df_ffff,
	"fffff": df_fffff,
}

func parseType(key string) dfType {
	if t, ok := keyToType[key]; ok {
		return t
	}

	return df_text
}

type dfToken struct {
	Type dfType
	Data string
}

// https://docs.microsoft.com/en-us/dotnet/standard/base-types/custom-date-and-time-format-strings
// yyyy-MM-ddTHH:mm:ss;GMT+8
type DateFormat struct {
	Loc       *time.Location
	StdLayout string    // 标准的格式
	Tokens    []dfToken // yyyy-MM这种格式
}

// NewDateFormat 创建DateFormat
func NewDateFormat(layout string) (*DateFormat, error) {
	df := &DateFormat{}
	err := df.Parse(layout)
	return df, err
}

func (d *DateFormat) Parse(layout string) error {
	// 默认配置
	if layout == "" {
		d.StdLayout = time.RFC3339
		return nil
	}

	tokens := strings.Split(layout, ";")
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if strings.HasPrefix(t, "GMT") || strings.HasPrefix(t, "UTC") {
			if err := d.parseZone(t); err != nil {
				return err
			}
		} else if strings.HasPrefix(t, "RFC") || strings.HasPrefix(t, "ISO") {
			if err := d.parseStdLayout(layout); err != nil {
				return err
			}
		} else {
			d.parseToken(t)
		}
	}

	return nil
}

func (d *DateFormat) parseZone(t string) error {
	offset, err := strconv.Atoi(t[3:])
	if err != nil {
		return err
	}
	offset *= 3600
	d.Loc = time.FixedZone(t[:3], offset)
	return nil
}

func (d *DateFormat) parseStdLayout(t string) error {
	switch t {
	case "RFC822":
		d.StdLayout = time.RFC822
	case "RFC822Z":
		d.StdLayout = time.RFC822Z
	case "RFC850":
		d.StdLayout = time.RFC850
	case "RFC1123":
		d.StdLayout = time.RFC1123
	case "RFC1123Z":
		d.StdLayout = time.RFC1123Z
	case "RFC3339":
		d.StdLayout = time.RFC3339
	case "RFC3339Nano":
		d.StdLayout = time.RFC3339Nano
	case "ISO8601":
		d.StdLayout = "2006-01-02T15:04:05-0700"
	default:
		return fmt.Errorf("not support time format, %+v", t)
	}

	return nil
}

func (d *DateFormat) parseToken(layout string) {
	// 通过比较与前边字符的差异,分隔token
	for idx := 0; idx < len(layout); {
		p := layout[idx]
		beg := idx
		for ; idx < len(layout); idx++ {
			if c := layout[idx]; c != p {
				key := layout[beg:idx]
				d.Tokens = append(d.Tokens, dfToken{Type: parseType(key), Data: key})
				break
			}
		}

		if idx == len(layout) {
			key := layout[beg:]
			d.Tokens = append(d.Tokens, dfToken{Type: parseType(key), Data: key})
		}
	}
}

func (d *DateFormat) Format(t time.Time) string {
	if d.Loc == nil {
		t = t.Local()
	} else {
		t = t.In(d.Loc)
	}

	b := bytes.Buffer{}
	b.Grow(32)

	for _, token := range d.Tokens {
		switch token.Type {
		case df_text:
			b.WriteString(token.Data)
		case df_d1:
			dfWrite(&b, "%d", t.Day())
		case df_d2:
			dfWrite(&b, "%02d", t.Day())
		case df_d3:
			dfWrite(&b, "%s", t.Weekday().String()[:3])
		case df_d4:
			dfWrite(&b, "%s", t.Weekday().String())
		case df_h1:
			dfWrite(&b, "%d", dfToHour(t.Hour()))
		case df_h2:
			dfWrite(&b, "%02d", dfToHour(t.Hour()))
		case df_H1:
			dfWrite(&b, "%d", t.Hour())
		case df_H2:
			dfWrite(&b, "%02d", t.Hour())
		case df_m1:
			dfWrite(&b, "%d", t.Minute())
		case df_m2:
			dfWrite(&b, "%02d", t.Minute())
		case df_M1:
			dfWrite(&b, "%d", t.Month())
		case df_M2:
			dfWrite(&b, "%02d", t.Month())
		case df_M3:
			dfWrite(&b, "%s", t.Month().String()[:3])
		case df_M4:
			dfWrite(&b, "%s", t.Month().String())
		case df_s1:
			dfWrite(&b, "%d", t.Second())
		case df_s2:
			dfWrite(&b, "%02d", t.Second())
		case df_f:
			dfWrite(&b, "%d", t.Nanosecond()/1e6)
		case df_ff:
			dfWrite(&b, "%02d", t.Nanosecond()/1e6)
		case df_fff:
			dfWrite(&b, "%03d", t.Nanosecond()/1e6)
		case df_ffff:
			dfWrite(&b, "%04d", t.Nanosecond()/1e6)
		case df_fffff:
			dfWrite(&b, "%05d", t.Nanosecond()/1e6)
		case df_t1:
			dfWrite(&b, "%s", dfToAMPM(t.Hour())[0])
		case df_t2:
			dfWrite(&b, "%s", dfToAMPM(t.Hour()))
		case df_y1:
			dfWrite(&b, "%d", t.Year()%10)
		case df_y2:
			dfWrite(&b, "%02d", t.Year()%100)
		case df_y3:
			dfWrite(&b, "%03d", t.Year()%1000)
		case df_y4:
			dfWrite(&b, "%04d", t.Year())
		case df_z1:
			dfWrite(&b, "%s", t.Format("Z07"))
		case df_z2:
			dfWrite(&b, "%s", t.Format("-07"))
		case df_z3:
			dfWrite(&b, "%s", t.Format("-07:00"))
		}
	}

	return b.String()
}

func dfWrite(b *bytes.Buffer, format string, value interface{}) {
	text := fmt.Sprintf(format, value)
	b.WriteString(text)
}

func dfToHour(h int) int {
	if h < 12 {
		return h
	}

	return h - 12
}

func dfToAMPM(h int) string {
	if h < 12 {
		return "AM"
	}

	return "PM"
}

func dfIsValidZone(zone int) bool {
	return zone > -12 && zone < 12
}
