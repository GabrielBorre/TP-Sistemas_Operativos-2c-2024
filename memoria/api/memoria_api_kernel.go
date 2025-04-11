package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	particiones "github.com/sisoputnfrba/tp-golang/memoria/particiones"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
)

/*
------------------------------------------------------------------------ 	KERNEL		------------------------------------------------------------------------
*/

// ////////////////////////// INICIAR PROCESO ////////////////////////////
type Solicitud_iniciar_proceso struct {
	PID     uint32 `json:"pid"`
	Tamanio uint32 `json:"tamanio"`
}

func Enviar_respuesta_solicitud_iniciar_proceso(ip string, puerto int) {

	for {

		log.Printf("Esperando que me pidan crear un proceso")

		respuesta := <-globals.Sem_int_inicializar_pcb

		log.Printf("Paso el semaforo y respondo que %d", respuesta)

		//por ahora porque siempre responde que SI/OK

		body, err := json.Marshal(respuesta)
		if err != nil {
			log.Printf("error codificando mensaje: %s", err.Error())
		}

		url := fmt.Sprintf("http://%s:%d/procesos", ip, puerto)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("error enviando confirmacion a ip:%s puerto:%d", ip, puerto)
		}
		//defer resp.Body.Close()

		log.Printf("respuesta del servidor: %s \n", resp.Status)
	}

}

func Recibir_solicitud_iniciar_proceso(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	log.Printf("RECIBO DEL KERNEL SOLICITUD DE INICIAR UN PROCESO")

	var solicitud Solicitud_iniciar_proceso
	err := decoder.Decode(&solicitud)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}
	//Printf("Recibi una solicitud para iniciar un proceso")
	//log.Printf("%+v\n", solicitud)

	var pid_int int = int(solicitud.PID)

	var pid_puntero *int = &pid_int

	particiones.Recibir_pedido_creacion_particion(int(solicitud.Tamanio), pid_puntero)

	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	//semaforo

}

// ////////////////////////// INICIAR HILO ////////////////////////////
type TCBRequest struct {
	TID                   uint32 `json:"tid"`
	Archivo_instrucciones string `json:"archivo_instrucciones"`
	PID                   uint32 `json:"pid"`
}

// Solo devuelve un OK ante las solicitudes del kernel de crear un hilo
func Recibir_solicitud_iniciar_hilo(w http.ResponseWriter, r *http.Request) {
	var request TCBRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Me llego la solicitud de crear el hilo <TID: %d > del <PID :%d> ", request.TID, request.PID)

	//Lógica de creación de hilos a partir
	//hilos.Creacion_hilo()

	//creo el hilo con el json que me pasaron
	tcb_nuevo := &pcb.T_TCB{
		TID:           request.TID,
		Instrucciones: request.Archivo_instrucciones,
		PID:           request.PID,
	}

	//busco la particion del proceso al que pertenece el TCB
	var particion *globals.Particion = particiones.Devolver_particion_asociada_al_proceso(int(tcb_nuevo.PID))

	//no deberia pasar por aca nunca, pero chequeo esto para q no haya seg fault
	if particion == nil {
		log.Printf("No encontre la particion del proceso al que pertenece el hilo a crear, calavera, ataud")
		return
	}

	//creo el los registros inicializados en 0 del contexto de ejecucion del hilo
	contexto_ejecucion_hilo := pcb.T_contexto_ejecucion{
		AX:     0,
		BX:     0,
		CX:     0,
		DX:     0,
		EX:     0,
		FX:     0,
		GX:     0,
		HX:     0,
		BASE:   uint32(particion.Base),
		LIMITE: uint32(particion.Limite),
	}

	//agregamos al tcb creado el contexto de ejecucion propio de ese tcb

	tcb_nuevo.Contexto_ejecucion = contexto_ejecucion_hilo
	log.Printf("TID <%d> del PID %d creado. Contexto Ejecucion %d , %d, %d", tcb_nuevo.TID, tcb_nuevo.PID, tcb_nuevo.Contexto_ejecucion.AX, tcb_nuevo.Contexto_ejecucion.BX, tcb_nuevo.Contexto_ejecucion.CX)

	//abro el archivo de instrucciones cuya ruta viene en el Json
	archivo, err := os.Open(tcb_nuevo.Instrucciones)
	if err != nil {
		log.Fatal(err)
	}
	defer archivo.Close()

	//leo el archivo linea x linea y cargo cada linea en el vector de instrucciones del tcb

	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		//Agrego cada linea al slice de instrucciones del tcb creado
		tcb_nuevo.Lista_instrucciones_a_ejecutar = append(tcb_nuevo.Lista_instrucciones_a_ejecutar, scanner.Text())
	}

	// Verificar si hubo errores al escanear el archivo
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//pusheo el hilo a la lista de hilos creados

	slice.PushMutex(&globals.Lista_hilos_creados, tcb_nuevo, &globals.Sem_mutex_lista_hilos)

	//imprimo las lineas para verificar que funcione bien

	color.Log_resaltado(color.BoldBlue, "Instrucciones del hilo <TID: %d > del <PID>: %d", tcb_nuevo.TID, tcb_nuevo.PID)

	for _, instruccion := range tcb_nuevo.Lista_instrucciones_a_ejecutar {
		log.Printf("Linea %s ", instruccion)
	}

	w.WriteHeader(http.StatusOK)

}

