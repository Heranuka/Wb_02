package main

import "fmt"

func main() {
	cont := Constructor([][]int{
		{3, 0, 1, 4, 2},
		{5, 6, 3, 2, 1},
		{1, 2, 0, 1, 5},
		{4, 1, 0, 1, 7},
		{1, 0, 3, 0, 5},
	})
	fmt.Println(cont.SumRegion(2, 1, 4, 3))
}

type NumMatrix struct {
	numP [][]int
}

func Constructor(matrix [][]int) NumMatrix {
	return NumMatrix{
		numP: matrix,
	}
}
func isPalindrome(s string) bool {
l, r := 0, len(s)-1
for l < r {
    for l < r && !som(rune(s[l])){
    l++
    }
    for r > l && !som(rune(s[r])){
        r--
    }

    if unicode.ToLower(rune(s[l])) != unicode.ToLower(rune(s[r])) {
        return false
    }
    l++
    r--
}
return true
}


func som(s rune) bool {
    return unicode.IsLetter(s) || unicode.IsDigit{
    
}

//Если рейс авиакомпании "Победа" был задержан, и из-за этого вы не смогли улететь другим самолётом,
//Если вы покупали билеты у разных авиакомпаний и опоздали на стыковочный рейс из-за задержки первого рейса, то ответственность несет первая авиакомпания. Она должна связаться со второй авиакомпанией, чтобы решить вашу проблему.
//В случае, если вы не согласны с действиями авиакомпании, вы можете подать претензию в письменной форме и, при необходимости, обратиться в суд. 