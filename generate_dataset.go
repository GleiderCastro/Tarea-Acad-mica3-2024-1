package main

//librerias que se utilizan
import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Generar datos para el dataset y guardarlos en un archivo CSV
func generateDataset(filename string, nSamples int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir encabezados
	header := []string{"ID", "Frecuencia", "GastoT", "DiasSinCompra", "VariedadDeProductos"}
	writer.Write(header)

	// Escribir datos aleatorios
	rand.Seed(time.Now().UnixNano())
	for i := 1; i <= nSamples; i++ {
		record := make([]string, 5)
		record[0] = strconv.Itoa(i)                         // ID
		record[1] = fmt.Sprintf("%.2f", rand.Float64()*100) // Frecuencia
		record[2] = fmt.Sprintf("%.2f", rand.Float64()*100) // GastoT
		record[3] = fmt.Sprintf("%.2f", rand.Float64()*100) // DiasSinCompra
		record[4] = fmt.Sprintf("%.2f", rand.Float64()*100) // VariedadDeProductos
		writer.Write(record)
	}

	return nil
}

func main() {
	filename := "dataset.csv"
	nSamples := 1000000 // Número de datos tal como pide el enunciado

	err := generateDataset(filename, nSamples)
	if err != nil {
		fmt.Println("Error al generar el dataset:", err)
		return
	}
	fmt.Println("Dataset generado correctamente en", filename)
}
