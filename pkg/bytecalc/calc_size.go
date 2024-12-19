package bytecalc

import "fmt"

var Sizes = [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

func CalculateSizeLiteral(bytesCount int) string {
	step := 0
	count := float64(bytesCount)
	for step < len(Sizes)-1 {
		divided := count / 1024
		if divided < 1 {
			return fmt.Sprintf("%.2f%s", count, Sizes[step])
		}

		count = divided
		step++
	}

	return fmt.Sprintf("%.2f%s", count, Sizes[len(Sizes)-1])
}
