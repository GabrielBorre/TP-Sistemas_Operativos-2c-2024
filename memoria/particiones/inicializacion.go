package particiones

import (
	"log"

	"sort"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
)

func Declarar_memoria() {
	globals.Memoria_usuario = make([]byte, globals.Config_memoria.Memory_size)
}

func Aplicar_esquema() {
	esquema := globals.Config_memoria.Schema
	estrategia := globals.Config_memoria.Search_algorithm

	color.Log_resaltado(color.Yellow, "Se seleccionó el esquema: %s, bajo la estrategia: %s", esquema, estrategia)
	//var particion *Particion

	switch esquema {
	case "FIJAS":
		log.Printf("Particionando memoria según PARTICIONES FIJAS")
		if globals.Config_memoria.Search_algorithm == "FIRST" || globals.Config_memoria.Search_algorithm == "BEST" {
			Inicializar_particiones_fijas()
		} else if globals.Config_memoria.Search_algorithm == "WORST" {
			Inicializar_particiones_fijas_worst()
		}

		//particion = buscarParticion(tamanio, estrategia) -- Creo que esta lógica de buscar iría en otro lado (?), o al menos en otro momento

	case "DINAMICAS":
		log.Printf("Particionando memoria según PARTICIONES DINÁMICAS")
		Inicializar_particiones_dinamicas()
		//particion = buscarParticion(tamanio, estrategia)
	}
	/*
		if particion != nil {
			asignarParticion(particion, tamanio)
			log.Printf("Partición asignada: %+v", particion)
		} else {

			log.Printf("No se encontró una partición adecuada para el tamaño solicitado.")
		}
	*/
}

func Inicializar_particiones_dinamicas() {
	// se inicializa una única partición del total de la memoria y se va subdividiendo A MEDIDA QUE LLEGUEN LOS PEDIDOS DE CREACIÓN
	var tamanio_memoria int = globals.Config_memoria.Memory_size

	particion := &globals.Particion{
		Base:    0,
		Limite:  tamanio_memoria - 1, // - 1 porque sino desborda
		Tamanio: tamanio_memoria,
		Libre:   true,
	}
	globals.Particiones = append(globals.Particiones, particion)
}

func Inicializar_particiones_fijas() { // chequear
	var index int = 0

	for i := 0; i < len(globals.Config_memoria.Partitions); i++ {

		var tamanio_particion int = globals.Config_memoria.Partitions[i]

		particion := &globals.Particion{
			Base:    index,                         // index representa el byte en el que empieza la particion si lo ponemos en Memoria_usuario
			Limite:  index + tamanio_particion - 1, // - 1 porque sino ocupo espacio del siguiente; ej: base = 0, tamanio = 512 => limite = 511
			Tamanio: tamanio_particion,
			Libre:   true,
		}
		globals.Particiones = append(globals.Particiones, particion)
		index += tamanio_particion // sig iteración: base = 512, tamanio 512 => limite = 1024 - 1 = 1023
	}
}

func Inicializar_particiones_fijas_worst() {
	// Ordenar las particiones de mayor a menor tamaño
	sort.Slice(globals.Config_memoria.Partitions, func(i, j int) bool {
		return globals.Config_memoria.Partitions[i] > globals.Config_memoria.Partitions[j]
	})

	log.Printf("%+v", globals.Config_memoria.Partitions)

	var index int = 0

	for i := 0; i < len(globals.Config_memoria.Partitions); i++ {
		var tamanio_particion int = globals.Config_memoria.Partitions[i]

		particion := &globals.Particion{
			Base:    index,                         // index representa el byte en el que empieza la partición
			Limite:  index + tamanio_particion - 1, // -1 para evitar superposiciones
			Tamanio: tamanio_particion,
			Libre:   true,
		}
		globals.Particiones = append(globals.Particiones, particion)

		index += tamanio_particion // Actualizamos el índice base para la siguiente partición
	}
	for i := 0; i < len(globals.Particiones); i++ {
		log.Printf("%d", globals.Particiones[i].Tamanio)
	}
}
