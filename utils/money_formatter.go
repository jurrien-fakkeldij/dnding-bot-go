package utils

import (
	"fmt"
	"strings"
)

func ToDNDMoneyFormat(amount int) string {
	negative := false
	if amount < 0 {
		negative = true
		amount *= -1
	}

	platinum := amount / 1000
	gold := (amount / 100) % 10
	silver := (amount / 10) % 10
	copper := amount % 10

	showPlatinum := platinum != 0
	showGold := showPlatinum || gold != 0
	showSilver := showPlatinum || showGold || silver != 0

	output := strings.Builder{}

	if negative {
		output.WriteString("- ")
	}
	if showPlatinum {
		output.WriteString(fmt.Sprintf("%dp ", platinum))
	}

	if showGold {
		output.WriteString(fmt.Sprintf("%dg ", gold))
	}

	if showSilver {
		output.WriteString(fmt.Sprintf("%ds ", silver))
	}

	output.WriteString(fmt.Sprintf("%dc", copper))

	return output.String()
}
