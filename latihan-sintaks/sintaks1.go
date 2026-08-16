package main

import "fmt"

func main() {
	var nama string = "naufal"
	var umur int = 21
	var ipk float32 = 3.75
	var status bool = true
	var matkuliah []string = []string{"A", "B", "C"}

	fmt.Println("nama:", nama)
	fmt.Println("umur:", umur)
	fmt.Println("ipk:", ipk)
	fmt.Println("status:", status)
	fmt.Println("matkuliah:", matkuliah)

	mahasiswa := map[string]string{
		"naufal": "A",
		"budi":   "B",
		"andi":   "C",
	}

	fmt.Println("add data baru ke map")
	mahasiswa["naufal"] = "A"
	fmt.Println(mahasiswa)

	value, x := mahasiswa["naufal"]
	if x {
		fmt.Println("status: ", value)
	} else {
		fmt.Println("gaada")
	}

	fmt.Println("mencari data dalam map")
	value_1, x := mahasiswa["bagaskara"]
	if x {
		fmt.Println("status: ", value_1)
	} else {
		fmt.Println("gaada")
	}

	fmt.Println("menghapus data dalam map")
	delete(mahasiswa, "naufal")
	fmt.Println(mahasiswa)

	fmt.Println("searching data dalam map")
	for nama, matkuliah := range mahasiswa {
		fmt.Println("nama:", nama, "matkuliah:", matkuliah)
	}
}