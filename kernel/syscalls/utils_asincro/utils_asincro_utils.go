package syscalls

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

////////////////////////////////////////////////////////////////////////////////
//////////////////////////FUNCIONES AUXILIARES PARA ATENDER SYSCALLS////////////////////
///////////////////////////////////////////////////////////////////////////////////

// /Funcion que me devuelve la lista de mutexs que tiene asignado un hilo
func Hallar_mutexes_del_TCB(tcbBuscado *pcb.T_TCB) []*pcb.T_MUTEX {

	var lista_mutex_asignados []*pcb.T_MUTEX
	//Hallo el pcb en la lista de procesos creados
	pcb := Devolver_pcb(int(tcbBuscado.PID))

	for i := 0; i < len(pcb.Mutexs); i++ {
		if pcb.Mutexs[i].Hilo_utilizador == tcbBuscado {
			lista_mutex_asignados = append(lista_mutex_asignados, pcb.Mutexs[i])
		}
	}
	return lista_mutex_asignados
}

///funcion que me deuelve el pcb cuando le paso como parametro un PID

func Devolver_pcb(pid int) *pcb.T_PCB {
	//busco en la lista de proceso aquel que tiene el pid pasado por parametro
	globals.Sem_m_lista_procesos.Lock()
	defer globals.Sem_m_lista_procesos.Unlock()
	log.Println("Intento buscar el PCB en la Lista_de_procesos")
	for i := 0; i < len(globals.Lista_de_Procesos); i++ {
		if int(globals.Lista_de_Procesos[i].PID) == pid {
			return globals.Lista_de_Procesos[i]
		}
	}
	log.Println("El proceso a finalizar no existe")
	return nil

}

// funcion que busca en la lista global de hilos el tid pasado por parametro

//Si se encuentra el hilo, se devuelve el mismo

// Si no se encuentra el hilo, se devuelve nil
func Devolver_tcb(tid uint32, lista_hilos_de_proceso []*pcb.T_TCB) *pcb.T_TCB {
	for i := 0; i < len(lista_hilos_de_proceso); i++ {
		if lista_hilos_de_proceso[i].TID == tid {
			return lista_hilos_de_proceso[i]

		}
	}
	log.Printf("No encontre el TID : %d", tid)
	return nil
}

// Devuelve un booleano (true o false), dependiendo de si un hilo se encuentra en la cola de Ready
// esta funcion la uso en THREAD_CANCEL
func Remuevo_de_la_cola_de_Ready(tcb *pcb.T_TCB) {
	for i := 0; i < len(globals.Cola_Ready); i++ {
		if globals.Cola_Ready[i].PID == tcb.PID && globals.Cola_Ready[i].TID == tcb.TID {
			//saco de la cola de ready en la posicion "i" a un hilo que va a finalizar
			slice.RemoveAtIndexMutex(&globals.Cola_Ready, i, &globals.Sem_m_cola_ready)
		}
	}

}

func Remuevo_de_la_cola_de_Blocked(tcb *pcb.T_TCB) {
	globals.Sem_m_cola_blocked.Lock()
	for i := 0; i < len(globals.Cola_Blocked); i++ {
		if globals.Cola_Blocked[i].PID == tcb.PID && globals.Cola_Blocked[i].TID == tcb.TID {
			//saco de la cola de ready en la posicion "i" a un hilo que va a finalizar
			slice.RemoveAtIndex(&globals.Cola_Blocked, i)
		}
	}
	globals.Sem_m_cola_blocked.Unlock()
}

func Remuevo_de_la_cola_de_Blocked_del_mutex(tcb *pcb.T_TCB, mutex pcb.T_MUTEX) {
	for i := 0; i < len(mutex.Lista_hilos_bloqueados); i++ {
		if mutex.Lista_hilos_bloqueados[i].TID == tcb.TID && mutex.Lista_hilos_bloqueados[i].PID == tcb.PID {
			slice.RemoveAtIndex(&mutex.Lista_hilos_bloqueados, i)
		}
	}

}

// ////////////////////////////////////////////////////////////////
// /////////////////////FUNCIONES AUXILIARES PARA MUTEXES///////////
// ////////////////////////////////////////////////////////////////
func Primer_hilo_bloqueado_recurso(recurso *pcb.T_MUTEX) *pcb.T_TCB {
	return slice.Shift(&recurso.Lista_hilos_bloqueados)
}

func Recurso_existente(nombre_mutex string, proceso *pcb.T_PCB) *pcb.T_MUTEX {
	i := 0
	for i < len(proceso.Mutexs) && proceso.Mutexs[i].Nombre != nombre_mutex {
		i++
	}

	// si me pasé, entonces el recurso NO existe para ese proc
	if i >= len(proceso.Mutexs) {
		return nil
	}
	return proceso.Mutexs[i]
}
func Mutex_tomado(mutex *pcb.T_MUTEX) bool {
	return mutex.Esta_asignado || mutex.Hilo_utilizador != nil
}

func Chequear_detener_planificacion() {
	if globals.Sem_b_detener_planificacion {
		globals.Sem_b_planificacion_detenida <- true
		<-globals.Sem_b_esperar_compactacion
	}
}
