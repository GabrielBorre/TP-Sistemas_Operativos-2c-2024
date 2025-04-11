package procesos

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	syscalls "github.com/sisoputnfrba/tp-golang/kernel/syscalls/utils_asincro"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
	
)

////////////////////////////////////////////////////////////////////
/////////////////////CREACION DE PROCESOS:FUNCIONES UTILES////////////
///////////////////////////////////////////////////////////////////

//funcion que crea un proceso con sus campos y lo devulve

// @annotation: esta funcion NO se tiene que usar por sí sola, usar Crear_proceso para crear procesos.
func Crear_pcb(tamanio uint32) *pcb.T_PCB {

	pcb_creado := &pcb.T_PCB{
		PID:           globals.Siguiente_PID,
		Estado:        "NEW",
		Tamanio:       tamanio,
		Ejecuciones:   0,
		Contador_TIDS: 0, //sirve para autoincrementar los TIDS del proceso y asignarlos
	}

	if globals.Proceso_actual == nil {
		//log.Println("LARGO PLAZO: Guardando en variable global al nuevo proceso")
		globals.Proceso_actual = &pcb.T_PCB{
			PID:           globals.Siguiente_PID,
			Estado:        "NEW",
			Tamanio:       tamanio,
			Ejecuciones:   0,
			Contador_TIDS: 0,
		}
	} else {
		globals.Proceso_actual.PID = globals.Siguiente_PID
		globals.Proceso_actual.Estado = "NEW"
		globals.Proceso_actual.Tamanio = tamanio
		globals.Proceso_actual.Ejecuciones = 0
		globals.Proceso_actual.Contador_TIDS = 0
	}

	color.Log_obligatorio("## (<PID>: %d) Se crea el proceso - Estado: NEW", pcb_creado.PID)
	slice.PushMutex(&globals.Lista_de_Procesos, pcb_creado, &globals.Sem_m_lista_procesos)

	//incrementamos la var global
	globals.Siguiente_PID = globals.Siguiente_PID + 1

	return pcb_creado
}

////////////////////////////////////////////////////////////////////
/////////////////////FINALIZACION DE PROCESOS:FUNCIONES UTILES////////////
///////////////////////////////////////////////////////////////////

func Process_exit() {
	//EN LA LISTA DE PROCESOS CREADOS BUSCO EL PROCESO CON EL PID QUE ME ENVIO LA CPU
	//globals.Sem_m_hilo_actual.Lock()
	//pcb := devolver_pcb(int(globals.Hilo_actual.PID))
	//globals.Sem_m_hilo_actual.Unlock()
	//Peticion_a_memoria_para_finalizar_proceso(pcb) //esta peticion se puede hacer en un hilo aparte
}

// //FUNCION QUE ELIMINA AL PCB QUE SOLICITO LA SYSCALLL PROCESS_EXIT
func EliminarPCB(pid uint32) {
	//obtengo el proceso asocido al hilo que recien salio de ejecutar a partir de su PID
	pcb_finalizado := syscalls.Devolver_pcb(int(pid))
	cantidad_hilos_del_pcb := len(pcb_finalizado.TCBs) //GUARDO EN UNA VARIABLE LA CANTIDAD DE HILOS DEL PCB A FINALIZAR
	for i := 0; i < cantidad_hilos_del_pcb; i++ {      //RECORRO TODOS LOS HILOS DEL PCB
		pcb_finalizado.TCBs[i].Estado = "EXIT"                                                //SETEO EL ESTADO DEL HILO A EXIT
		slice.PushMutex(&globals.Cola_Exit, pcb_finalizado.TCBs[i], &globals.Sem_m_cola_exit) //MANDO CADA HILO A LA COLA DE EXIT
		pcb_finalizado.TCBs[i].PCB = nil    
		                                                  //HAGO QUE CADA HILO DEJE DE APUNTAR AL PCB PARA QUE EL GARBAGE COLLECTOR LIBERE LA MEMORIA DEL PCB FINALIZADO                                            //MARCO
	}

	switch globals.Config_kernel.Algoritmo_planificacion {

	//saco los hilos asociados al proceso de la cola de Ready
	case "FIFO":
		for i := 0; i < len(globals.Cola_Ready); i++ {
			if globals.Cola_Ready[i].PID == pid {
				slice.RemoveAtIndexMutex(&globals.Cola_Ready, i, &globals.Sem_m_cola_ready)
			}
		}
	case "PRIORIDADES":
		for i := 0; i < len(globals.Cola_Ready); i++ {
			if globals.Cola_Ready[i].PID == pid {
				slice.RemoveAtIndexMutex(&globals.Cola_Ready, i, &globals.Sem_m_cola_ready)
			}
		}

	case "CMN":
		globals.Sem_m_lista_colas_multinivel.Lock()
		for i := 0; i < len(globals.Lista_colas_multinivel); i++ {
			for j := 0; j < len(globals.Lista_colas_multinivel[i].Cola); j++ {
				if globals.Lista_colas_multinivel[i].Cola[j].PID == pid {
					slice.RemoveAtIndex(&globals.Lista_colas_multinivel[i].Cola, j)
				}
			}
		}
		globals.Sem_m_lista_colas_multinivel.Unlock()

	}

	//saco a los hilos del pcb finalizados de la cola de blocked de los mutex
	for i := 0; i < len(pcb_finalizado.Mutexs); i++ {
		for j := 0; j < len(pcb_finalizado.Mutexs[i].Lista_hilos_bloqueados); j++ {
			if pcb_finalizado.Mutexs[i].Lista_hilos_bloqueados[j].PID == pid {
				slice.RemoveAtIndex(&pcb_finalizado.Mutexs[i].Lista_hilos_bloqueados, j)
			}
		}
	}

	/*
	// Eliminar el PCB del slice de PCBs totales (Hago esto para que el garbage collector libere la memoria del proceso finalizado)
	for i, pcb := range globals.Lista_de_Procesos {
		if pcb.PID == pcb_finalizado.PID {
			globals.Lista_de_Procesos[i] = nil // Eliminar la referencia en el slice
			break
		}
	}
	*/

	// Eliminar el PCB del slice de PCBs totales usando RemoveAtIndex
	for i, pcb := range globals.Lista_de_Procesos {
		if pcb.PID == pcb_finalizado.PID {
			// Usar RemoveAtIndex para eliminar el PCB
			removed := slice.RemoveAtIndex(&globals.Lista_de_Procesos, i)
			log.Printf("Se eliminó el PCB con PID %d: %+v", removed.PID, removed)
			break
		}
	}

	color.Log_obligatorio("## Finaliza el proceso <PID; %d >", pid)

	pcb_finalizado = nil

}
