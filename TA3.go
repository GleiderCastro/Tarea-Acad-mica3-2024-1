package main

//librerias que se utilizan (no se utiliza librerias externas)
import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// Struc de los vectores
type Vector struct {
	Data []float64
}

// La función squaredDistance sirve para calcular la distancia euclidiana entre
// los vectores
func squaredDistance(p1, p2 *Vector) float64 {
	var sum float64
	for i := 0; i < len(p1.Data); i++ {
		sum += (p1.Data[i] - p2.Data[i]) * (p1.Data[i] - p2.Data[i])
	}
	return sum
}

// La función initializeCentroids sirve para inicializar los centroides de manera
// aleatoria
func initializeCentroids(data [][]float64, k int) [][]float64 {
	nSamples := len(data)
	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		idx := rand.Intn(nSamples)
		centroids[i] = make([]float64, len(data[0]))
		copy(centroids[i], data[idx])
	}
	return centroids
}

// La función mean sirve para calcular la media de los vectores
func mean(vectors []*Vector) *Vector {
	n := len(vectors)
	meanVector := make([]float64, len(vectors[0].Data))
	for _, v := range vectors {
		for i := range v.Data {
			meanVector[i] += v.Data[i] / float64(n)
		}
	}
	return &Vector{meanVector}
}

// La función kMeans contiene el algoritmo de K-means
// en esta funcion se hace el uso de los canales para las asignaciones
// y actualizacion de centroides
// ademas se emplea el uso de goroutines para que calcules las distancias al centroide
// y envie los resultados a travez de los canales
func kMeans(data [][]float64, k int, maxIterations int) ([][]float64, []int) {
	nSamples := len(data)
	centroids := initializeCentroids(data, k)
	assignments := make([]int, nSamples)
	vectors := make([]*Vector, nSamples)
	for i := range data {
		vectors[i] = &Vector{data[i]}
	}

	for iter := 0; iter < maxIterations; iter++ {
		// Canales para asignaciones y actualización de centroides
		// el canal assign se utiliza para pasar las asignaciones de los puntos
		// a los centroides
		// y el canal update se utiliza para pasar las actualizaciones de los centroides
		// el canal done se utiliza para sincronizar la finalizacion de las operaciones
		// de actulizacion y asignacion
		assignChan := make(chan struct{ index, assignment int }, nSamples)
		updateChan := make(chan struct {
			index    int
			centroid []float64
		}, k)
		doneChan := make(chan bool)

		// Asignar puntos a los centroides más cercanos en paralelo
		go func() {
			var wg sync.WaitGroup
			wg.Add(nSamples)
			for i := 0; i < nSamples; i++ {
				go func(i int) {
					defer wg.Done()
					minDist := squaredDistance(vectors[i], &Vector{centroids[0]})
					assignment := 0
					for j := 1; j < k; j++ {
						dist := squaredDistance(vectors[i], &Vector{centroids[j]})
						if dist < minDist {
							minDist = dist
							assignment = j
						}
					}
					assignChan <- struct{ index, assignment int }{i, assignment}
				}(i)
			}
			wg.Wait()
			close(assignChan)
		}()

		go func() {
			for assignment := range assignChan {
				assignments[assignment.index] = assignment.assignment
			}
			doneChan <- true
		}()

		<-doneChan

		// Actualizar centroides en paralelo
		clusters := make([][]*Vector, k)
		for i := range clusters {
			clusters[i] = make([]*Vector, 0)
		}
		for i, idx := range assignments {
			clusters[idx] = append(clusters[idx], vectors[i])
		}

		go func() {
			var wg sync.WaitGroup
			wg.Add(k)
			for j := 0; j < k; j++ {
				go func(j int) {
					defer wg.Done()
					centroid := mean(clusters[j]).Data
					updateChan <- struct {
						index    int
						centroid []float64
					}{j, centroid}
				}(j)
			}
			wg.Wait()
			close(updateChan)
		}()

		go func() {
			for update := range updateChan {
				centroids[update.index] = update.centroid
			}
			doneChan <- true
		}()

		<-doneChan
	}

	return centroids, assignments
}

// Leer dataset desde archivo CSV
func readDataset(filename string) ([][]float64, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	data := make([][]float64, len(records)-1)
	ids := make([]string, len(records)-1)
	for i := 1; i < len(records); i++ {
		data[i-1] = make([]float64, len(records[i])-1)
		ids[i-1] = records[i][0]
		for j := 1; j < len(records[i]); j++ {
			data[i-1][j-1], err = strconv.ParseFloat(records[i][j], 64)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return data, ids, nil
}

// En esta funcion se crea el csv resultante pudiendo ver a que datos esta asignado a que cluster

func saveResults(filename string, ids []string, data [][]float64, assignments []int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"ID", "Frecuencia", "GastoT", "DiasSinCompra", "VariedadDeProductos", "Cluster"}
	writer.Write(header)

	for i := range ids {
		record := make([]string, len(data[i])+2)
		record[0] = ids[i]
		for j := range data[i] {
			record[j+1] = fmt.Sprintf("%.2f", data[i][j])
		}
		record[len(data[i])+1] = strconv.Itoa(assignments[i])
		writer.Write(record)
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	inputFilename := "dataset.csv"     //leemos el data set (1000000 de datos tal como pide el enunciado)
	outputFilename := "resultados.csv" // escribimos el dataset
	k := 3                             // Número de clusters
	maxIterations := 100               // Número de iteraciones

	data, ids, err := readDataset(inputFilename)
	if err != nil {
		fmt.Println("Error al leer el dataset:", err)
		return
	}

	start := time.Now()
	centroids, assignments := kMeans(data, k, maxIterations)
	elapsed := time.Since(start)
	// se imprimen los centroides calculados
	fmt.Println("Centroids:")
	for _, c := range centroids {
		fmt.Println(c)
	}
	// se imprime el tiempo transcurrido para obetenr mas
	fmt.Printf("Tiempo transcurrido: %s\n", elapsed)
	// los datos se guardan el el data set resultante
	err = saveResults(outputFilename, ids, data, assignments)
	if err != nil {
		fmt.Println("Error al guardar los resultados:", err)
		return
	}
	fmt.Println("Resultados guardados correctamente en", outputFilename)
}
