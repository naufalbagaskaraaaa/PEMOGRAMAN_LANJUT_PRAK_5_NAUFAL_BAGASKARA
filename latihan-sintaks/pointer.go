package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func passByValue(x int) {
	x = 100
}

func passByPointer(x *int) {
	*x = 100
}

func main() {
	fmt.Println("SWAP POINTER")
	a := 10
	b := 20

	fmt.Println("Sebelum swap:", a, b)

	swap(&a, &b)
	fmt.Println("Sesudah swap:", a, b)

	fmt.Println("UPDATE SLICE")

	nama := []string{"Sari", "Budi"}
	fmt.Println("Sebelum update:", nama)

	updateSlice(&nama, "Andi")
	fmt.Println("Sesudah update:", nama)
	
	fmt.Println("PASS BY VALUE")

	nilai1 := 10
	fmt.Println("Sebelum function:", nilai1)

	passByValue(nilai1)
	fmt.Println("Sesudah function:", nilai1)

	fmt.Println("PASS BY POINTER")

	nilai2 := 10
	fmt.Println("Sebelum function:", nilai2)

	passByPointer(&nilai2)
	fmt.Println("Sesudah function:", nilai2)
}
