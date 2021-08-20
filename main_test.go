package main

import "testing"

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestParseIllust(t *testing.T) {
	value, err := parseIllustId("92065303")
	assertNoError(t, err)
	assertEqual(t, value, 92065303)
	value, err = parseIllustId("https://www.pixiv.net/artworks/92065303")
	assertNoError(t, err)
	assertEqual(t, value, 92065303)
}

func TestIsAscii(t *testing.T) {
	assertEqual(t, isAscii("background"), true)
	assertEqual(t, isAscii("风景"), false)
}
