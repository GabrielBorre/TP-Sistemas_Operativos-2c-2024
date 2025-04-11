package mmu

import (
	"fmt"
	//"io"
	"log"
	//"net/http"
	"strconv"
	"strings"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
)

/////////////////////////////// READ_MEM Y WRITE_MEM ///////////////////////////////

func Traducir_direccion_logica(direccion_logica uint32, tid uint32, pid uint32) (uint32, error) {

	//El codigo que comente no es necesario porque la base y el limite ya vienen dados en el contexto de ejecucion

	//Si la base + la dir logica (desplazamiento) es > al limite ----> Seg fault
	if direccion_logica+globals.Contexto_de_ejecucion.BASE > globals.Contexto_de_ejecucion.LIMITE {

		globals.Hubo_seg_fault = true

		cpu_api.Enviar_syscall_a_kernel("SEGMENTATION FAULT", nil)

		return 0, fmt.Errorf("Segmentation Fault: Direccion logica %d fuera del limite", direccion_logica)
		/*FALTA PARA EL CHK3*/
		//Llamar a Actualizar_contexto Y Notificar a Kernel con motivo de Segmentation Fault-------------------------
	}

	//calculamos la direcion fisica
	direccion_fisica := globals.Contexto_de_ejecucion.BASE + direccion_logica
	log.Printf("TID: %d - Traducción de dirección lógica %d a física %d", tid, direccion_logica, direccion_fisica)

	return direccion_fisica, nil
}

func Obtener_direcciones_fisicas(direcciones_logicas []uint32, tid uint32, pid uint32) []uint32 {

	var direcciones_fisicas []uint32

	for _, dir_logica := range direcciones_logicas {
		direccion_fisica, err := Traducir_direccion_logica(dir_logica, tid, pid)
		if err != nil {
			log.Printf("Error en la traducción: %s", err.Error())
			//notificar al kernel del SEG FAULT
			return nil
		}
		direcciones_fisicas = append(direcciones_fisicas, direccion_fisica)
	}

	return direcciones_fisicas
}

// ell data tiene que venir como "base,limite"
func Parsear_base_y_limite(data string) (uint32, uint32, error) {
	partes := strings.Split(data, ",")
	if len(partes) != 2 {
		return 0, 0, fmt.Errorf("Formato incorrecto: %s", data)
	}

	//pasamos la base y limite a enteros
	base, err := strconv.ParseUint(strings.TrimSpace(partes[0]), 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("Error al convertir la base: %v", err)
	}

	limite, err := strconv.ParseUint(strings.TrimSpace(partes[1]), 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("Error al convertir el limite: %v", err)
	}

	return uint32(base), uint32(limite), nil

}
