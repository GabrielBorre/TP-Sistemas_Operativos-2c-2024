package kernel_syscalls

import (
	"log"

	//"time"
	//"io/ioutil"

	"strconv"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	syscalls_hilos "github.com/sisoputnfrba/tp-golang/kernel/syscalls/hilos"
	syscalls_mutex "github.com/sisoputnfrba/tp-golang/kernel/syscalls/mutex"
	syscalls_utils "github.com/sisoputnfrba/tp-golang/kernel/syscalls/utils_asincro"
	kernel_utils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

//////////////////////// TRATAMIENTO DE SYSCALLS  ////////////////////////

func Tratar_syscall() {

	for {

		//log.Printf("Esperando a una syscall")
		<-globals.Sem_b_llego_syscall

		//log.Printf("Llego syscall")

		switch globals.Nueva_syscall.Syscall {
		case "PROCESS_CREATE":
			//saco los params de la sys
			archivo := globals.Nueva_syscall.Params["archivo"].(string)

			tamanio := globals.Nueva_syscall.Params["tamanio"].(string)

			prioridad := globals.Nueva_syscall.Params["prioridad"].(string)

			//primero transformo el string a uint64 tamanio y prioridad
			tamanio_uint_64, err := strconv.ParseUint(tamanio, 10, 32)
			if err != nil {
				log.Printf("Error al convertir tamanio a uint64")
				return
			}
			prioridad_uint_64, err := strconv.ParseUint(prioridad, 10, 32)
			if err != nil {
				log.Printf("Error al convertir prioridad a uint64")
			}
			//despues tengo que transformar el uint64 a uint 32 (chat GPT me dice que es asi)

			log.Printf("El tamanio del proceso a crear es %d", tamanio_uint_64)

			log.Printf("La prioridad del hilo main del proceso a crear es %d", prioridad_uint_64)
			//sexo, el error es porque todavia no las cree xD
			kernel_utils.Crear_proceso(archivo, uint32(tamanio_uint_64), uint32(prioridad_uint_64))

			//enviamos devuelta el hilo a ejecutar
			kernel_api.Enviar_tid_a_ejecutar()

		case "PROCESS_EXIT":
			//tcb := globals.Hilo_actual

			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))

			//le aviso a una rutina que le pida a memoria finalizar un proceso
			//globals.Sem_tcb_para_finalizar_proceso <- tcb

			globals.Sem_tcb_para_finalizar_proceso <- pcb

			//le aviso al planificador de corto plazo que debe replanificar

			globals.Sem_m_round_robin_corriendo.Lock()
			if globals.Config_kernel.Algoritmo_planificacion == "CMN" && globals.Round_robin_corriendo {
				globals.Sem_b_cancelar_Round_Robin <- true
			}
			globals.Sem_m_round_robin_corriendo.Unlock()

			syscalls_utils.Chequear_detener_planificacion()

		case "THREAD_CREATE":
			archivo := globals.Nueva_syscall.Params["archivo"].(string)
			prioridad := globals.Nueva_syscall.Params["prioridad"].(string)

			prioridad_uint_64, err := strconv.ParseUint(prioridad, 10, 32)
			if err != nil {
				log.Printf("Error al convertir prioridad a uint64")
			}

			//log.Printf("La prioridad del hilo a crear es %d ", prioridad_uint_64)

			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))

			//creamos el hilo
			//log.Printf("LOG DE CONTROL : PID %d TID %d ", globals.Hilo_actual.PID, globals.Hilo_actual.TID)
			syscalls_hilos.Crear_hilo(pcb, archivo, int(prioridad_uint_64))
			//log.Printf("LOG DE CONTROL : PID %d TID %d ", globals.Hilo_actual.PID, globals.Hilo_actual.TID)
			//enviamos el hilo que solicito la syscall devuelta a ejecutar
			kernel_api.Enviar_tid_a_ejecutar()

		case "THREAD_EXIT":

			//primero busco el proceso al q pertenece el hilo que solicito la syscall
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))

			//busco el hilo que solicito la syscall en la lista de pcbs del proceso
			tcb := syscalls_utils.Devolver_tcb(uint32(globals.Hilo_actual.TID), pcb.TCBs)

			if tcb != nil {
				//log.Printf("Voy a terminar el hilo TID: %d ", tcb.TID)

				syscalls_hilos.Finalizar_hilo(tcb)

				//replanificamos (activamos el planificador de corto plazo para que mande un hilo a exec)

				log.Printf("Paso antes de la cancelacion del RR")

				globals.Sem_m_round_robin_corriendo.Lock()
				log.Printf("paso despues del lock")
				if globals.Config_kernel.Algoritmo_planificacion == "CMN" && globals.Round_robin_corriendo {
					globals.Sem_b_cancelar_Round_Robin <- true
					log.Printf("Paso despues de la cancelacion del RR")
				}
				globals.Sem_m_round_robin_corriendo.Unlock()

			} else {

				log.Printf("Che, el TCB es nul")
			}

			syscalls_utils.Chequear_detener_planificacion()

			globals.Sem_b_plani_corto <- true

		case "THREAD_JOIN":
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))
			tid := globals.Nueva_syscall.Params["tid"].(string)

			tid_uint_64, err := strconv.ParseUint(tid, 10, 32)
			if err != nil {
				log.Printf("Error al convertir prioridad a uint64")
			}

			//la logica de la replanificacion se encuentra dentro de la funcion THREAD_JOIN
			syscalls_hilos.Thread_join(uint32(tid_uint_64), pcb.TCBs)

		//falta probar, la logica esta
		case "THREAD_CANCEL":
			tid_a_cancelar := globals.Nueva_syscall.Params["tid_a_cancelar"].(string)
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))

			tid_a_cancelar_uint_64, err := strconv.ParseUint(tid_a_cancelar, 10, 32)
			if err != nil {
				log.Printf("Error al convertir prioridad a uint64")
			}

			syscalls_hilos.Thread_cancel(uint32(tid_a_cancelar_uint_64), pcb.TCBs)

			kernel_api.Enviar_tid_a_ejecutar() //envio de vuelta el hilo a ejecutar a CPU

		case "MUTEX_LOCK":
			recurso := globals.Nueva_syscall.Params["recurso"].(string)
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))
			syscalls_mutex.Mutex_lock(recurso, pcb, globals.Hilo_actual)

		case "MUTEX_UNLOCK":
			recurso := globals.Nueva_syscall.Params["recurso"].(string)
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))
			syscalls_mutex.Mutex_unlock(recurso, pcb, globals.Hilo_actual)

		case "MUTEX_CREATE":
			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))
			recurso := globals.Nueva_syscall.Params["recurso"].(string)
			syscalls_mutex.Mutex_create(pcb, recurso)
			//log.Printf("Se quiere mandar de vuelta a ejecutar el TID %d del PID %d ", globals.Hilo_actual.TID, globals.Hilo_actual.PID)
			kernel_api.Enviar_tid_a_ejecutar()

		case "DUMP_MEMORY":
			//bloquear hilo
			globals.Cambiar_estado_hilo(globals.Hilo_actual, "BLOCKED")
			kernel_api.Solicitar_dump_memory(globals.Hilo_actual.PID, globals.Hilo_actual.TID)
			respuesta := <-globals.Sem_int_respuesta_dump_memory
			if respuesta != 1 {
				//en caso de error, el proceso se enviará a EXIT. (por que wtf)
				globals.Cambiar_estado_proceso(globals.Proceso_actual, "EXIT")
				globals.Sem_tcb_finalizar_hilo <- globals.Hilo_actual
				log.Println("error en DUMP_MEMORY")

			} else {
				//desbloquear hilo
				globals.Cambiar_estado_hilo(globals.Hilo_actual, "READY")

				globals.Sem_tcb_enviar_a_cola_de_Ready <- globals.Hilo_actual

				log.Println("exito en DUMP_MEMORY")

			}

			globals.Sem_b_plani_corto <- true

		case "IO":
			tiempo_string := globals.Nueva_syscall.Params["tiempo"].(string)
			tiempo, err := strconv.ParseUint(tiempo_string, 10, 32)
			if err != nil {
				log.Printf("Error al convertir tamanio a uint64")
				return
			}

			log.Printf("PID %d TID %d ANTES DE LA IO ", globals.Hilo_actual.PID, globals.Hilo_actual.TID)

			globals.Sem_m_round_robin_corriendo.Lock()
			if globals.Config_kernel.Algoritmo_planificacion == "CMN" && globals.Round_robin_corriendo {
				globals.Sem_b_cancelar_Round_Robin <- true
			}
			globals.Sem_m_round_robin_corriendo.Unlock()

			syscalls_utils.Chequear_detener_planificacion()

			globals.Sem_b_plani_corto <- true

			globals.Sem_tcb_realizar_operacion_io <- globals.Hilo_actual
			globals.Sem_int_tiempo_milisegundos <- int(tiempo)

		//en caso de que el motivo de interrupcion sea prioridades, se va a mandar a ejecutar el hilo de mayor prioridad
		case "PRIORIDADES":

			log.Printf("<TID %d> <PID %d> fue desalojado por la llegada de un hilo de mayor prioridad", globals.Hilo_actual.TID, globals.Hilo_actual.PID)

			syscalls_utils.Chequear_detener_planificacion()

			//insertamos el hilo que acaba de ser interrumpido a la cola de Ready, de acuerdo a su prioridad
			switch globals.Config_kernel.Algoritmo_planificacion {
			case "PRIORIDADES":
				kernel_utils.Insertar_en_cola_ready_prioridades(globals.Hilo_actual)

			case "CMN":
				cola_multinivel := kernel_utils.Buscar_cola_multinivel(globals.Hilo_actual.Prioridad)

				log.Printf("Se inserta el hilo TID %d PID %d en la cola de prioridad %d ", globals.Hilo_actual.TID, globals.Hilo_actual.PID, globals.Hilo_actual.Prioridad)

				slice.Push(&cola_multinivel.Cola, globals.Hilo_actual)

			}

			globals.Sem_b_plani_corto <- true

		case "QUANTUM":
			log.Printf("<PID %d> <TID: %d> -Desalojado por fin de quantum ", globals.Hilo_actual.PID, globals.Hilo_actual.TID)

			syscalls_utils.Chequear_detener_planificacion()

			//busco la cola multinivel dentro de la
			cola_multinivel := kernel_utils.Buscar_cola_multinivel(globals.Hilo_actual.Prioridad)

			log.Printf("Se inserta el hilo TID %d PID %d al final de la cola cuya prioridad es %d", globals.Hilo_actual.TID, globals.Hilo_actual.PID, globals.Hilo_actual.Prioridad)

			slice.Push(&cola_multinivel.Cola, globals.Hilo_actual)

			globals.Sem_b_plani_corto <- true

		case "SEGMENTATION FAULT":

			syscalls_utils.Chequear_detener_planificacion()

			log.Printf("El hilo TID %d del PID %d causo Segmentation Fault. Se elimina todo el proceso PID %d ", globals.Hilo_actual.TID, globals.Hilo_actual.PID, globals.Hilo_actual.PID)

			pcb := syscalls_utils.Devolver_pcb(int(globals.Hilo_actual.PID))

			globals.Sem_tcb_para_finalizar_proceso <- pcb

			if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
				globals.Sem_b_cancelar_Round_Robin <- true
			}

			//globals.Sem_b_plani_corto <- true

		default:
			log.Printf("Syscall no reconocida: %s", globals.Nueva_syscall.Syscall)
		}

	}
}

