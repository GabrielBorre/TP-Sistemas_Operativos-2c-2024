package entrada_salida

import (
	"log"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	syscalls_utils "github.com/sisoputnfrba/tp-golang/kernel/syscalls/utils_asincro"
	planning_utils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

//////////////////////////////////////////////////////////////////////
////////////////////////CREACION DE HILOS: FUNCIONES UTILES////////////////
////////////////////////////////////////////////////////////////////////

// ///Función que crea y devuelve un tcb con todos sus campos creados (debemos usarla cuando se solicite crear un hilo)
func Crear_tcb(pcb_creado *pcb.T_PCB, archivo_pseudocodigo string, prioridad_hilo int) *pcb.T_TCB {

	tcb_creado := &pcb.T_TCB{
		TID:           uint32(pcb_creado.Contador_TIDS),
		PCB:           pcb_creado,
		Instrucciones: archivo_pseudocodigo,
		Prioridad:     prioridad_hilo,
		Estado:        "NEW",
		PID:           pcb_creado.PID,
	}

	/*
		if globals.Hilo_actual == nil {
			globals.Hilo_actual = &pcb.T_TCB{
				TID:           uint32(pcb_creado.Contador_TIDS),
				PCB:           pcb_creado,
				Instrucciones: archivo_pseudocodigo,
				Prioridad:     prioridad_hilo,
				Estado:        "NEW",
				PID:           pcb_creado.PID,
			}
			tcb_creado = globals.Hilo_actual
		}
		*\
		/*el hilo actual no es necesariamente el que se acaba de crear. Por eso comente lo de abajo
		globals.Hilo_actual.TID = uint32(pcb_creado.Contador_TIDS)
		globals.Hilo_actual.PCB = pcb_creado
		globals.Hilo_actual.PID = pcb_creado.PID
		globals.Hilo_actual.Instrucciones = archivo_pseudocodigo
		globals.Hilo_actual.Prioridad = prioridad_hilo
		globals.Hilo_actual.Estado = "NEW"
	*/

	//SE AGREGA EL HILO AL SLICE DE HILOS DEL PROCESO: lo hacemos sin mutex porque en teoría la manipulación de esta cola es siempre SECUENCIAL (estuvimos 2hs analizando jaja)
	slice.Push(&pcb_creado.TCBs, tcb_creado)
	slice.Push(&globals.Lista_de_Hilos, tcb_creado) // agrego el nuevo hilo a la lista globals de hilos

	pcb_creado.Contador_TIDS = pcb_creado.Contador_TIDS + 1

	return tcb_creado
}

////Funcion que se encarga de implementar la logica

// Logica de la syscall thread_create completa (creamos)

func Crear_hilo(pcb_creado *pcb.T_PCB, archivo_pseudocodigo string, prioridad_hilo int) {
	//creo el hilo
	tcb := Crear_tcb(pcb_creado, archivo_pseudocodigo, prioridad_hilo)

	color.Log_obligatorio("## <PID: %d > <TID: %d > Se crea el hilo - Estado: Ready", tcb.PID, tcb.TID)

	log.Printf("El archivo de pseudocodigo es %s y la prioridad del hilo es %d ", tcb.Instrucciones, tcb.Prioridad)

	//le pedimos a memoria crear un hilo
	planning_utils.Pedido_a_memoria_para_crear_hilo(tcb)

	//una vez que lo crea, lo mandamos a la cola de Ready
	globals.Sem_tcb_enviar_a_cola_de_Ready <- tcb
	//esta peticion podria correr en un hilo aparte, al cual le deberiamos enviar una señal a traves de un semaforo
	//planning_utils.Pedido_a_memoria_para_crear_hilo(tcb)

}

//////////////////////////////////////////////////////////////////
///////////////////FINALIZACION DE HILOS: FUNCIONES UTILES/////////////
///////////////////////////////////////////////////////////////////

//////logica de la syscall THREAD_EXIT/////////

func Finalizar_hilo(tcb_a_finalizar *pcb.T_TCB) {

	globals.Sem_tcb_finalizar_hilo <- tcb_a_finalizar //le aviso a una rutina que le va a avisar a memoria que el hilo finalizo
	//log.Printf("Le envio a memoria la solicitud para que finalice el hilo TID %d PID %d ", tcb_a_finalizar.TID, tcb_a_finalizar.PID)
	//mando el hilo que finalizo al estado exit
	tcb_a_finalizar.Estado = "EXIT"

	//mando el hilo a la cola de exit
	slice.PushMutex(&globals.Cola_Exit, tcb_a_finalizar, &globals.Sem_m_cola_exit)

	//Recorro la lista de los hilos bloqueados por un join del tcb a finalizar y los mando a READY
	for i := 0; i < len(tcb_a_finalizar.Hilos_bloqueados_por_thread_join); i++ {

		//modifico el estado del hilo a Ready
		tcb_a_finalizar.Hilos_bloqueados_por_thread_join[i].Estado = "READY"
		//mando el hilo a la cola de Ready
		switch globals.Config_kernel.Algoritmo_planificacion {
		case "FIFO":
			slice.PushMutex(&globals.Cola_Ready, tcb_a_finalizar.Hilos_bloqueados_por_thread_join[i], &globals.Sem_m_cola_ready)
		case "PRIORIDADES":
			globals.Sem_m_cola_ready.Lock()
			planning_utils.Insertar_en_cola_ready_prioridades(tcb_a_finalizar.Hilos_bloqueados_por_thread_join[i])
			globals.Sem_m_cola_ready.Unlock()
		case "CMN":
			cola_multinivel := planning_utils.Buscar_cola_multinivel(tcb_a_finalizar.Hilos_bloqueados_por_thread_join[i].Prioridad)

			slice.Push(&cola_multinivel.Cola, tcb_a_finalizar.Hilos_bloqueados_por_thread_join[i])
		}

	}
	//me devuelve la lista de mutexs asignados que tiene el tcb a finalizar
	lista_mutex_asignados := syscalls_utils.Hallar_mutexes_del_TCB(tcb_a_finalizar)
	//busco en la lista de mutex los mutex que tenia asignado el hilo que finalizo
	for j := 0; j < len(lista_mutex_asignados); j++ {
		//marco que cada mutex que el hilo que finalizo tenia asignado se libero
		lista_mutex_asignados[j].Esta_asignado = false
		lista_mutex_asignados[j].Hilo_utilizador = nil
		//por cada mutex que tenia asignado el hilo que finalizo, mando a READY los hilos que cada mutex tenia bloqueado
		for k := 0; k < len(lista_mutex_asignados[j].Lista_hilos_bloqueados); k++ {
			//cambio el estado del hilo a READY
			lista_mutex_asignados[j].Lista_hilos_bloqueados[k].Estado = "READY"
			//mando a la cola de Ready el hilo
			switch globals.Config_kernel.Algoritmo_planificacion {
			case "FIFO":
				slice.PushMutex(&globals.Cola_Ready, lista_mutex_asignados[j].Lista_hilos_bloqueados[k], &globals.Sem_m_cola_ready)

			case "PRIORIDADES":
				planning_utils.Insertar_en_cola_ready_prioridades(lista_mutex_asignados[j].Lista_hilos_bloqueados[k])

			case "CMN":
				cola_multinivel := planning_utils.Buscar_cola_multinivel(lista_mutex_asignados[j].Lista_hilos_bloqueados[k].Prioridad)

				slice.Push(&cola_multinivel.Cola, lista_mutex_asignados[j].Lista_hilos_bloqueados[k])
			}
		}
	}

	color.Log_obligatorio("## <PID: %d > <TID: %d > Finaliza el hilo - Estado: Exit", tcb_a_finalizar.PID, tcb_a_finalizar.TID)

}

//////////////////LOGICA DE THREAD_JOIN///////////////////////////

func Thread_join(tid uint32, lista_de_hilos_del_proceso []*pcb.T_TCB) {
	//busco el tcb que le corresponde al tid pasado por parametro
	tcb := syscalls_utils.Devolver_tcb(tid, lista_de_hilos_del_proceso)

	//si el tcb es distinto de nil (o sea, se encuentra en la lista de hilos del pcb), bloqueamos el hilo que solicito la syscall en la lista de hilos bloqueados de ese tcb
	if tcb != nil {
		color.Log_resaltado(color.Blue, "El hilo <TID:%d> <PID:%d> QUIERE HACER UN JOIN DEL HILO <TID: %d>", globals.Hilo_actual.TID, globals.Hilo_actual.PID, tid)
		//(no se si es necesario) protejo la variable global hilo_actual (el hilo que esta ejecutando)
		globals.Sem_m_hilo_actual.Lock()
		defer globals.Sem_m_hilo_actual.Unlock()

		globals.Hilo_actual.Estado = "BLOCKED" //cambio el estado a BLOCKED

		//mando el hilo a la cola de blocked
		slice.PushMutex(&globals.Cola_Blocked, globals.Hilo_actual, &globals.Sem_m_cola_blocked)

		//agrego el hilo que solicito THREAD_JOIN () a la lista de hilos bloqueados del tid pasasdo por parametro
		slice.Push(&tcb.Hilos_bloqueados_por_thread_join, globals.Hilo_actual)

		color.Log_obligatorio("## <PID: %d > <TID: %d > - Bloqueado por THREAD_JOIN", globals.Hilo_actual.PID, globals.Hilo_actual.TID)

		//cancelo el quantum si el quantum ya estaba corriendo para algun hilo
		globals.Sem_m_round_robin_corriendo.Lock()
		if globals.Config_kernel.Algoritmo_planificacion == "CMN" && globals.Round_robin_corriendo {
			log.Printf("busca cancelar el round robin")
			globals.Sem_b_cancelar_Round_Robin <- true
		}
		globals.Sem_m_round_robin_corriendo.Unlock()

		syscalls_utils.Chequear_detener_planificacion()

		//replanificamos
		globals.Sem_b_plani_corto <- true

	} else { //si el tid pasado por parametro no existe, mando a ejecutar el hilo_actual
		log.Printf("El hilo <TID:%d> <PID:%d> QUIERE HACER UN JOIN DEL HILO QUE NO EXISTE EN SU PROCESO", globals.Hilo_actual.TID, globals.Hilo_actual.PID)
		kernel_api.Enviar_tid_a_ejecutar()

	}

}

//////////////////LOGICA DE THREAD_CANCEL//////////////

func Thread_cancel(tid uint32, lista_de_hilos_del_proceso []*pcb.T_TCB) {
	//obtengo el tcb a partir del tid que le pase por parametro
	tcb := syscalls_utils.Devolver_tcb(tid, lista_de_hilos_del_proceso)
	if tcb != nil { //si el tcb es distinto de nil (o sea, el tcb existe),finalizo ese tcb
		//busco en la cola de Ready si encuentro el hilo a terminar,y si es asi, lo saco de ahi
		syscalls_utils.Remuevo_de_la_cola_de_Ready(tcb)
		syscalls_utils.Remuevo_de_la_cola_de_Blocked(tcb)
		Finalizar_hilo(tcb) //finalizo el tcb al que le corresponde el tid pasado por parametro
	}

}
