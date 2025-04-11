package mutex

import (
	"log"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	syscalls_hilos "github.com/sisoputnfrba/tp-golang/kernel/syscalls/hilos"
	syscalls_utils "github.com/sisoputnfrba/tp-golang/kernel/syscalls/utils_asincro"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

// crea un mutex para el proceso SIN ASIGNARLO a ningun hilo particular
func Mutex_create(proceso *pcb.T_PCB, nombre string) {
	nuevo_mutex := &pcb.T_MUTEX{
		Nombre:          nombre,
		Esta_asignado:   false,
		Hilo_utilizador: nil,
	}
	log.Printf(color.Blue+"El hilo <TID :%d> del PID <%d> creo el MUTEX llamado %s "+color.Reset, globals.Hilo_actual.TID, globals.Hilo_actual.PID, nombre)
	//lo pusheo a la cola de mutex del proceso
	slice.Push(&proceso.Mutexs, nuevo_mutex)
}

// lockea un mutex o recurso
func Mutex_lock(nombre_mutex string, proceso *pcb.T_PCB, hilo *pcb.T_TCB) {
	posible_recurso := syscalls_utils.Recurso_existente(nombre_mutex, proceso)

	if posible_recurso == nil {
		log.Printf("NO SE ENCONTRO EL MUTEX LLAMADO %s SOLICITADO POR EL HILO TID: %d> <PID:%d>", nombre_mutex, hilo.TID, hilo.PID)

		//aviso que debo replanificar y finalizo el hilo

		if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
			globals.Sem_b_cancelar_Round_Robin <- true
		}

		syscalls_utils.Chequear_detener_planificacion()

		globals.Sem_b_plani_corto <- true

		syscalls_hilos.Finalizar_hilo(hilo)
		return // corto la ejecucion de la funcion
	} else if syscalls_utils.Mutex_tomado(posible_recurso) {
		log.Printf("EL hilo <TID : %d> solicito una MUTEX_LOCK de un mutex TOMADO por un hilo", hilo.TID)
		// si está tomado, bloqueo el hilo
		//globals.Sem_b_cancelar_Round_Robin <- true
		slice.Push(&posible_recurso.Lista_hilos_bloqueados, hilo)
		slice.PushMutex(&globals.Cola_Blocked, hilo, &globals.Sem_m_cola_blocked)
		color.Log_obligatorio("## (<PID: %d > : <TID: %d > -Bloqueado por MUTEX llamado %s", hilo.PID, hilo.TID, nombre_mutex)
		//Si el hilo se bloquea, debo replanificar

		if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
			globals.Sem_b_cancelar_Round_Robin <- true
		}

		syscalls_utils.Chequear_detener_planificacion()
		globals.Sem_b_plani_corto <- true
		return
	} else {
		log.Printf("SE ENCONTRO EL MUTEX LLAMADO %s SOLICITADO POR EL HILO TID: %d> <PID:%d>. SE ASIGNA EL MUTEX A DICHO HILO", nombre_mutex, hilo.TID, hilo.PID)
		posible_recurso.Hilo_utilizador = hilo
		posible_recurso.Esta_asignado = true
		//Se vuelve a mandar a ejecutar al hilo
		kernel_api.Enviar_tid_a_ejecutar()
	}
}

func Mutex_unlock(nombre_mutex string, proceso *pcb.T_PCB, hilo *pcb.T_TCB) {
	posible_recurso := syscalls_utils.Recurso_existente(nombre_mutex, proceso)
	if posible_recurso == nil {
		log.Printf("NO SE ENCONTRO EL MUTEX LLAMADO %s SOLICITADO POR EL HILO TID: %d> <PID:%d>", nombre_mutex, hilo.TID, hilo.PID)
		//aviso que hay que replanificar y al mismo tiempo finalizo el hilo

		if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
			globals.Sem_b_cancelar_Round_Robin <- true
		}

		syscalls_utils.Chequear_detener_planificacion()

		globals.Sem_b_plani_corto <- true

		syscalls_hilos.Finalizar_hilo(hilo)
		return
	} else if syscalls_utils.Mutex_tomado(posible_recurso) && posible_recurso.Hilo_utilizador != hilo {
		log.Printf("EL MUTEX %s, pero NO POR EL HILO <TID %d> : PID <%d> ", nombre_mutex, hilo.TID, hilo.PID)

		kernel_api.Enviar_tid_a_ejecutar()
		// si está tomado pero no es él, no hago nada
		return
	} else if len(posible_recurso.Lista_hilos_bloqueados) == 0 {
		log.Printf("SE UNLOCKEO EL MUTEX LLAMADO %s DEL <TID:%d > y <PID:%d>. NO HAY HILOS EN LA COLA DE BLOCKED QUE PASEN A SER DESBLOQUEADOS", nombre_mutex, hilo.TID, hilo.PID)

		posible_recurso.Hilo_utilizador = nil
		posible_recurso.Esta_asignado = false

		kernel_api.Enviar_tid_a_ejecutar()

		return
	} else {
		log.Printf("SE UNLOCKEA EL MUTEX LLAMADO %s del hilo <TID:%d> del <PID:%d>", nombre_mutex, hilo.TID, hilo.PID)

		//obtenemos el primer tcb de la cola de blocked del MUTEX
		posible_recurso.Hilo_utilizador = syscalls_utils.Primer_hilo_bloqueado_recurso(posible_recurso)

		//marcamos el hilo como asignado
		posible_recurso.Esta_asignado = true
		//mandamos ese tcb a la cola de Ready

		log.Printf("Se desbloqueo el hilo TID %d PID %d por el MUTEX_UNLOCK del recurso %s. El hilo pasa a Ready ", posible_recurso.Hilo_utilizador.TID, posible_recurso.Hilo_utilizador.PID, posible_recurso.Nombre)

		globals.Sem_tcb_enviar_a_cola_de_Ready <- posible_recurso.Hilo_utilizador

		kernel_api.Enviar_tid_a_ejecutar()

	}
}
