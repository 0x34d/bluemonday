package bluemonday

import (
	"regexp"
	"sync"
	"testing"
)

func TestEmpty(t *testing.T) {
	p := StrictPolicy()

	_, err := p.Sanitize(``)
	if err == nil {
		t.Error(err)
	}
}

func TestStrictPolicy(t *testing.T) {
	p := StrictPolicy()

	in := "Hello, <b>World</b>!"
	expected := "Hello, !"

	out, err := p.Sanitize(in)
	if err != nil {
		t.Error(err)
	}
	if out != expected {
		t.Errorf(
			"test 1 failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			in,
			out,
			expected,
		)
	}
}

func TestAllowDocType(t *testing.T) {
	p := NewPolicy()
	p.AllowElements("b")

	in := "<!DOCTYPE html>Hello, <b>World</b>!"
	expected := "Hello, <b>World</b>!"

	out, err := p.Sanitize(in)
	if err != nil {
		t.Error(err)
	}
	if out != expected {
		t.Errorf(
			"test 1 failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			in,
			out,
			expected,
		)
	}

	// Allow the doctype and run the test again
	p.AllowDocType(true)

	expected = "<!DOCTYPE html>Hello, <b>World</b>!"

	out, err = p.Sanitize(in)
	if err != nil {
		t.Error(err)
	}
	if out != expected {
		t.Errorf(
			"test 1 failed;\ninput   : %s\noutput  : %s\nexpected: %s",
			in,
			out,
			expected,
		)
	}
}

func TestUGCPolicy(t *testing.T) {

	type test struct {
		in       string
		expected string
	}

	tests := []test{
		// Simple formatting
		test{in: "Hello, World!", expected: "Hello, World!"},
		test{in: "Hello, <b>World</b>!", expected: "Hello, <b>World</b>!"},
		// Blocks and formatting
		test{
			in:       "<p>Hello, <b onclick=alert(1337)>World</b>!</p>",
			expected: "<p>Hello, <b>World</b>!</p>",
		},
		test{
			in:       "<p onclick=alert(1337)>Hello, <b>World</b>!</p>",
			expected: "<p>Hello, <b>World</b>!</p>",
		},
		// Inline tags featuring globals
		test{
			// TODO: Need to add rel="nofollow" to this
			in: `<a href="http://example.org/" rel="nofollow">Hello, <b>World</b></a>` +
				`<a href="https://example.org/#!" rel="nofollow">!</a>`,
			expected: `<a href="http://example.org/">Hello, <b>World</b></a>` +
				`<a href="https://example.org/#!">!</a>`,
		},
		test{
			// TODO: Need to add rel="nofollow" to this
			in: `Hello, <b>World</b>` +
				`<a title="!" href="https://example.org/#!" rel="nofollow">!</a>`,
			expected: `Hello, <b>World</b>` +
				`<a title="!" href="https://example.org/#!">!</a>`,
		},
		// Images
		test{
			in:       `<a href="javascript:alert(1337)">foo</a>`,
			expected: `foo`,
		},
		test{
			in:       `<img src="http://example.org/foo.gif">`,
			expected: `<img src="http://example.org/foo.gif">`,
		},
		test{
			in:       `<img src="http://example.org/x.gif" alt="y" width=96 height=64 border=0>`,
			expected: `<img src="http://example.org/x.gif" alt="y" width="96" height="64">`,
		},
		test{
			in:       `<img src="http://example.org/x.png" alt="y" width="widgy" height=64 border=0>`,
			expected: `<img src="http://example.org/x.png" alt="y" height="64">`,
		},
		// Anchors
		// TODO: Need to add rel="nofollow" to all of these
		// test{
		// 	// TODO: Need to add support for local links
		// 	in:       `<a href="foo.html">Link text</a>`,
		// 	expected: `<a href="foo.html">Link text</a>`,
		// },
		// // test{
		// 	// TODO: Need to add support for local links
		// 	in:       `<a href="foo.html" onclick="alert(1337)">Link text</a>`,
		// 	expected: `<a href="foo.html">Link text</a>`,
		// },
		test{
			in:       `<a href="http://example.org/x.html" onclick="alert(1337)">Link text</a>`,
			expected: `<a href="http://example.org/x.html">Link text</a>`,
		},
		test{
			in:       `<a href="https://example.org/x.html" onclick="alert(1337)">Link text</a>`,
			expected: `<a href="https://example.org/x.html">Link text</a>`,
		},
		test{
			in:       `<a href="HTTPS://example.org/x.html" onclick="alert(1337)">Link text</a>`,
			expected: `<a href="HTTPS://example.org/x.html">Link text</a>`,
		},
		// test{
		// 	// TODO: Need to add support for protocol links: //example.org
		// 	in:       `<a href="//example.org/x.html" onclick="alert(1337)">Link text</a>`,
		// 	expected: `<a href="//example.org/x.html">Link text</a>`,
		// },
		test{
			in:       `<a href="javascript:alert(1337).html" onclick="alert(1337)">Link text</a>`,
			expected: `Link text`,
		},
		test{
			in:       `<a name="header" id="header">Header text</a>`,
			expected: `<a id="header">Header text</a>`,
		},
		// Tables
		test{
			in: `<table style="color: rgb(0, 0, 0);">` +
				`<tbody>` +
				`<tr>` +
				`<th>Column One</th><th>Column Two</th>` +
				`</tr>` +
				`<tr>` +
				`<td align="center"` +
				` style="background-color: rgb(255, 255, 254);">` +
				`<font size="2">Size 2</font></td>` +
				`<td align="center"` +
				` style="background-color: rgb(255, 255, 254);">` +
				`<font size="7">Size 7</font></td>` +
				`</tr>` +
				`</tbody>` +
				`</table>`,
			expected: "" +
				`<table>` +
				`<tbody>` +
				`<tr>` +
				`<th>Column One</th><th>Column Two</th>` +
				`</tr>` +
				`<tr>` +
				`<td align="center"></td>` +
				`<td align="center"></td>` +
				`</tr>` +
				`</tbody>` +
				`</table>`,
		},
		// Ordering
		test{
			in: `xss<a href="http://www.google.de" style="color:red;"` +
				` onmouseover=alert(1) onmousemove="alert(2)" onclick=alert(3)>` +
				`g<img src="http://example.org"/>oogle</a>`,
			expected: `xss<a href="http://www.google.de"` +
				`>` +
				`g<img src="http://example.org"/>oogle</a>`,
		},
	}

	p := UGCPolicy()

	for ii, test := range tests {
		out, err := p.Sanitize(test.in)
		if err != nil {
			t.Error(err)
		}
		if out != test.expected {
			t.Errorf(
				"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
				ii,
				test.in,
				out,
				test.expected,
			)
		}
	}
}

