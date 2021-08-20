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

func expectError(t *testing.T, err error, message string) {
	if err != nil {
		if err.Error() == message {
			return
		}
		t.Fatalf("expected error with message `%s`, but got `%s`", message, err.Error())
	}
	t.Fatalf("expected error but not")
}

func TestParseIllust(t *testing.T) {
	value, err := parseIllustId("92065303")
	assertNoError(t, err)
	assertEqual(t, value, 92065303)
	value, err = parseIllustId("https://www.pixiv.net/artworks/92065303")
	assertNoError(t, err)
	assertEqual(t, value, 92065303)
	value, err = parseIllustId("https://www.pixiv.net/member_illust.php?mode=medium&illust_id=92065303")
	assertNoError(t, err)
	assertEqual(t, value, 92065303)
	_, err = parseIllustId("http://www.pixiv.net")
	expectError(t, err, "not a pixiv link")
	_, err = parseIllustId("https://www.pixiv.net/invalid")
	expectError(t, err, "not a illust link")
	_, err = parseIllustUrl(" ")
	expectError(t, err, "not a pixiv link")
}

func TestIsAscii(t *testing.T) {
	assertEqual(t, isAscii("background"), true)
	assertEqual(t, isAscii("风景"), false)
}
