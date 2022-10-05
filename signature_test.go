package git

import (
	"fmt"
	"testing"
	"time"
)

func TestSignatureToString(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Moscow")
	checkFatal(t, err)

	sig := &Signature{
		Name:  "Alice",
		Email: "alice@example.com",
		When:  time.Date(2022, 8, 14, 11, 22, 33, 0, loc),
	}

	fmt.Println(sig.When.Unix())

	actual := sig.ToString(true)
	expected := "Alice <alice@example.com> 1660465353 +0300"

	compareStrings(t, actual, expected)
}
