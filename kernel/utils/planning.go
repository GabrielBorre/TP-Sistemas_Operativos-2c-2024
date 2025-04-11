package utils

import (
	"fmt"
	"log"

	/*
		Para el RR
		"io"
		"os"
		"time"
	*/

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"

	"github.com/sisoputnfrba/tp-golang/utils/slice"

	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
)

////////////////////////////////// VARIABLES LOCALES //////////////////////////////////

type T_Quantum struct {
	TimeExpired chan bool
}

// //////////////////////////////// PLANIFICACION //////////////////////////////////
// Esta funcion amerita cambios
func Largo_plazo() {

	// Hardcodeamos algunos procesos/instrucciones para el largo plazo
	//Crear_proceso_inicial("/home/utnso/tp-2024-2c-GOlazo/pruebas/archivo_pseudo", 1, 0)
	//log.Printf("LARGO PLAZO: Creamos el proceso inicial, Esperando nuevo proceso")

	for {

		<-globals.Sem_b_plani_largo

		//Cuando kernel reciba syscall de CPU en "kernel_api_cpu" ya va a haber sido insertada en la lista de largo plazo
		log.Printf("LARGO PLAZO: Llega nuevo proceso")

		// Sacamos el sigueinte proceso de la lista de largo plazo (NEW)
		pcb_nuevo := slice.ShiftMutex(&globals.Lista_plani_largo, &globals.Sem_m_plani_largo)

		//Nos fijamos que no sea el 1er proceso
		if pcb_nuevo.PID != 0 {

			//globals.Sem_c_grado_multiprog <- int(pcb_nuevo.PID)
			// Aumentamos grado multiprog (dps vemos si sacamos)

			//FALTA API Y SEMAFORO DE CPU CON ARCHIVO Y TAMANIO
			var archivoPseudoCodigo string
			var tamanio uint32
			var prioridad uint32
			Crear_proceso(archivoPseudoCodigo, tamanio, prioridad) //si ya sacamos el proceso de la cola de new, por qué volvemos a crear otro?
		}

		log.Printf("LARGO PLAZO: Esperando nuevo proceso")

	}
}

func Corto_plazo() {
	for {
		<-globals.Sem_b_plani_corto

		color.Log_resaltado(color.BgGreen, "Algoritmo de planificacion seleccionado: %s", globals.Config_kernel.Algoritmo_planificacion)
		//log.Println("CORTO PLAZO: Esperando a que me avisen que hay un proceso")

		switch globals.Config_kernel.Algoritmo_planificacion {

		case "FIFO":

			// Esperamos a que le llegue algo y empezamos
			//log.Println("CORTO PLAZO: Entramos al FIFO")
			Algoritmo_FIFO()

		case "PRIORIDADES":
			//log.Println("Algoritmo de planificacion seleccionado: PRIORIDADES CON DESALOJO")

			// Esperamos a que le llegue algo y empezamos
			//log.Println("CORTO PLAZO: Entramos al algoritmo de prioridades")

			Algoritmo_PRIORIDADES()
		case "CMN":

			// Esperamos a que le llegue algo y empezamos
			//log.Println("Algoritmo de planificacion seleccionado: COLAS MULTINIVEL")
			//log.Println("CORTO PLAZO: Entramos al algoritmo de Colas Multinivel")

			Algoritmo_COLAS_MULTINIVEL()

		default:
			fmt.Println("CORTO PLAZO: Algoritmo  de planificacion inexistente")
		}
	}
}

/*
	ADENTRO DE CORTO PLAZO CUANDO QUERAMOS PROBAR DETENER PLANIFICACION
	// Si esta detenida activa sus semaforos y espera la llamada
	if globals.Estado_planificacion == "DETENIDA" {
		globals.Sem_b_plani_corto <- true
		<-globals.Sem_b_plani_corto
		continue
	}
*/

////////////////////////////////// ALGORITMOS //////////////////////////////////

func Algoritmo_FIFO() {

	// Agarramos de la cola
	log.Println("CORTO PLAZO: Agarramos de la cola de ready al 1er hilo")

	globals.Sem_m_cola_ready.Lock()
	//si la cola de Ready tiene al menos un hilo esperando, lo mando a ejecutar
	if len(globals.Cola_Ready) > 0 {
		globals.Sem_m_cola_ready.Unlock()
		globals.Hilo_actual = slice.ShiftMutex(&globals.Cola_Ready, &globals.Sem_m_cola_ready)

		// Cambiamos a EXEC
		globals.Cambiar_estado_hilo(globals.Hilo_actual, "EXEC")

		globals.Cambiar_estado_proceso(globals.Proceso_actual, "EXEC")

		//Enviamos el tid con su pid a ejecutar a la cpu
		kernel_api.Enviar_tid_a_ejecutar()

		// Esperamos a que CPU diga que lo recibio para comenzar a esperar
		//<-globals.Sem_b_tcb_y_pid_recibido

		log.Println("CORTO PLAZO: Paso el semaforo, empezamos a esperar las interrupciones")
		//Cuando la CPU ya recibio el PID y TID podemos empezar a esperar a las interrupciones
		//EvictionManagement() //me cerraria mas que esto este en una API en donde se gestione el motivo de desalojo y en funcion de eso se hagan cosas
	} else {
		globals.Sem_m_cola_ready.Unlock()
		globals.Hilo_actual = nil
		log.Printf("No hay elementos en la cola de Ready para mandar a ejecutar")

	}

}

