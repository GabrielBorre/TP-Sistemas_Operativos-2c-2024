package operaciones

import (
	"log"

	"encoding/binary"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	particiones "github.com/sisoputnfrba/tp-golang/memoria/particiones"
)

func Read_mem(direccion uint32, pid *int) uint32 {
	//log.Println("Vamos a leer memoria")
	particion := particiones.Buscar_particion(pid)

	//log.Printf("Paso despues de particion")

	if particion == nil {
		return 0
	}

	if direccion+3 >= uint32(len(globals.Memoria_usuario)) || direccion+3 > uint32(particion.Limite) {
		log.Println("Dirección de memoria fuera de rango")
		return 0
	}

	bytes_a_leer := make([]byte, 4)
	//busco los 4 bytes a leer y los cargo en el slice bytes_a_leer
	for i := 0; i < 4; i++ {
		bytes_a_leer[i] = globals.Memoria_usuario[int(direccion)+i]
	}

	respuesta := binary.BigEndian.Uint32(bytes_a_leer)
	// Asumimos big-endian

	//log.Printf(" Se leyó todo ok el valor: %d", respuesta)

	return respuesta
}

func Write_mem(direccion_fisica uint32, valor uint32, pid *int) uint32 {
	//log.Println("Vamos a escribir memoria")

	particion := particiones.Buscar_particion(pid)
	
	if particion == nil {
		log.Printf("El PID %d no tiene partición asignada", *pid)
		return 0
	}

	if direccion_fisica+3 >= uint32(len(globals.Memoria_usuario)) || direccion_fisica+3 > uint32(particion.Limite) {
		log.Println("Dirección de memoria fuera de rango")
		return 0
	}

	//transformo el valor uint32 que me llega a un slice de bytes de 4 posiciones
	bytes_del_valor_a_escribir := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes_del_valor_a_escribir, valor)

	//copio los 4 bytes del valor a escribir al slice de bytes global de la memoria
	copy(globals.Memoria_usuario[direccion_fisica:], bytes_del_valor_a_escribir)

	return valor
}
