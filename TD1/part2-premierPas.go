package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

func initGrille(m int, n int) [][]int {
	var grille = make([][]int, m)
	for i := 0; i < m; i++ {
		grille[i] = make([]int, n)
		for j := 0; j < n; j++ {
			grille[i][j] = rand.Intn(2)
		}
	}
	return grille
}

func compterVoisinsVivants(grille [][]int, i int, j int) int {
	var compteur = 0
	for x := i - 1; x <= i+1; x++ {
		for y := j - 1; y <= j+1; y++ {
			if x >= 0 && x < len(grille) && y >= 0 && y < len(grille[0]) && !(x == i && y == j) {
				compteur += grille[x][y]
			}
		}
	}
	return compteur
}

func update(grille [][]int) [][]int {
	var newGrille = make([][]int, len(grille))
	for i := 0; i < len(grille); i++ {
		newGrille[i] = make([]int, len(grille[0]))
		for j := 0; j < len(grille[0]); j++ {
			var nbVoisins = compterVoisinsVivants(grille, i, j)
			if grille[i][j] == 1 {
				if nbVoisins < 2 || nbVoisins > 3 {
					newGrille[i][j] = 0
				} else {
					newGrille[i][j] = 1
				}
			} else {
				if nbVoisins == 3 {
					newGrille[i][j] = 1
				} else {
					newGrille[i][j] = 0
				}
			}
		}
	}
	return newGrille
}

func afficherGrille(grille [][]int) {
	for i := 0; i < len(grille); i++ {
		for j := 0; j < len(grille[0]); j++ {
			if grille[i][j] == 1 {
				fmt.Print("■ ")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

func main() {
	var grille = initGrille(30, 50)
	for i := 0; i < 100; i++ {
		c := exec.Command("clear")
		c.Stdout = os.Stdout
		c.Run()
		afficherGrille(grille)
		grille = update(grille)
		time.Sleep(1 * time.Second)
	}
}