// // TODO: Fix bad HTML attributes
// // Currently fails as disabled on the textarea should be converted to
// // disabled="disabled" in the HTML rather than disabled=""
// func TestEmptyAttributes(t *testing.T) {
//
// 	p := NewPolicy()
// 	// Do not do this, especially without a Matching() clause, this is a test
// 	p.AllowAttrs("disabled").OnElements("textarea")
// 	p.AllowElements("span", "div")//
//
// 	type test struct {
// 		in       string
// 		expected string
// 	}
//
// 	tests := []test{
// 		// Empty elements
// 		test{
// 			in: `<textarea>text</textarea><textarea disabled></textarea>` +
// 				`<div onclick='redirect()'><span>Styled by span</span></div>`,
// 			expected: `text<textarea disabled="disabled"></textarea>` +
// 				`<div><span>Styled by span</span></div>`,
// 		},
// 	}
//
// 	for ii, test := range tests {
// 		out, err := p.Sanitize(test.in)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		if out != test.expected {
// 			t.Errorf(
// 				"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
// 				ii,
// 				test.in,
// 				out,
// 				test.expected,
// 			)
// 		}
// 	}
// }

func TestAntiSamy(t *testing.T) {

	standardUrls := regexp.MustCompile(`(?i)^https?|mailto`)

	p := NewPolicy()

	p.AllowElements(
		"a", "b", "br", "div", "font", "i", "img", "input", "li", "ol", "p",
		"span", "td", "ul",
	)
	p.AllowAttrs("checked", "type").OnElements("input")
	p.AllowAttrs("color").OnElements("font")
	p.AllowAttrs("href").Matching(standardUrls).OnElements("a")
	p.AllowAttrs("src").Matching(standardUrls).OnElements("img")
	p.AllowAttrs("class", "id", "title").Globally()
	p.AllowAttrs("char").Matching(
		regexp.MustCompile(`p{L}`), // Single character or HTML entity only
	).OnElements("td")

	type test struct {
		in       string
		expected string
	}

	tests := []test{
		// Base64 strings
		//
		// first string is
		// <a - href="http://www.owasp.org">click here</a>
		test{
			in:       `PGEgLSBocmVmPSJodHRwOi8vd3d3Lm93YXNwLm9yZyI+Y2xpY2sgaGVyZTwvYT4=`,
			expected: `PGEgLSBocmVmPSJodHRwOi8vd3d3Lm93YXNwLm9yZyI+Y2xpY2sgaGVyZTwvYT4=`,
		},
		// the rest are randomly generated 300 byte sequences which generate
		// parser errors, turned into Strings
		test{
			in:       `uz0sEy5aDiok6oufQRaYPyYOxbtlACRnfrOnUVIbOstiaoB95iw+dJYuO5sI9nudhRtSYLANlcdgO0pRb+65qKDwZ5o6GJRMWv4YajZk+7Q3W/GN295XmyWUpxuyPGVi7d5fhmtYaYNW6vxyKK1Wjn9IEhIrfvNNjtEF90vlERnz3wde4WMaKMeciqgDXuZHEApYmUcu6Wbx4Q6WcNDqohAN/qCli74tvC+Umy0ZsQGU7E+BvJJ1tLfMcSzYiz7Q15ByZOYrA2aa0wDu0no3gSatjGt6aB4h30D9xUP31LuPGZ2GdWwMfZbFcfRgDSh42JPwa1bODmt5cw0Y8ACeyrIbfk9IkX1bPpYfIgtO7TwuXjBbhh2EEixOZ2YkcsvmcOSVTvraChbxv6kP`,
			expected: `uz0sEy5aDiok6oufQRaYPyYOxbtlACRnfrOnUVIbOstiaoB95iw+dJYuO5sI9nudhRtSYLANlcdgO0pRb+65qKDwZ5o6GJRMWv4YajZk+7Q3W/GN295XmyWUpxuyPGVi7d5fhmtYaYNW6vxyKK1Wjn9IEhIrfvNNjtEF90vlERnz3wde4WMaKMeciqgDXuZHEApYmUcu6Wbx4Q6WcNDqohAN/qCli74tvC+Umy0ZsQGU7E+BvJJ1tLfMcSzYiz7Q15ByZOYrA2aa0wDu0no3gSatjGt6aB4h30D9xUP31LuPGZ2GdWwMfZbFcfRgDSh42JPwa1bODmt5cw0Y8ACeyrIbfk9IkX1bPpYfIgtO7TwuXjBbhh2EEixOZ2YkcsvmcOSVTvraChbxv6kP`,
		},
		test{
			in:       `PIWjMV4y+MpuNLtcY3vBRG4ZcNaCkB9wXJr3pghmFA6rVXAik+d5lei48TtnHvfvb5rQZVceWKv9cR/9IIsLokMyN0omkd8j3TV0DOh3JyBjPHFCu1Gp4Weo96h5C6RBoB0xsE4QdS2Y1sq/yiha9IebyHThAfnGU8AMC4AvZ7DDBccD2leZy2Q617ekz5grvxEG6tEcZ3fCbJn4leQVVo9MNoerim8KFHGloT+LxdgQR6YN5y1ii3bVGreM51S4TeANujdqJXp8B7B1Gk3PKCRS2T1SNFZedut45y+/w7wp5AUQCBUpIPUj6RLp+y3byWhcbZbJ70KOzTSZuYYIKLLo8047Fej43bIaghJm0F9yIKk3C5gtBcw8T5pciJoVXrTdBAK/8fMVo29P`,
			expected: `PIWjMV4y+MpuNLtcY3vBRG4ZcNaCkB9wXJr3pghmFA6rVXAik+d5lei48TtnHvfvb5rQZVceWKv9cR/9IIsLokMyN0omkd8j3TV0DOh3JyBjPHFCu1Gp4Weo96h5C6RBoB0xsE4QdS2Y1sq/yiha9IebyHThAfnGU8AMC4AvZ7DDBccD2leZy2Q617ekz5grvxEG6tEcZ3fCbJn4leQVVo9MNoerim8KFHGloT+LxdgQR6YN5y1ii3bVGreM51S4TeANujdqJXp8B7B1Gk3PKCRS2T1SNFZedut45y+/w7wp5AUQCBUpIPUj6RLp+y3byWhcbZbJ70KOzTSZuYYIKLLo8047Fej43bIaghJm0F9yIKk3C5gtBcw8T5pciJoVXrTdBAK/8fMVo29P`,
		},
		test{
			in:       `uCk7HocubT6KzJw2eXpSUItZFGkr7U+D89mJw70rxdqXP2JaG04SNjx3dd84G4bz+UVPPhPO2gBAx2vHI0xhgJG9T4vffAYh2D1kenmr+8gIHt6WDNeD+HwJeAbJYhfVFMJsTuIGlYIw8+I+TARK0vqjACyRwMDAndhXnDrk4E5U3hyjqS14XX0kIDZYM6FGFPXe/s+ba2886Q8o1a7WosgqqAmt4u6R3IHOvVf5/PIeZrBJKrVptxjdjelP8Xwjq2ujWNtR3/HM1kjRlJi4xedvMRe4Rlxek0NDLC9hNd18RYi0EjzQ0bGSDDl0813yv6s6tcT6xHMzKvDcUcFRkX6BbxmoIcMsVeHM/ur6yRv834o/TT5IdiM9/wpkuICFOWIfM+Y8OWhiU6BK`,
			expected: `uCk7HocubT6KzJw2eXpSUItZFGkr7U+D89mJw70rxdqXP2JaG04SNjx3dd84G4bz+UVPPhPO2gBAx2vHI0xhgJG9T4vffAYh2D1kenmr+8gIHt6WDNeD+HwJeAbJYhfVFMJsTuIGlYIw8+I+TARK0vqjACyRwMDAndhXnDrk4E5U3hyjqS14XX0kIDZYM6FGFPXe/s+ba2886Q8o1a7WosgqqAmt4u6R3IHOvVf5/PIeZrBJKrVptxjdjelP8Xwjq2ujWNtR3/HM1kjRlJi4xedvMRe4Rlxek0NDLC9hNd18RYi0EjzQ0bGSDDl0813yv6s6tcT6xHMzKvDcUcFRkX6BbxmoIcMsVeHM/ur6yRv834o/TT5IdiM9/wpkuICFOWIfM+Y8OWhiU6BK`,
		},
		test{
			in:       `Bb6Cqy6stJ0YhtPirRAQ8OXrPFKAeYHeuZXuC1qdHJRlweEzl4F2z/ZFG7hzr5NLZtzrRG3wm5TXl6Aua5G6v0WKcjJiS2V43WB8uY1BFK1d2y68c1gTRSF0u+VTThGjz+q/R6zE8HG8uchO+KPw64RehXDbPQ4uadiL+UwfZ4BzY1OHhvM5+2lVlibG+awtH6qzzx6zOWemTih932Lt9mMnm3FzEw7uGzPEYZ3aBV5xnbQ2a2N4UXIdm7RtIUiYFzHcLe5PZM/utJF8NdHKy0SPaKYkdXHli7g3tarzAabLZqLT4k7oemKYCn/eKRreZjqTB2E8Kc9Swf3jHDkmSvzOYE8wi1vQ3X7JtPcQ2O4muvpSa70NIE+XK1CgnnsL79Qzci1/1xgkBlNq`,
			expected: `Bb6Cqy6stJ0YhtPirRAQ8OXrPFKAeYHeuZXuC1qdHJRlweEzl4F2z/ZFG7hzr5NLZtzrRG3wm5TXl6Aua5G6v0WKcjJiS2V43WB8uY1BFK1d2y68c1gTRSF0u+VTThGjz+q/R6zE8HG8uchO+KPw64RehXDbPQ4uadiL+UwfZ4BzY1OHhvM5+2lVlibG+awtH6qzzx6zOWemTih932Lt9mMnm3FzEw7uGzPEYZ3aBV5xnbQ2a2N4UXIdm7RtIUiYFzHcLe5PZM/utJF8NdHKy0SPaKYkdXHli7g3tarzAabLZqLT4k7oemKYCn/eKRreZjqTB2E8Kc9Swf3jHDkmSvzOYE8wi1vQ3X7JtPcQ2O4muvpSa70NIE+XK1CgnnsL79Qzci1/1xgkBlNq`,
		},
		test{
			in:       `FZNVr4nOICD1cNfAvQwZvZWi+P4I2Gubzrt+wK+7gLEY144BosgKeK7snwlA/vJjPAnkFW72APTBjY6kk4EOyoUef0MxRnZEU11vby5Ru19eixZBFB/SVXDJleLK0z3zXXE8U5Zl5RzLActHakG8Psvdt8TDscQc4MPZ1K7mXDhi7FQdpjRTwVxFyCFoybQ9WNJNGPsAkkm84NtFb4KjGpwVC70oq87tM2gYCrNgMhBfdBl0bnQHoNBCp76RKdpq1UAY01t1ipfgt7BoaAr0eTw1S32DezjfkAz04WyPTzkdBKd3b44rX9dXEbm6szAz0SjgztRPDJKSMELjq16W2Ua8d1AHq2Dz8JlsvGzi2jICUjpFsIfRmQ/STSvOT8VsaCFhwL1zDLbn5jCr`,
			expected: `FZNVr4nOICD1cNfAvQwZvZWi+P4I2Gubzrt+wK+7gLEY144BosgKeK7snwlA/vJjPAnkFW72APTBjY6kk4EOyoUef0MxRnZEU11vby5Ru19eixZBFB/SVXDJleLK0z3zXXE8U5Zl5RzLActHakG8Psvdt8TDscQc4MPZ1K7mXDhi7FQdpjRTwVxFyCFoybQ9WNJNGPsAkkm84NtFb4KjGpwVC70oq87tM2gYCrNgMhBfdBl0bnQHoNBCp76RKdpq1UAY01t1ipfgt7BoaAr0eTw1S32DezjfkAz04WyPTzkdBKd3b44rX9dXEbm6szAz0SjgztRPDJKSMELjq16W2Ua8d1AHq2Dz8JlsvGzi2jICUjpFsIfRmQ/STSvOT8VsaCFhwL1zDLbn5jCr`,
		},
		test{
			in:       `RuiRkvYjH2FcCjNzFPT2PJWh7Q6vUbfMadMIEnw49GvzTmhk4OUFyjY13GL52JVyqdyFrnpgEOtXiTu88Cm+TiBI7JRh0jRs3VJRP3N+5GpyjKX7cJA46w8PrH3ovJo3PES7o8CSYKRa3eUs7BnFt7kUCvMqBBqIhTIKlnQd2JkMNnhhCcYdPygLx7E1Vg+H3KybcETsYWBeUVrhRl/RAyYJkn6LddjPuWkDdgIcnKhNvpQu4MMqF3YbzHgyTh7bdWjy1liZle7xR/uRbOrRIRKTxkUinQGEWyW3bbXOvPO71E7xyKywBanwg2FtvzOoRFRVF7V9mLzPSqdvbM7VMQoLFob2UgeNLbVHkWeQtEqQWIV5RMu3+knhoqGYxP/3Srszp0ELRQy/xyyD`,
			expected: `RuiRkvYjH2FcCjNzFPT2PJWh7Q6vUbfMadMIEnw49GvzTmhk4OUFyjY13GL52JVyqdyFrnpgEOtXiTu88Cm+TiBI7JRh0jRs3VJRP3N+5GpyjKX7cJA46w8PrH3ovJo3PES7o8CSYKRa3eUs7BnFt7kUCvMqBBqIhTIKlnQd2JkMNnhhCcYdPygLx7E1Vg+H3KybcETsYWBeUVrhRl/RAyYJkn6LddjPuWkDdgIcnKhNvpQu4MMqF3YbzHgyTh7bdWjy1liZle7xR/uRbOrRIRKTxkUinQGEWyW3bbXOvPO71E7xyKywBanwg2FtvzOoRFRVF7V9mLzPSqdvbM7VMQoLFob2UgeNLbVHkWeQtEqQWIV5RMu3+knhoqGYxP/3Srszp0ELRQy/xyyD`,
		},
		test{
			in:       `mqBEVbNnL929CUA3sjkOmPB5dL0/a0spq8LgbIsJa22SfP580XduzUIKnCtdeC9TjPB/GEPp/LvEUFaLTUgPDQQGu3H5UCZyjVTAMHl45me/0qISEf903zFFqW5Lk3TS6iPrithqMMvhdK29Eg5OhhcoHS+ALpn0EjzUe86NywuFNb6ID4o8aF/ztZlKJegnpDAm3JuhCBauJ+0gcOB8GNdWd5a06qkokmwk1tgwWat7cQGFIH1NOvBwRMKhD51MJ7V28806a3zkOVwwhOiyyTXR+EcDA/aq5acX0yailLWB82g/2GR/DiaqNtusV+gpcMTNYemEv3c/xLkClJc29DSfTsJGKsmIDMqeBMM7RRBNinNAriY9iNX1UuHZLr/tUrRNrfuNT5CvvK1K`,
			expected: `mqBEVbNnL929CUA3sjkOmPB5dL0/a0spq8LgbIsJa22SfP580XduzUIKnCtdeC9TjPB/GEPp/LvEUFaLTUgPDQQGu3H5UCZyjVTAMHl45me/0qISEf903zFFqW5Lk3TS6iPrithqMMvhdK29Eg5OhhcoHS+ALpn0EjzUe86NywuFNb6ID4o8aF/ztZlKJegnpDAm3JuhCBauJ+0gcOB8GNdWd5a06qkokmwk1tgwWat7cQGFIH1NOvBwRMKhD51MJ7V28806a3zkOVwwhOiyyTXR+EcDA/aq5acX0yailLWB82g/2GR/DiaqNtusV+gpcMTNYemEv3c/xLkClJc29DSfTsJGKsmIDMqeBMM7RRBNinNAriY9iNX1UuHZLr/tUrRNrfuNT5CvvK1K`,
		},
		test{
			in:       `IMcfbWZ/iCa/LDcvMlk6LEJ0gDe4ohy2Vi0pVBd9aqR5PnRj8zGit8G2rLuNUkDmQ95bMURasmaPw2Xjf6SQjRk8coIHDLtbg/YNQVMabE8pKd6EaFdsGWJkcFoonxhPR29aH0xvjC4Mp3cJX3mjqyVsOp9xdk6d0Y2hzV3W/oPCq0DV03pm7P3+jH2OzoVVIDYgG1FD12S03otJrCXuzDmE2LOQ0xwgBQ9sREBLXwQzUKfXH8ogZzjdR19pX9qe0rRKMNz8k5lqcF9R2z+XIS1QAfeV9xopXA0CeyrhtoOkXV2i8kBxyodDp7tIeOvbEfvaqZGJgaJyV8UMTDi7zjwNeVdyKa8USH7zrXSoCl+Ud5eflI9vxKS+u9Bt1ufBHJtULOCHGA2vimkU`,
			expected: `IMcfbWZ/iCa/LDcvMlk6LEJ0gDe4ohy2Vi0pVBd9aqR5PnRj8zGit8G2rLuNUkDmQ95bMURasmaPw2Xjf6SQjRk8coIHDLtbg/YNQVMabE8pKd6EaFdsGWJkcFoonxhPR29aH0xvjC4Mp3cJX3mjqyVsOp9xdk6d0Y2hzV3W/oPCq0DV03pm7P3+jH2OzoVVIDYgG1FD12S03otJrCXuzDmE2LOQ0xwgBQ9sREBLXwQzUKfXH8ogZzjdR19pX9qe0rRKMNz8k5lqcF9R2z+XIS1QAfeV9xopXA0CeyrhtoOkXV2i8kBxyodDp7tIeOvbEfvaqZGJgaJyV8UMTDi7zjwNeVdyKa8USH7zrXSoCl+Ud5eflI9vxKS+u9Bt1ufBHJtULOCHGA2vimkU`,
		},
		test{
			in:       `AqC2sr44HVueGzgW13zHvJkqOEBWA8XA66ZEb3EoL1ehypSnJ07cFoWZlO8kf3k57L1fuHFWJ6quEdLXQaT9SJKHlUaYQvanvjbBlqWwaH3hODNsBGoK0DatpoQ+FxcSkdVE/ki3rbEUuJiZzU0BnDxH+Q6FiNsBaJuwau29w24MlD28ELJsjCcUVwtTQkaNtUxIlFKHLj0++T+IVrQH8KZlmVLvDefJ6llWbrFNVuh674HfKr/GEUatG6KI4gWNtGKKRYh76mMl5xH5qDfBZqxyRaKylJaDIYbx5xP5I4DDm4gOnxH+h/Pu6dq6FJ/U3eDio/KQ9xwFqTuyjH0BIRBsvWWgbTNURVBheq+am92YBhkj1QmdKTxQ9fQM55O8DpyWzRhky0NevM9j`,
			expected: `AqC2sr44HVueGzgW13zHvJkqOEBWA8XA66ZEb3EoL1ehypSnJ07cFoWZlO8kf3k57L1fuHFWJ6quEdLXQaT9SJKHlUaYQvanvjbBlqWwaH3hODNsBGoK0DatpoQ+FxcSkdVE/ki3rbEUuJiZzU0BnDxH+Q6FiNsBaJuwau29w24MlD28ELJsjCcUVwtTQkaNtUxIlFKHLj0++T+IVrQH8KZlmVLvDefJ6llWbrFNVuh674HfKr/GEUatG6KI4gWNtGKKRYh76mMl5xH5qDfBZqxyRaKylJaDIYbx5xP5I4DDm4gOnxH+h/Pu6dq6FJ/U3eDio/KQ9xwFqTuyjH0BIRBsvWWgbTNURVBheq+am92YBhkj1QmdKTxQ9fQM55O8DpyWzRhky0NevM9j`,
		},
		test{
			in:       `qkFfS3WfLyj3QTQT9i/s57uOPQCTN1jrab8bwxaxyeYUlz2tEtYyKGGUufua8WzdBT2VvWTvH0JkK0LfUJ+vChvcnMFna+tEaCKCFMIOWMLYVZSJDcYMIqaIr8d0Bi2bpbVf5z4WNma0pbCKaXpkYgeg1Sb8HpKG0p0fAez7Q/QRASlvyM5vuIOH8/CM4fF5Ga6aWkTRG0lfxiyeZ2vi3q7uNmsZF490J79r/6tnPPXIIC4XGnijwho5NmhZG0XcQeyW5KnT7VmGACFdTHOb9oS5WxZZU29/oZ5Y23rBBoSDX/xZ1LNFiZk6Xfl4ih207jzogv+3nOro93JHQydNeKEwxOtbKqEe7WWJLDw/EzVdJTODrhBYKbjUce10XsavuiTvv+H1Qh4lo2Vx`,
			expected: `qkFfS3WfLyj3QTQT9i/s57uOPQCTN1jrab8bwxaxyeYUlz2tEtYyKGGUufua8WzdBT2VvWTvH0JkK0LfUJ+vChvcnMFna+tEaCKCFMIOWMLYVZSJDcYMIqaIr8d0Bi2bpbVf5z4WNma0pbCKaXpkYgeg1Sb8HpKG0p0fAez7Q/QRASlvyM5vuIOH8/CM4fF5Ga6aWkTRG0lfxiyeZ2vi3q7uNmsZF490J79r/6tnPPXIIC4XGnijwho5NmhZG0XcQeyW5KnT7VmGACFdTHOb9oS5WxZZU29/oZ5Y23rBBoSDX/xZ1LNFiZk6Xfl4ih207jzogv+3nOro93JHQydNeKEwxOtbKqEe7WWJLDw/EzVdJTODrhBYKbjUce10XsavuiTvv+H1Qh4lo2Vx`,
		},
		test{
			in:       `O900/Gn82AjyLYqiWZ4ILXBBv/ZaXpTpQL0p9nv7gwF2MWsS2OWEImcVDa+1ElrjUumG6CVEv/rvax53krqJJDg+4Z/XcHxv58w6hNrXiWqFNjxlu5RZHvj1oQQXnS2n8qw8e/c+8ea2TiDIVr4OmgZz1G9uSPBeOZJvySqdgNPMpgfjZwkL2ez9/x31sLuQxi/FW3DFXU6kGSUjaq8g/iGXlaaAcQ0t9Gy+y005Z9wpr2JWWzishL+1JZp9D4SY/r3NHDphN4MNdLHMNBRPSIgfsaSqfLraIt+zWIycsd+nksVxtPv9wcyXy51E1qlHr6Uygz2VZYD9q9zyxEX4wRP2VEewHYUomL9d1F6gGG5fN3z82bQ4hI9uDirWhneWazUOQBRud5otPOm9`,
			expected: `O900/Gn82AjyLYqiWZ4ILXBBv/ZaXpTpQL0p9nv7gwF2MWsS2OWEImcVDa+1ElrjUumG6CVEv/rvax53krqJJDg+4Z/XcHxv58w6hNrXiWqFNjxlu5RZHvj1oQQXnS2n8qw8e/c+8ea2TiDIVr4OmgZz1G9uSPBeOZJvySqdgNPMpgfjZwkL2ez9/x31sLuQxi/FW3DFXU6kGSUjaq8g/iGXlaaAcQ0t9Gy+y005Z9wpr2JWWzishL+1JZp9D4SY/r3NHDphN4MNdLHMNBRPSIgfsaSqfLraIt+zWIycsd+nksVxtPv9wcyXy51E1qlHr6Uygz2VZYD9q9zyxEX4wRP2VEewHYUomL9d1F6gGG5fN3z82bQ4hI9uDirWhneWazUOQBRud5otPOm9`,
		},
		test{
			in:       `C3c+d5Q9lyTafPLdelG1TKaLFinw1TOjyI6KkrQyHKkttfnO58WFvScl1TiRcB/iHxKahskoE2+VRLUIhctuDU4sUvQh/g9Arw0LAA4QTxuLFt01XYdigurz4FT15ox2oDGGGrRb3VGjDTXK1OWVJoLMW95EVqyMc9F+Fdej85LHE+8WesIfacjUQtTG1tzYVQTfubZq0+qxXws8QrxMLFtVE38tbeXo+Ok1/U5TUa6FjWflEfvKY3XVcl8RKkXua7fVz/Blj8Gh+dWe2cOxa0lpM75ZHyz9adQrB2Pb4571E4u2xI5un0R0MFJZBQuPDc1G5rPhyk+Hb4LRG3dS0m8IASQUOskv93z978L1+Abu9CLP6d6s5p+BzWxhMUqwQXC/CCpTywrkJ0RG`,
			expected: `C3c+d5Q9lyTafPLdelG1TKaLFinw1TOjyI6KkrQyHKkttfnO58WFvScl1TiRcB/iHxKahskoE2+VRLUIhctuDU4sUvQh/g9Arw0LAA4QTxuLFt01XYdigurz4FT15ox2oDGGGrRb3VGjDTXK1OWVJoLMW95EVqyMc9F+Fdej85LHE+8WesIfacjUQtTG1tzYVQTfubZq0+qxXws8QrxMLFtVE38tbeXo+Ok1/U5TUa6FjWflEfvKY3XVcl8RKkXua7fVz/Blj8Gh+dWe2cOxa0lpM75ZHyz9adQrB2Pb4571E4u2xI5un0R0MFJZBQuPDc1G5rPhyk+Hb4LRG3dS0m8IASQUOskv93z978L1+Abu9CLP6d6s5p+BzWxhMUqwQXC/CCpTywrkJ0RG`,
		},
		// Basic XSS
		test{
			in:       `test<script>alert(document.cookie)</script>`,
			expected: `test`,
		},
		test{
			in:       `<<<><<script src=http://fake-evil.ru/test.js>`,
			expected: `&lt;&lt;&lt;&gt;&lt;`,
		},
		test{
			in:       `<script<script src=http://fake-evil.ru/test.js>>`,
			expected: ``,
		},
		test{
			in:       `<SCRIPT/XSS SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		test{
			in:       "<BODY onload!#$%&()*~+-_.,:;?@[/|\\]^`=alert(\"XSS\")>",
			expected: ``,
		},
		test{
			in:       `<BODY ONLOAD=alert('XSS')>`,
			expected: ``,
		},
		test{
			in:       `<iframe src=http://ha.ckers.org/scriptlet.html <`,
			expected: ``,
		},
		test{
			in:       `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');"">`,
			expected: `<input type="IMAGE">`,
		},
		test{
			in:       `<a onblur="alert(secret)" href="http://www.google.com">Google</a>`,
			expected: `<a href="http://www.google.com">Google</a>`,
		},
		// IMG attacks
		test{
			in:       `<img src="http://www.myspace.com/img.gif"/>`,
			expected: `<img src="http://www.myspace.com/img.gif"/>`,
		},
		test{
			in:       `<img src=javascript:alert(document.cookie)>`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;&#39;&#88;&#83;&#83;&#39;&#41;>`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC='&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041'>`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC="jav&#x0D;ascript:alert('XSS');">`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`,
			expected: ``,
		},
		test{
			in:       `<IMG SRC="javascript:alert('XSS')"`,
			expected: ``,
		},
		test{
			in:       `<IMG LOWSRC="javascript:alert('XSS')">`,
			expected: ``,
		},
		test{
			in:       `<BGSOUND SRC="javascript:alert('XSS');">`,
			expected: ``,
		},
		// HREF attacks
		test{
			in:       `<LINK REL="stylesheet" HREF="javascript:alert('XSS');">`,
			expected: ``,
		},
		test{
			in:       `<LINK REL="stylesheet" HREF="http://ha.ckers.org/xss.css">`,
			expected: ``,
		},
		test{
			in:       `<STYLE>@import'http://ha.ckers.org/xss.css';</STYLE>`,
			expected: ``,
		},
		test{
			in:       `<STYLE>BODY{-moz-binding:url("http://ha.ckers.org/xssmoz.xml#xss")}</STYLE>`,
			expected: ``,
		},
		test{
			in:       `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}</STYLE><UL><LI>XSS`,
			expected: `<ul><li>XSS`,
		},
		test{
			in:       `<IMG SRC='vbscript:msgbox("XSS")'>`,
			expected: ``,
		},
		test{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0; URL=http://;URL=javascript:alert('XSS');">`,
			expected: ``,
		},
		test{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=javascript:alert('XSS');">`,
			expected: ``,
		},
		test{
			in:       `<META HTTP-EQUIV="refresh" CONTENT="0;url=data:text/html;base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K">`,
			expected: ``,
		},
		test{
			in:       `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`,
			expected: ``,
		},
		test{
			in:       `<FRAMESET><FRAME SRC="javascript:alert('XSS');"></FRAMESET>`,
			expected: ``,
		},
		test{
			in:       `<TABLE BACKGROUND="javascript:alert('XSS')">`,
			expected: ``,
		},
		test{
			in:       `<TABLE><TD BACKGROUND="javascript:alert('XSS')">`,
			expected: `<td>`,
		},
		test{
			in:       `<DIV STYLE="background-image: url(javascript:alert('XSS'))">`,
			expected: `<div>`,
		},
		test{
			in:       `<DIV STYLE="width: expression(alert('XSS'));">`,
			expected: `<div>`,
		},
		test{
			in:       `<IMG STYLE="xss:expr/*XSS*/ession(alert('XSS'))">`,
			expected: ``,
		},
		test{
			in:       `<STYLE>@im\\port'\\ja\\vasc\\ript:alert("XSS")';</STYLE>`,
			expected: ``,
		},
		test{
			in:       `<BASE HREF="javascript:alert('XSS');//">`,
			expected: ``,
		},
		test{
			in:       `<BaSe hReF="http://arbitrary.com/">`,
			expected: ``,
		},
		test{
			in:       `<OBJECT TYPE="text/x-scriptlet" DATA="http://ha.ckers.org/scriptlet.html"></OBJECT>`,
			expected: ``,
		},
		test{
			in:       `<OBJECT classid=clsid:ae24fdae-03c6-11d1-8b76-0080c744f389><param name=url value=javascript:alert('XSS')></OBJECT>`,
			expected: ``,
		},
		test{
			in:       `<EMBED SRC="http://ha.ckers.org/xss.swf" AllowScriptAccess="always"></EMBED>`,
			expected: ``,
		},
		test{
			in:       `<EMBED SRC="data:image/svg+xml;base64,PHN2ZyB4bWxuczpzdmc9Imh0dH A6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcv MjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hs aW5rIiB2ZXJzaW9uPSIxLjAiIHg9IjAiIHk9IjAiIHdpZHRoPSIxOTQiIGhlaWdodD0iMjAw IiBpZD0ieHNzIj48c2NyaXB0IHR5cGU9InRleHQvZWNtYXNjcmlwdCI+YWxlcnQoIlh TUyIpOzwvc2NyaXB0Pjwvc3ZnPg==" type="image/svg+xml" AllowScriptAccess="always"></EMBED>`,
			expected: ``,
		},
		test{
			in:       `<SCRIPT a=">" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		test{
			in:       `<SCRIPT a=">" '' SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		test{
			in:       "<SCRIPT a=`>` SRC=\"http://ha.ckers.org/xss.js\"></SCRIPT>",
			expected: ``,
		},
		test{
			in:       `<SCRIPT a=">'>" SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: ``,
		},
		test{
			in:       `<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="http://ha.ckers.org/xss.js"></SCRIPT>`,
			expected: `PT SRC=&#34;http://ha.ckers.org/xss.js&#34;&gt;`,
		},
		test{
			in:       `<SCRIPT SRC=http://ha.ckers.org/xss.js`,
			expected: ``,
		},
		test{
			in:       `<div/style=&#92&#45&#92&#109&#111&#92&#122&#92&#45&#98&#92&#105&#92&#110&#100&#92&#105&#110&#92&#103:&#92&#117&#114&#108&#40&#47&#47&#98&#117&#115&#105&#110&#101&#115&#115&#92&#105&#92&#110&#102&#111&#46&#99&#111&#46&#117&#107&#92&#47&#108&#97&#98&#115&#92&#47&#120&#98&#108&#92&#47&#120&#98&#108&#92&#46&#120&#109&#108&#92&#35&#120&#115&#115&#41&>`,
			expected: `<div>`,
		},
		test{
			in:       `<a href='aim: &c:\\windows\\system32\\calc.exe' ini='C:\\Documents and Settings\\All Users\\Start Menu\\Programs\\Startup\\pwnd.bat'>`,
			expected: ``,
		},
		test{
			in:       `<!--\n<A href=\n- --><a href=javascript:alert:document.domain>test-->`,
			expected: `test--&gt;`,
		},
		test{
			in:       `<a></a style="xx:expr/**/ession(document.appendChild(document.createElement('script')).src='http://h4k.in/i.js')">`,
			expected: ``,
		},
		// CSS attacks
		test{
			in:       `<div style="position:absolute">`,
			expected: `<div>`,
		},
		test{
			in:       `<style>b { position:absolute }</style>`,
			expected: ``,
		},
		test{
			in:       `<div style="z-index:25">test</div>`,
			expected: `<div>test</div>`,
		},
		test{
			in:       `<style>z-index:25</style>`,
			expected: ``,
		},
		// Strings that cause issues for tokenizers
		test{
			in:       `<a - href="http://www.test.com">`,
			expected: `<a href="http://www.test.com">`,
		},
		// Comments
		test{
			in:       `text <!-- comment -->`,
			expected: `text `,
		},
		test{
			in:       `<div>text <!-- comment --></div>`,
			expected: `<div>text </div>`,
		},
		test{
			in:       `<div>text <!--[if IE]> comment <[endif]--></div>`,
			expected: `<div>text </div>`,
		},
		test{
			in:       `<div>text <!--[if IE]> <!--[if gte 6]> comment <[endif]--><[endif]--></div>`,
			expected: `<div>text &lt;[endif]--&gt;</div>`,
		},
		test{
			in:       `<div>text <!--[if IE]> <!-- IE specific --> comment <[endif]--></div>`,
			expected: `<div>text  comment &lt;[endif]--&gt;</div>`,
		},
		test{
			in:       `<div>text <!-- [ if lte 6 ]>\ncomment <[ endif\n]--></div>`,
			expected: `<div>text </div>`,
		},
		test{
			in:       `<div>text <![if !IE]> comment <![endif]></div>`,
			expected: `<div>text  comment </div>`,
		},
		test{
			in:       `<div>text <![ if !IE]> comment <![endif]></div>`,
			expected: `<div>text  comment </div>`,
		},
	}

	// These tests are run concurrently to enable the race detector to pick up
	// potential issues
	wg := sync.WaitGroup{}
	wg.Add(len(tests))
	for ii, tt := range tests {
		go func(ii int, tt test) {
			out, err := p.Sanitize(tt.in)
			if err != nil {
				t.Error(err)
			}
			if out != tt.expected {
				t.Errorf(
					"test %d failed;\ninput   : %s\noutput  : %s\nexpected: %s",
					ii,
					tt.in,
					out,
					tt.expected,
				)
			}
			wg.Done()
		}(ii, tt)
	}
	wg.Wait()
}
