package utils

import "fmt"

func UniqueCodeGeneration(codeIdentifier string, id uint) string {
	code := codeIdentifier + "-"
	idStr := fmt.Sprint(id)
	idLength := len(idStr)
	amountOfZeroes := 4 - idLength
	for amountOfZeroes != 0 {
		code += "0"
		amountOfZeroes--
	}

	code += idStr
	return code
}
