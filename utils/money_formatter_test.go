package utils

import (
	"math"
	"testing"
)

func TestMoneyFormatter(t *testing.T) {
	want := "1c"
	formatted := ToDNDMoneyFormat(1)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "- 1c"
	formatted = ToDNDMoneyFormat(-1)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "0c"
	formatted = ToDNDMoneyFormat(0)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "0c"
	formatted = ToDNDMoneyFormat(-0)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "1p 1g 1s 1c"
	formatted = ToDNDMoneyFormat(1111)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "- 1p 1g 1s 1c"
	formatted = ToDNDMoneyFormat(-1111)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}

	want = "2147483p 6g 4s 7c"
	formatted = ToDNDMoneyFormat(math.MaxInt32)
	if want != formatted {
		t.Fatalf("ToDNDMoneyFormat(1) = [%s], want [%s]", formatted, want)
	}
}