/*
func eliminarPCB(pcb_finalizado *pcb.T_PCB) {
	///ELIMINA AL PCB QUE SOLICITO LA SYSCALLL PROCESS_EXIT
	cantidad_hilos_del_pcb := len(pcb_finalizado.TCBs) //GUARDO EN UNA VARIABLE LA CANTIDAD DE HILOS DEL PCB A FINALIZAR
	for i := 0; i < cantidad_hilos_del_pcb; i++ {      //RECORRO TODOS LOS HILOS DEL PCB
		pcb_finalizado.TCBs[i].Estado = "EXIT"                                                //SETEO EL ESTADO DEL HILO A EXIT
		slice.PushMutex(&globals.Cola_Exit, pcb_finalizado.TCBs[i], &globals.Sem_m_cola_exit) //MANDO CADA HILO A LA COLA DE EXIT
		pcb_finalizado.TCBs[i].PCB = nil                                                      //HAGO QUE CADA HILO DEJE DE APUNTAR AL PCB PARA QUE EL GARBAGE COLLECTOR LIBERE LA MEMORIA DEL PCB FINALIZADO                                            //MARCO
	}

}

func Delimitador(sysActual string) []string {
	// separa el nombre de la syscall de los parámetros usando " - " como delimitador
	partes := strings.SplitN(sysActual, " - ", 2)

	// si hay menos de 2 partes, significa que no hay parámetros (como en THREAD_EXIT o PROCESS_EXIT)
	if len(partes) < 2 {
		return []string{strings.TrimSpace(partes[0])}
	}

	// la primera parte es el nombre de la syscall
	nombreSyscall := strings.TrimSpace(partes[0])

	// la segunda parte contiene los parámetros, separarlos por comas si es necesario
	parametros := strings.Split(partes[1], ",")

	// quita espacios en blanco
	for i := range parametros {
		parametros[i] = strings.TrimSpace(parametros[i])
	}

	// devuelve una lista con el nombre de la syscall seguido de los parámetros
	return append([]string{nombreSyscall}, parametros...)
}

func Enviar_pcb() error {
	jsonData, err := json.Marshal(globals.Proceso_actual)
	if err != nil {
		return fmt.Errorf("failed to encode PCB: %v", err)
	}

	client := &http.Client{
		Timeout: 0,
	}

	// Manda por el canal de DISPATCH a la CPU
	// Todavia falta cambiar en la parte de CPU()
	url := fmt.Sprintf("http://%s:%d/dispatch", globals.Config_kernel.Ip_cpu, globals.Config_kernel.Port_cpu)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("POST request failed. Failed to send PCB: %v", err)
	}

	// Esperamos respuesta
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("respuesta no esperada: %s", resp.Status)
	}

	// Decode response and update value
	// Esto es la respuesta del CPU creeeeeeo
	err = json.NewDecoder(resp.Body).Decode(&globals.Proceso_actual) // ? Semaforo?
	if err != nil {
		return fmt.Errorf("failed to decode PCB response: %v", err)
	}

	globals.Sem_b_pcb_recibido <- true

	return nil
}
*/
