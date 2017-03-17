package utility

import (
	"regexp"
	"testing"
	"time"

	"github.com/boatilus/peppercorn/db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestObfuscateEmail(t *testing.T) {
	cases := []struct {
		email string
		want  string
	}{
		{"user@test.com", "u***@t***.com"},
		{"u@test.com", "u***@t***.com"},
		{"user@t.com", "u***@t***.com"},
		{"@", "@"},
		{"@.", "@."},
		{".@", ".@"},
		{"ad@.com", "ad@.com"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, ObfuscateEmail(c.email))
	}
}

func benchmarkObfuscateEmail(b *testing.B, v string) {
	for n := 0; n < b.N; n++ {
		ObfuscateEmail(v)
	}
}

func BenchmarkObfuscateEmail_full(b *testing.B)        { benchmarkObfuscateEmail(b, "user@test.com") }
func BenchmarkObfuscateEmail_shortname(b *testing.B)   { benchmarkObfuscateEmail(b, "u@test.com") }
func BenchmarkObfuscateEmail_shortdomain(b *testing.B) { benchmarkObfuscateEmail(b, "user@t.com") }
func BenchmarkObfuscateEmail_justamp(b *testing.B)     { benchmarkObfuscateEmail(b, "@") }

func TestParseUserAgent(t *testing.T) {
	cases := []struct {
		ua          string
		wantBrowser string
		wantOS      string
	}{
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.97 Safari/537.11", "Chrome", "Linux "},
		{"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36", "Chrome", "Windows 10"},
	}

	for _, c := range cases {
		got := ParseUserAgent(c.ua)
		assert.Equal(t, c.wantBrowser, got.Browser)
		assert.Equal(t, c.wantOS, got.OS)
	}
}

func TestFormatTime(t *testing.T) {
	ref, err := time.Parse(time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006")
	assert.Nil(t, err)

	cases := []struct {
		then time.Time
		want string
	}{
		{ref, "less than a minute ago"},
		{ref.Add(-70 * time.Second), "about a minute ago"},
		{ref.Add(-2 * time.Minute), "2 minutes ago"},
		{ref.Add(-40 * time.Minute), "40 minutes ago"},
		{ref.Add(-59 * time.Minute), "59 minutes ago"},
		{ref.Add(-1 * time.Hour), "2:04 PM"},
		{ref.Add(-15 * time.Hour), "12:04 AM"},
		{ref.Add(-16 * time.Hour), "January 1, 2006 at 11:04 PM"},
		{ref.Add(-24 * time.Hour), "January 1, 2006 at 3:04 PM"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, FormatTime(c.then, ref))
	}
}

func benchmarkFormatTime(b *testing.B, t time.Time, current time.Time) {
	for n := 0; n < b.N; n++ {
		FormatTime(t, current)
	}
}

func setupBenchmarkFormatTime() time.Time {
	ref, err := time.Parse(time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006")
	if err != nil {
		panic(err)
	}

	return ref
}

func BenchmarkFormatTime_LT_min(b *testing.B) {
	ref := setupBenchmarkFormatTime()

	benchmarkFormatTime(b, ref, ref)
}

func BenchmarkFormatTime_about_min(b *testing.B) {
	ref := setupBenchmarkFormatTime()

	benchmarkFormatTime(b, ref.Add(-70*time.Second), ref)
}

func BenchmarkFormatTime_min_ago(b *testing.B) {
	ref := setupBenchmarkFormatTime()

	benchmarkFormatTime(b, ref.Add(-2*time.Minute), ref)
}

func BenchmarkFormatTime_timestamp(b *testing.B) {
	ref := setupBenchmarkFormatTime()

	benchmarkFormatTime(b, ref.Add(-15*time.Hour), ref)
}

func BenchmarkFormatTime_fulldate(b *testing.B) {
	ref := setupBenchmarkFormatTime()

	benchmarkFormatTime(b, ref.Add(-36*time.Hour), ref)
}

func TestGetVersionString(t *testing.T) {
	// We'll just look for a correct format..
	got := GetVersionString()

	assert.Regexp(t, regexp.MustCompile(`\d+.\d+.\d+`), got)
}

func TestGetTitle(t *testing.T) {
	viper.Set("title", "A Given Title")

	got := GetTitle()

	assert.Equal(t, "A Given Title", got)
}

func TestCommifyInt64(t *testing.T) {
	cases := []struct {
		num  int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{10000, "10,000"},
		{100000, "100,000"},
		{399313, "399,313"},
		{9223372036854775807, "9,223,372,036,854,775,807"},
		{-1, "-1"},
		{-999, "-999"},
		{-1000, "-1,000"},
		{-10000, "-10,000"},
		{-100000, "-100,000"},
		{-399313, "-399,313"},
		{-9223372036854775808, "-9,223,372,036,854,775,808"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, CommifyInt64(c.num))
	}
}

func benchmarkCommifyInt64(b *testing.B, v int64) {
	for n := 0; n < b.N; n++ {
		CommifyInt64(v)
	}
}

func BenchmarkCommifyInt64_0(b *testing.B)       { benchmarkCommifyInt64(b, 0) }
func BenchmarkCommifyInt64_8(b *testing.B)       { benchmarkCommifyInt64(b, 8) }
func BenchmarkCommifyInt64_17(b *testing.B)      { benchmarkCommifyInt64(b, 17) }
func BenchmarkCommifyInt64_371(b *testing.B)     { benchmarkCommifyInt64(b, 371) }
func BenchmarkCommifyInt64_1993(b *testing.B)    { benchmarkCommifyInt64(b, 1993) }
func BenchmarkCommifyInt64_72759(b *testing.B)   { benchmarkCommifyInt64(b, 72759) }
func BenchmarkCommifyInt64_497167(b *testing.B)  { benchmarkCommifyInt64(b, 497167) }
func BenchmarkCommifyInt64_8881679(b *testing.B) { benchmarkCommifyInt64(b, 8881679) }

func TestComputePage(t *testing.T) {
	cases := []struct {
		numPosts  db.CountType
		pageEvery db.CountType
		want      db.CountType
	}{
		{1, 5, 1},
		{5, 5, 1},
		{6, 5, 2},
		{7, 5, 2},
		{9, 5, 2},
		{10, 5, 2},
		{11, 5, 3},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, ComputePage(c.numPosts, c.pageEvery))
	}
}

func TestRemoveCRs(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Hello\r\nWorld!", "Hello\nWorld!"},
		{"Hello\r\nWorld! I like...\r\n...cake.", "Hello\nWorld! I like...\n...cake."},
		{"\r\r\n", "\n"},
	}

	for _, c := range cases {
		assert.Equal(t, c.want, RemoveCRs(c.in))
	}
}

func TestGetISO8601String(t *testing.T) {
	assert := assert.New(t)

	tm, err := time.Parse("2006-01-02T15:04:05-0700", "2006-01-02T15:04:05-0700")
	if !assert.NoError(err) {
		t.FailNow()
	}

	s := GetISO8601String(&tm)
	assert.Equal("2006-01-02T15:04:05-0700", s)

	var badTime *time.Time

	s = GetISO8601String(badTime)
	assert.Equal("", s)
}

func TestGenerateRandomNonce(t *testing.T) {
	got := GenerateRandomNonce()
	assert.NotEmpty(t, got)
}

func TestGenerateRandomRecoveryCode(t *testing.T) {
	got := GenerateRandomRecoveryCode()
	assert.NotEmpty(t, got)
	assert.Len(t, got, 12)
}
