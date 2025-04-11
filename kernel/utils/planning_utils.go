package utils

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	slice "github.com/sisoputnfrba/tp-golang/utils/slice"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////// PLANIFICACION DE LARGO PLAZO //////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////// CREACION DE PROCESOS //////////////////////////////////

// esta funcion es un pasamanos
func Crear_proceso_inicial(archivoPseudoCodigo string, tamanio uint32, prioridad uint32) {
	Crear_proceso(archivoPseudoCodigo, tamanio, prioridad)
}

// @annotation: esta funcion NO se tiene que usar por sí sola, usar Crear_proceso para crear procesos.
func Crear_pcb(tamanio uint32, archivo_pseudoCodigo string, prioridad_hilo_main uint32) *pcb.T_PCB {

	pcb_creado := &pcb.T_PCB{
		PID:           globals.Siguiente_PID,
		Estado:        "NEW",
		Tamanio:       tamanio,
		Ejecuciones:   0,
		Contador_TIDS: 0,
	}

	if globals.Proceso_actual == nil {
		log.Println("LARGO PLAZO: Guardando en variable global al nuevo proceso")
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

	tcb_principal := Crear_tcb(pcb_creado, archivo_pseudoCodigo, int(prioridad_hilo_main))
	color.Log_resaltado(color.Green, "## (<PID>: %d) Se crea el proceso - Estado: NEW", tcb_principal.PID)
	slice.PushMutex(&globals.Lista_de_Procesos, pcb_creado, &globals.Sem_m_lista_procesos)

	//incrementamos la var global
	globals.Siguiente_PID++

	return pcb_creado
}

func Crear_tcb(pcb_creado *pcb.T_PCB, archivo_pseudocodigo string, prioridad_hilo int) *pcb.T_TCB {

	tcb_creado := &pcb.T_TCB{
		TID:           uint32(pcb_creado.Contador_TIDS),
		Instrucciones: archivo_pseudocodigo,
		Prioridad:     prioridad_hilo,
		Estado:        "NEW",
		PID:           pcb_creado.PID,
	}

	//SE AGREGA EL HILO AL SLICE DE HILOS DEL PROCESO: lo hacemos sin mutex porque en teoría la manipulación de
	//esta cola es siempre SECUENCIAL (estuvimos 2hs analizando jaja)

	slice.Push(&pcb_creado.TCBs, tcb_creado)

	pcb_creado.Contador_TIDS = pcb_creado.Contador_TIDS + 1

	return tcb_creado

}

// función para crear un Proceso (va a activar un semaforo de la funcion pedido_a_memoria_para_inicializacion)
func Crear_proceso(archivo_pseudocodigo string, tamanio uint32, prioridad uint32) {

	//CREO EL PCB
	pcb_creado := Crear_pcb(tamanio, archivo_pseudocodigo, prioridad)

	log.Printf("el pcb creado tiene PID: %d ", pcb_creado.PID)

	log.Printf("El tamaño del proceso creado es %d", pcb_creado.Tamanio)

	log.Printf("La prioridad del hilo main del proceso creado es %d ", pcb_creado.TCBs[0].Prioridad)

	//tcb_principal := Crear_tcb(pcb_creado, archivo_pseudocodigo, 0)
	/*
		Revisar cuando estemos codeando PROCESS_CREATE
	*/

	// cantidad de procesos en new antes de agregar el proceso creado
	cantidad_procesos_new := len(globals.Lista_plani_largo)

	// SE AGREGA EL PROCESO A LA COLA DE NEW
	slice.PushMutex(&globals.Lista_plani_largo, pcb_creado, &globals.Sem_m_cola_new)

	// se chequea si la cola estaba vacía, si lo estaba se manda a inicializar de una
	if cantidad_procesos_new == 0 {
		globals.Sem_pcb_pedido_a_memoria_para_inicializar_proceso <- pcb_creado
	}
}

func Pedido_a_memoria_para_inicializacion() {
	for {

		//semaforo que activa el pedido a memoria para que inicialice un proceso
		pcb := <-globals.Sem_pcb_pedido_a_memoria_para_inicializar_proceso
		tcb_principal := pcb.TCBs[0]
		API.Peticion_a_memoria(pcb, tcb_principal)
		//esperamos a que nos confirme la recepcion de la peticion
		inicializo_proceso := <-globals.Sem_b_inicializar_pcb

		if inicializo_proceso {
			//Son dos peticiones distintas la de pedir crear un proceso en memoria y pedir crear un hilo.
			//Si la peticion de crear un hilo resulta exitosa, lo mando a READY (se podria hacer con un semaforo y no con un IF)
			API.Peticion_a_memoria_para_crear_hilo(tcb_principal)

			globals.Sem_m_cola_new.Lock()

			if len(globals.Lista_plani_largo) > 0 {

				globals.Lista_plani_largo = globals.Lista_plani_largo[1:] // SACO EL PRIMER ELEMENTO DE LA COLA DE NEW y lo desprecio

				globals.Sem_m_cola_new.Unlock()

				///Semaforo que va a activar al hilo enviarColaDeReady
				globals.Sem_tcb_enviar_a_cola_de_Ready <- tcb_principal
				///EnviarAColaDeReady(tcb_principal) esto se deberia ejecutar en un hilo aparte por modularidad

			}
		}
	}
}

////Funcion que sirve para insertar un hilo en una cola de Ready ordenada por la prioridad de los hilos

func Insertar_en_cola_ready_prioridades(nuevoTCB *pcb.T_TCB) {
	// encuentra la posición donde insertar el nuevo TCB
	pos := 0

	log.Printf("Cantidad de hilos en la cola de Ready: %d", len(globals.Cola_Ready))

	for pos < len(globals.Cola_Ready) && globals.Cola_Ready[pos].Prioridad <= nuevoTCB.Prioridad {
		pos++

	}
	// inserta el nuevo TCB en la posición encontrada
	// globals.Cola_Ready = append(globals.Cola_Ready[:pos], append([]*pcb.T_TCB{nuevoTCB}, globals.Cola_Ready[pos:]...)...)
	// tenemos la funcion InsertAtIndex:
	color.Log_resaltado(color.LightBlue, "Voy a insertar el hilo en la posicion %d", pos)
	slice.InsertAtIndex(&globals.Cola_Ready, pos, nuevoTCB)

	log.Printf("Se inserto el hilo TID %d del PID %d en la posicion %d ", nuevoTCB.TID, nuevoTCB.PID, pos)
}

func Enviar_a_cola_de_ready() {
	for {

		tcb := <-globals.Sem_tcb_enviar_a_cola_de_Ready
		algoritmo := globals.Config_kernel.Algoritmo_planificacion
		//mando el hilo recibido a la lista global de Hilos creados
		//slice.PushMutex(&globals.Lista_de_Hilos, tcb, &globals.Sem_m_lista_hilos)
		switch algoritmo {
		case "FIFO":

			slice.PushMutex(&globals.Cola_Ready, tcb, &globals.Sem_m_cola_ready)

			tcb.Estado = "READY"
			//log.Printf("## (PID:<%d> TID:<%d>) Se crea el Hilo - Estado: Ready", tcb.PCB.PID, tcb.TID)

			// Avisamos a semáforo de corto plazo

			//entra en el if cuando se manda a ejecutar por PRIMERA VEZ al TID 0 del PID 0. Despues no entra mas
			if globals.Primera_planificacion || globals.Hilo_actual == nil { //si el tcb a planificar es el hilo main del primer proceso, lo mando a ejecutar (activo el canal de corto plazo)
				//log.Printf("Avisamos a semáforo de corto plazo")
				globals.Primera_planificacion = false //lo pongo en false para que solo entre cuando se planifique por primera vez
				globals.Sem_b_plani_corto <- true
			}

			// Le mandamos el TID al contador de corto plazo para que empiece
			//globals.Sem_c_plani_corto_plazo <- int(tcb.TID)

			//Avisamos al de largo plazo
			//globals.Sem_b_plani_largo <- true

		case "PRIORIDADES":
			color.Log_obligatorio("## <PID : %d > <TID: %d > Pasa a la cola de Ready", tcb.PID, tcb.TID)
			tcb.Estado = "READY"

			//si la prioridad del hilo que entra a la cola de Ready es mayor (mientras menor sea el numero de prioridad, mayor sera la prioridad del hilo) que la del hilo que esta ejecutando, mandamos una señal interrupt a la CPU
			globals.Sem_m_hilo_actual.Lock()
			if globals.Hilo_actual != nil && tcb.Prioridad < globals.Hilo_actual.Prioridad {
				globals.Sem_m_hilo_actual.Unlock()
				globals.Sem_m_cola_ready.Lock()

				//quedaria en la primera posicion de la cola de Ready
				//pero por un tiempo muy acotado porque inmediatamente
				//se tiene que desalojar el hilo en ejecucion para poner a este ultimo hilo a ejecutar
				globals.Cola_Ready = append([]*pcb.T_TCB{tcb}, globals.Cola_Ready...)
				log.Printf("EL hilo TID %d PID %d se inserto en la primera posicion de la cola de Ready por tener mayor prioridad que el hilo que esta ejecutando ", globals.Cola_Ready[0].TID, globals.Cola_Ready[0].PID)
				globals.Sem_m_cola_ready.Unlock()

				globals.Sem_string_interrumpir_hilo <- "PRIORIDADES"
				//esperamos a semaforo que nos dice que se desalojo el proceso con exito
				//activamos el semaforo de corto plazo para avisar que hay que mandar este hilo a ejecutar

			} else {
				globals.Sem_m_hilo_actual.Unlock()
				tcb.Estado = "READY"

				//mando el hilo a la posicion que le corresponda en funcion de su prioridad (ver definicion de la funcion)
				Insertar_en_cola_ready_prioridades(tcb) //la funcion incluye el mutex correspondiente a la cola de Ready
				if globals.Primera_planificacion || globals.Hilo_actual == nil {
					globals.Primera_planificacion = false
					globals.Sem_b_plani_corto <- true
				}
			}

			// FIFO CON SORTBY !

		case "CMN":
			color.Log_obligatorio("## <PID : %d > <TID: %d > a cola multinivel - Estado: Ready", tcb.PID, tcb.TID)
			tcb.Estado = "READY"
			//SI ES COLA MULTINIVEL DEBEMOS AGREGAR ESE TCB A LA COLA QUE CORRESPONDA A SU PRIORIDAD

			//Si no existe la cola multinivel de proceso que entra a Ready, la creo
			if !existe_cola_multinivel(tcb.Prioridad) {
				//le agrego a la prioridad a la cola creada
				cola_multinivel := &pcb.T_Cola_Multinivel{
					Prioridad: tcb.Prioridad,
				}
				//agrego el hilo a la cola correspondiente a su prioridad
				slice.Push(&cola_multinivel.Cola, tcb)

				log.Printf("Se crea la cola y agrega el hilo <TID %d> <PID %d> a la Cola Multinivel de prioridad %d ", tcb.TID, tcb.PID, tcb.Prioridad)

				// agrego la cola a la lista de colas multinivel
				slice.PushMutex(&globals.Lista_colas_multinivel, cola_multinivel, &globals.Sem_m_lista_colas_multinivel)

			} else {
				//busco la cola multinivel que le corresponde a la prioridad del hilo
				var cola_multinivel *pcb.T_Cola_Multinivel = Buscar_cola_multinivel(tcb.Prioridad)
				slice.Push(&cola_multinivel.Cola, tcb)

				log.Printf("Se agrega el hilo <TID %d> <PID %d> a la Cola multinivel de prioridad %d", tcb.TID, tcb.PID, tcb.Prioridad)
			}
			globals.Sem_m_hilo_actual.Lock()
			if globals.Hilo_actual != nil && tcb.Prioridad < globals.Hilo_actual.Prioridad {
				globals.Sem_m_hilo_actual.Unlock()
				globals.Sem_string_interrumpir_hilo <- "PRIORIDADES"
				//mismo que con prioridades (mandamos una señal a la funcion que manda interrupciones para que el hilo actual en ejecucion se desaloje )

			} else if globals.Primera_planificacion || globals.Hilo_actual == nil {
				globals.Sem_m_hilo_actual.Unlock()
				globals.Primera_planificacion = false
				//log.Printf("Aviso la necesidad de planificar el TID %d del PID %d", tcb.TID, tcb.PID)
				globals.Sem_b_plani_corto <- true

			} else {
				globals.Sem_m_hilo_actual.Unlock()
			}

			/// falta logica

			//avisamos con semaforo al corto plazo
			/*
				log.Printf("Avisamos a semaforo de corto plazo")
				globals.Sem_b_plani_corto <- true

				//Avisamos al de largo plazo
				globals.Sem_b_plani_largo <- true
			*/
		}
	}
}

///////////////////////////////////////////////////////////////////
///////////////////////////FUNCIONES AUXILIARES///////////////////
//////////////////////////////////////////////////////////////////

////Funcion que recorre la lista de colas multiniveles y me dice si existe esa cola con el nivel de prioridad pasado por parametro

func existe_cola_multinivel(prioridad int) bool {
	globals.Sem_m_lista_colas_multinivel.Lock()
	defer globals.Sem_m_lista_colas_multinivel.Unlock()
	for i := 0; i < len(globals.Lista_colas_multinivel); i++ {
		if globals.Lista_colas_multinivel[i].Prioridad == prioridad {
			return true
		}

	}
	return false
}

///Funcion que me devuelve la cola multinivel que corresponde a la prioridad pasada por parametro

func Buscar_cola_multinivel(prioridad int) *pcb.T_Cola_Multinivel {
	globals.Sem_m_lista_colas_multinivel.Lock()
	defer globals.Sem_m_lista_colas_multinivel.Unlock()
	for i := 0; i < len(globals.Lista_colas_multinivel); i++ {
		if globals.Lista_colas_multinivel[i].Prioridad == prioridad {
			return globals.Lista_colas_multinivel[i]
		}
	}
	return nil
}

////////////SYSCALL: THREAD_CREATE/////////////////////////////////

// Logica de la syscall thread_create completa (creamos)
func Crear_hilo(pcb_creado *pcb.T_PCB, archivo_pseudocodigo string, prioridad_hilo int) {
	tcb := Crear_tcb(pcb_creado, archivo_pseudocodigo, prioridad_hilo)
	//esta peticion podria correr en un hilo aparte, al cual le deberiamos enviar una señal a traves de un semaforo
	Pedido_a_memoria_para_crear_hilo(tcb)
}

// Si la memoria nos permite crear un hilo, lo mandamos a la cola de READY
func Pedido_a_memoria_para_crear_hilo(tcb *pcb.T_TCB) {
	API.Peticion_a_memoria_para_crear_hilo(tcb)

}

///////////SYSCALL: THREAD_EXIT y THREAD_CANCEL//////////////

// /Funcio que me devuelve la lista de mutexs que tiene asignado un hilo
func FiltrarPorTCB(tcbBuscado *pcb.T_TCB) []*pcb.T_MUTEX {
	var lista_mutex_asignados []*pcb.T_MUTEX

	globals.Sem_m_lista_mutex.Lock()
	for _, mutex := range globals.Lista_mutexs {
		if mutex.Hilo_utilizador == tcbBuscado {
			lista_mutex_asignados = append(lista_mutex_asignados, mutex)
		}
	}
	globals.Sem_m_lista_mutex.Unlock()
	return lista_mutex_asignados
}

// crea un mutex para el proceso SIN ASIGNARLO a ningun hilo particular
func Mutex_create(proceso *pcb.T_PCB, nombre string) {
	nuevo_mutex := &pcb.T_MUTEX{
		Nombre:          nombre,
		Esta_asignado:   false,
		Hilo_utilizador: nil,
	}

	//lo pusheo a la cola de mutex del proceso
	slice.PushMutex(&proceso.Mutexs, nuevo_mutex, &globals.Sem_m_lista_mutex) // esta bien el globals.Sem_m_lista_mutex?
}

// lockea un mutex o recurso
func Mutex_lock(nombre_mutex string, proceso *pcb.T_PCB, hilo *pcb.T_TCB) {
	posible_recurso := Recurso_existente(nombre_mutex, proceso)

	if posible_recurso == nil {
		//necesito Finalizar_hilo() pero esa funcion recibe el hilo a finalizar desde la API, no por param
		return // corto la ejecucion de la funcion
	} else if Mutex_tomado(posible_recurso) {
		// si está tomado, bloqueo el hilo
		slice.Push(&posible_recurso.Lista_hilos_bloqueados, hilo) // CREEEEEEO que no hace falta sincronizar esta lista
		return
	} else {
		posible_recurso.Hilo_utilizador = hilo
		posible_recurso.Esta_asignado = true
	}
}

func Recurso_existente(nombre_mutex string, proceso *pcb.T_PCB) *pcb.T_MUTEX {
	i := 0
	for i < len(proceso.Mutexs) && proceso.Mutexs[i].Nombre != nombre_mutex {
		i++
	}

	// si me pasé, entonces el recurso NO existe para ese proc
	if i > len(proceso.Mutexs) {
		return nil
	}
	return proceso.Mutexs[i]
}

func Mutex_tomado(mutex *pcb.T_MUTEX) bool {
	return mutex.Esta_asignado
}

func Mutex_unlock(nombre_mutex string, proceso *pcb.T_PCB, hilo *pcb.T_TCB) {
	posible_recurso := Recurso_existente(nombre_mutex, proceso)

	if posible_recurso == nil {
		//necesito Finalizar_hilo() pero esa funcion recibe el hilo a finalizar desde la API, no por param
		return
	} else if Mutex_tomado(posible_recurso) && posible_recurso.Hilo_utilizador != hilo {
		// si está tomado pero no es él, no hago nada
		return
	} else if len(posible_recurso.Lista_hilos_bloqueados) == 0 {
		posible_recurso.Hilo_utilizador = nil
		return
	} else {
		posible_recurso.Hilo_utilizador = Primer_hilo_bloqueado_recurso(posible_recurso)
	}
}

func Primer_hilo_bloqueado_recurso(recurso *pcb.T_MUTEX) *pcb.T_TCB {
	return slice.Shift(&recurso.Lista_hilos_bloqueados) // CREERIA que esta tampoco hace falta sincronizarla
}
