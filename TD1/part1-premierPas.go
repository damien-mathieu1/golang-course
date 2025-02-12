package main

import (
	"fmt"
	"math"
	"math/rand"
)

func estBissextile(annee int) bool {
	return (annee%4 == 0 && annee%100 != 0) || annee%400 == 0
}

func estPremier(nombre int) bool {
	for i := 2; i <= int(math.Sqrt(float64(nombre))); i++ {
		if nombre%i == 0 {
			return false
		}
	}
	fmt.Println(nombre)
	return true
}

func premiersNombrePremiers(n int) []int {
	var slice = make([]int, 0)
	for i := 1; i < n; i++ {
		if estPremier(i) {
			slice = append(slice, i)
		}
	}
	return slice
}

func genererTableauAleatoire(length int) []int {
	var array = make([]int, 0)

	for i := 0; i < length; i++ {
		array = append(array, rand.Intn(100))
	}

	return array
}

func triBulle(tableau []int) []int {
	for i := 0; i < len(tableau)-1; i++ {
		for j := 0; j < len(tableau)-i-1; j++ {
			if tableau[j] > tableau[j+1] {
				tableau[j], tableau[j+1] = tableau[j+1], tableau[j]
			}
		}
	}
	return tableau
}

func triSelection(tableau []int) []int {
	for i := 0; i < len(tableau)-1; i++ {
		min := i
		for j := i + 1; j < len(tableau); j++ {
			if tableau[j] < tableau[min] {
				min = j
			}
		}
		if min != i {
			tableau[i], tableau[min] = tableau[min], tableau[i]
		}
	}
	return tableau
}

func rechercheDichotomique(tableau []int, value int) (bool, int) {
	start, end := 0, len(tableau)-1
	for start <= end {
		mid := (start + end) / 2
		if tableau[mid] == value {
			return true, mid
		} else if tableau[mid] < value {
			start = mid + 1
		} else {
			end = mid - 1
		}
	}
	return false, -1

}

func organiserParTaille(tableau []string) []string {
	lengthMap := make(map[int][]string)
	for _, str := range tableau {
		length := len(str)
		lengthMap[length] = append(lengthMap[length], str)
	}

	var result []string
	for _, group := range lengthMap {
		result = append(result, group...)
	}

	return result
}

func main() {
	fmt.Println(estBissextile(2020))

	fmt.Println(estPremier(11))

	fmt.Println(premiersNombrePremiers(12))

	fmt.Println(genererTableauAleatoire(10))

	fmt.Println(triBulle(genererTableauAleatoire(10)))

	fmt.Println(triSelection(genererTableauAleatoire(10)))

	fmt.Println(rechercheDichotomique([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 11))

	fmt.Println(organiserParTaille([]string{"a", "aa", "aaa", "aaaa", "aaaaa", "aaaaaa", "aaaaaaa", "aaaaaaaa"}))
}