//////////////////////////// FINALIZAR PROCESO ////////////////////////////

func Recibir_solicitud_finalizar_proceso(w http.ResponseWriter, r *http.Request) {
	pid_string := r.PathValue("PID")
	tamanio_string := r.PathValue("Tamanio")
	pid, err := strconv.Atoi(pid_string)
	if err != nil {
		log.Printf("Error al convertir el string de pid a int")
		http.Error(w, "PID inválido", http.StatusBadRequest)
		return
	}
	tamanio, err := strconv.Atoi(tamanio_string)
	if err != nil {
		log.Printf("Error al convertir el string de tamanio a int")
		http.Error(w, "Tamaño inválido", http.StatusBadRequest)
		return
	}
	particiones.Liberar_particion_asociada(pid)
	particiones.Finalizar_hilos_asociados_al_proceso(pid)
	w.WriteHeader(http.StatusOK)
	color.Log_obligatorio("## Proceso <Destruido> - PID: %d - Tamaño: %d", pid, tamanio)
	//log.Printf("Kernel solicito la finalizacion del proceso <PID :%d >", pid)

}

func Recibir_solicitud_finalizar_hilo(w http.ResponseWriter, r *http.Request) {

	tid_string := r.PathValue("TID")
	pid_string := r.PathValue("PID")
	tid, err := strconv.Atoi(tid_string) //convirto el tid de string a int
	if err != nil {
		log.Printf("Error al convertir el TID de string a int")
	}
	pid, err := strconv.Atoi(pid_string) //convirto el pid de string a int
	if err != nil {
		log.Printf("Error al convertir el PID de string a int")
	}

	w.WriteHeader(http.StatusOK)

	//hilos.Finalizar_hilo(pid, tid)

	color.Log_obligatorio("##Kernel solicito la finalizacion del hilo <TID: %d > del proceso <PID : %d >", tid, pid)

}

// ////////////////////////// COMPACTACION ////////////////////////////

func Recibir_solicitud_y_compactar(w http.ResponseWriter, r *http.Request) {
	// Verificar que el método sea PUT
	if r.Method != http.MethodPut {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	//compacto
	particiones.Compactarr()

	// Responder con estado OK al cliente
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Compactación completada")
	globals.Sem_b_compactacion_recibida_y_realizada <- true
}

//////////////////////////// DUMP MEMORY ////////////////////////////

func Recibir_solicitud_dump_memory(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var solicitud_dump_memory globals.T_DumpMemoryRequest
	err := decoder.Decode(&solicitud_dump_memory)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	//log.Printf("Me llego un solicitud de un hilo para un dump memory:")
	//log.Printf("PID: %d, TID: %d", solicitud_dump_memory.PID, solicitud_dump_memory.TID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	//ACA TENEMOS QUE AGREGAR EL CONTENIDO DEL PROCESO PARA MANDARSELO AL FS

	Enviar_dump_memory_a_fs(solicitud_dump_memory)

	respuesta := <-globals.Sem_b_finalizo_dump_memory

	Enviar_respuesta_dump_memory(respuesta)
}

func Enviar_respuesta_dump_memory(respuesta uint32) {

	body, err := json.Marshal(respuesta)
	if err != nil {
		log.Printf("Error codificando respuesta a dump memory: %s", err.Error())
		return
	}

	//log.Printf("Enviando respuesta dump memory: %s", string(body))

	url := fmt.Sprintf("http://%s:%d/dumpMemoryMem", globals.Config_memoria.Ip_kernel, globals.Config_memoria.Port_kernel)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_kernel, globals.Config_memoria.Port_kernel, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a kernel falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Respuesta de dump memory enviados a kernel correctamente")
	}

}
