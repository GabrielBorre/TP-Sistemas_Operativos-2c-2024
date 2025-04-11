package ciclo_instruccion

import (
	//"bytes"
	//"encoding/json"
	//"fmt"
	//"io"
	"log"
	//"net/http"
	//"reflect"
	//"strconv"
	//"strings"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	//mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	//"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

//////////////////// INTERRUPCIONES ////////////////////
/*
Check Interrupt
En este momento, se deberá chequear si el Kernel nos envió una interrupción al TID
que se está ejecutando, en caso afirmativo, se actualiza el Contexto de Ejecución en la Memoria y
se devuelve el TID al Kernel con motivo de la interrupción. Caso contrario, se descarta la interrupción.
*/

// se comporta como las syscalls o el seg_fault
func Check_interrupt() {

	// Chequeo en la cola de interrupciones (variable global) si hay alguna

	if len(globals.Lista_interrupciones) > 0 {

		//si en envio de la interrupcion es al TID y PID que esta ejecutando
		if globals.CurrentTCB.PID == globals.Tid_y_pid.PID && globals.CurrentTCB.TID == globals.Tid_y_pid.TID {
			log.Printf("CHECK INTERRUPT: hay interrupcion")
			globals.Sem_m_Lista_interrupciones_mutex.Lock()
			interrupcion := globals.Lista_interrupciones[0]
			globals.Lista_interrupciones = globals.Lista_interrupciones[1:] // Eliminar la interrupción de la cola
			globals.Sem_m_Lista_interrupciones_mutex.Unlock()

			aumentar_PC()
			hubo_interrupcion = true
			cpu_api.Actualizar_contexto()

			cpu_api.Enviar_syscall_a_kernel(interrupcion.InterruptionReason, nil)

		} else {
			globals.Sem_m_Lista_interrupciones_mutex.Lock()
			globals.Lista_interrupciones = globals.Lista_interrupciones[1:] // Eliminar la interrupción de la cola
			globals.Sem_m_Lista_interrupciones_mutex.Unlock()
		}

	} else {
		log.Printf("No hubo interrupciones")
	}
}