func Algoritmo_PRIORIDADES() {
	globals.Sem_m_cola_ready.Lock()
	if len(globals.Cola_Ready) > 0 {
		globals.Sem_m_cola_ready.Unlock()
		// Agarramos de la cola
		log.Println("CORTO PLAZO: Agarramos de la cola de ready al hilo de mayor prioridad")

		globals.Sem_m_hilo_actual.Lock()
		globals.Hilo_actual = slice.ShiftMutex(&globals.Cola_Ready, &globals.Sem_m_cola_ready)

		// Cambiamos a EXEC
		globals.Cambiar_estado_hilo(globals.Hilo_actual, "EXEC")

		globals.Cambiar_estado_proceso(globals.Proceso_actual, "EXEC")

		globals.Sem_m_hilo_actual.Unlock()
		//globals.Proceso_actual.Ejecuciones = globals.Proceso_actual.Ejecuciones + 1
		//log.Println("CORTO PLAZO: Aumentamos ejecuciones de proceso")

		//Enviamos el tid con su pid a ejecutar a la cpu
		kernel_api.Enviar_tid_a_ejecutar()

	} else {
		globals.Sem_m_cola_ready.Unlock()
		globals.Hilo_actual = nil
		log.Printf("No hay elementos en la cola de Ready para mandar a ejecutar")
	}
	// Esperamos a que CPU diga que lo recibio para comenzar a esperar
	//<-globals.Sem_b_tcb_y_pid_recibido

	//EvictionManagement() //me cerraria mas que esto este en una API en donde se gestione el motivo de desalojo y en funcion de eso se hagan cosas (se active el semaforo de corto plazo si hay que replanificar, etc)
	log.Println("CORTO PLAZO: Paso el semaforo, empezamos a esperar las interrupciones")

}

func Algoritmo_COLAS_MULTINIVEL() {
	//obtengo la cola de mayor prioridad que tenga hilos esperando para ejecutar
	if Alguna_cola_multinivel_tiene_hilos() {
		cola_de_mayor_prioridad := Buscar_cola_de_mayor_prioridad_y_con_tcbs()

		///Mando a ejecutar el hilo de la cola de mayor prioridad que tenga

		globals.Sem_m_hilo_actual.Lock()
		//Obtengo el primer hilo de la cola de mayor prioridad que va a ser el que ejecute en CPU
		globals.Hilo_actual = slice.Shift(&cola_de_mayor_prioridad.Cola)
		globals.Cambiar_estado_hilo(globals.Hilo_actual, "EXEC")
		globals.Sem_m_hilo_actual.Unlock()
		kernel_api.Enviar_tid_a_ejecutar()

		globals.Sem_m_round_robin_corriendo.Lock()
		if globals.Config_kernel.Algoritmo_planificacion == "CMN" && !globals.Round_robin_corriendo {
			log.Printf("Se activa el RR para el hilo TID %d del PID %d ", globals.Hilo_actual.TID, globals.Hilo_actual.PID)
			globals.Sem_b_iniciar_round_Robin <- true ///le aviso al RR que tiene que empezar a contar el quantum
		}
		globals.Sem_m_round_robin_corriendo.Unlock()

	} else {
		log.Printf("No existen hilos en ninguna de las colas multinivel")
		globals.Hilo_actual = nil
	}
}

func Alguna_cola_multinivel_tiene_hilos() bool {
	globals.Sem_m_lista_colas_multinivel.Lock()
	defer globals.Sem_m_lista_colas_multinivel.Unlock()
	for i := 0; i < len(globals.Lista_colas_multinivel); i++ {
		if len(globals.Lista_colas_multinivel[i].Cola) > 0 {
			return true
		}
	}
	return false

}

// funcion que me sirve para encontrar la cola de mayor prioridad de la listas de colas multinivel
func Buscar_cola_de_mayor_prioridad_y_con_tcbs() *pcb.T_Cola_Multinivel {
	var cola_de_mayor_prioridad *pcb.T_Cola_Multinivel = nil
	//semaforos para proteger la lista de colas multinivel (variable global)
	globals.Sem_m_lista_colas_multinivel.Lock()
	defer globals.Sem_m_lista_colas_multinivel.Unlock()
	//recorro la lista de colas multinivel
	for i := 0; i < len(globals.Lista_colas_multinivel); i++ {
		//si la cola es la primera, va a ser la de mayor prioridad hasta el momento. Si no lo es, comparo si la siguiente cola tiene mayor prioridad que la que encontre
		if cola_de_mayor_prioridad == nil || cola_de_mayor_prioridad.Prioridad > globals.Lista_colas_multinivel[i].Prioridad {
			//si despues de la comparacion, la cola tiene al menos un tcb, me quedo con esa cola como la de mayor prioridad y vuelvo a comparar
			if len(globals.Lista_colas_multinivel[i].Cola) > 0 {
				cola_de_mayor_prioridad = globals.Lista_colas_multinivel[i]
			}
		}

	}
	return cola_de_mayor_prioridad
}
