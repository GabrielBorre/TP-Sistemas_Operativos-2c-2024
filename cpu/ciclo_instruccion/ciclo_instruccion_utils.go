package ciclo_instruccion

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"io"
	"log"
	//"net/http"
	//"reflect"
	"strconv"
	//"strings"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	//mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	//"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

// ////////////////// PCB Y TCB RECUPERACION ////////////////////
func Recuperacion_pcb_y_tcb() {

	cpu_api.Pedir_pcb_a_memoria(globals.Tid_y_pid.PID)
	<-globals.Sem_b_recepcion_pcb
	cpu_api.Pedir_tcb_a_memoria(globals.Tid_y_pid.TID)
	<-globals.Sem_b_recepcion_tcb

	globals.Sem_b_recuperacion_completa <- true

}

//////////////////// CONVERSIONES ////////////////////

func Convertir_Uint32(parametro string) uint32 {
	parametro_convertido, err := strconv.Atoi(parametro)
	if err != nil {
		fmt.Print("Error al convertir el parametro a uint32")
	}
	return uint32(parametro_convertido)
}

type Uint interface{ ~uint32 }

func Convertir[T Uint](tipo string, parametro interface{}) T {

	if parametro == "" {
		log.Printf("La cadena de texto está vacía")
	}

	switch tipo {
	case "uint8":
		valor := parametro.(uint8)
		return T(valor)
	case "uint32":
		valor := parametro.(uint32)
		return T(valor)
	case "float64":
		valor := parametro.(float64)
		return T(valor)
	case "int":
		valor := parametro.(int)
		return T(valor)
	}
	return T(0)
}

//////////////////// TIMEOUT DE MEMORIA (VER SI VA O NO) ////////////////////
/*
	for {

		globals.MotivoDesalojoMutex.Lock()
		if pcb.FlagDesalojo {
			globals.MotivoDesalojoMutex.Unlock()
			break
		}
		globals.MotivoDesalojoMutex.Unlock()

		if globals.MemDelay > int(globals.CurrentTCB.Quantum) {
			globals.CurrentTCB.MotivoDesalojo = "TIMEOUT"
			pcb.FlagDesalojo = true
		}

		//ciclo_instruccion.Decode_and_execute(globals.CurrentTCB)

		//fmt.Println("Los registros de la cpu son", globals.CurrentTCB.Registros_CPU)

	}
*/
