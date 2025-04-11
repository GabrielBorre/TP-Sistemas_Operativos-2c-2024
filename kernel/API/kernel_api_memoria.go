package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"sync"
	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	syscalls_procesos "github.com/sisoputnfrba/tp-golang/kernel/syscalls/procesos"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	//"github.com/sisoputnfrba/tp-golang/utils/slice"
)

////////////////////////////////////////////////////////////////////////////
//////////////////////// PLANIFICADOR A LARGO PLAZO ////////////////////////
////////////////////////////////////////////////////////////////////////////

///////////////////////// CREACION DE PROCESOS ////////////////////////

type PCBRequest struct {
	PID     uint32 `json:"pid"`
	Tamanio uint32 `json:"tamanio"` //TAMAÑO DEL PROCESO EN MEMORIA
	//TID          uint32 `json:"tid"`
	//Pseudocodigo string `json:"pseudocodigo"`
}

type MemoriaResponse struct {
	Codigo int `json:"codigo"`
}

// /////////////////////Peticion a memoria para inicializar un proceso//////////////////
func Peticion_a_memoria(pcb *pcb.T_PCB, tcb *pcb.T_TCB) {

	// Crear PCBRequest con los campos que se enviarán a memoria
	PCBreq := PCBRequest{
		PID:     pcb.PID,
		Tamanio: pcb.Tamanio,
		//TID:          tcb.TID,
		//Pseudocodigo: tcb.Instrucciones,
	}

	body, err := json.Marshal(PCBreq)
	if err != nil {
		log.Printf("Error codificando PCB: %s", err.Error())

	}
	log.Printf("JSON enviado: %s", string(body))

	url := fmt.Sprintf("http://%s:%d/procesos", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria falló con el código de estado: %d", resp.StatusCode)
	}

	//log.Printf("Petición para crear un proceso a memoria enviada correctamente")

}

func Recibir_solicitud_iniciar_proceso(w http.ResponseWriter, r *http.Request) {

	log.Printf("Recibi una respuesta para iniciar un proceso")

	decoder := json.NewDecoder(r.Body)

	var respuesta int
	err := decoder.Decode(&respuesta)

	if err != nil {
		fmt.Printf("Error al decodificar respuesta: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar respuesta"))
		return
	}

	log.Println("Me llego un respuesta de un cliente")
	log.Printf("%+v\n", respuesta)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	if respuesta == 1 {
		globals.Sem_b_inicializar_pcb <- true
		//log.Println("Memoria me dio el ok. :D")

	} else if respuesta == 0 {
		globals.Sem_b_inicializar_pcb <- false
		log.Println("Memoria no tiene espacio suficiente para inicializar el proceso. :(")
	} else if respuesta == 2 {
		//parar la ejecucion
		globals.Sem_b_detener_planificacion = true

		//canal que bloquea hasta que se detenga la planificacion
		<-globals.Sem_b_planificacion_detenida
		//pedir a memoria que compacte -- ya podes compactar
		Peticion_a_memoria_compactar()

		//esperar a que memoria finalice la compactacion

		log.Printf("Memoria finalizo compactacion. Reanudo la planificacion")

		// le aviso a la API tratar_syscalls que se termino la espera de la compactacion y ya puede replanificar
		globals.Sem_b_esperar_compactacion <- true
		globals.Sem_b_detener_planificacion = false

		//En tratar_syscalls //IF DETENER PLANIFICACION==true --> ESPERAMOS QUE SE ACTIVE EL CANAL DE PLANIFICACION REANUDADA
		//Recibir_solicitud_iniciar_proceso(w, r)
	}
}

func Peticion_a_memoria_compactar() {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/compactar", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Printf("Error al enviar peticion a memoria")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return
	}
	if respuesta.StatusCode != http.StatusOK {
		return
	}

	log.Println("Le pido a memoria que COMPACTE!!!!")
}

func Respuesta_de_memoria_compactacion() {
	for {
		cliente := &http.Client{}
		url := fmt.Sprintf("http://%s:%d/compactar", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error al enviar peticion a memoria")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		respuesta, err := cliente.Do(req)
		if err != nil {
			return
		}
		if respuesta.StatusCode != http.StatusOK {
			return
		}

		log.Println("Memoria me confirmó que compactó <3")
		//globals.Sem_b_detener_planificacion <- false
	}
}

//////////////////////// PETICION A MEMORIA PARA FINALIZAR UN PROCESO ////////////////////////

func Peticion_a_memoria_para_finalizar_proceso() {

	for {
		pcb := <-globals.Sem_tcb_para_finalizar_proceso

		pid := pcb.PID
		tamanio := pcb.Tamanio
		syscalls_procesos.EliminarPCB(pcb.PID)
		//replanifico al mismo timempo que aviso a memoria que tiene que finalizar un proceso
		globals.Sem_b_plani_corto <- true
		log.Printf("Peticion a memoria para la finalizacion del proceso %d", pid)
		cliente := &http.Client{}
		url := fmt.Sprintf("http://%s:%d/finalizarProceso/%d/%d", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory, pid, tamanio)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Printf("Error al enviar peticion a memoria")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		respuesta, err := cliente.Do(req)
		if err != nil {
			return
		}
		if respuesta.StatusCode != http.StatusOK {
			return
		}

		//pcb_finalizado := syscalls_utils.Devolver_pcb(int(pcb.PID))

		color.Log_resaltado(color.BgRed, "Memoria confirma la finalizacion del proceso <PID: %d > - <Tamaño: %d >", pid, tamanio)

		pcb = nil

		//una vez que memoria me confirma la finalizacion del proceso, lo elimino

		globals.Sem_m_plani_largo.Lock()
		if len(globals.Lista_plani_largo) > 0 {
			pcb_a_pedir_inicializacion := globals.Lista_plani_largo[0]
			globals.Sem_pcb_pedido_a_memoria_para_inicializar_proceso <- pcb_a_pedir_inicializacion
		}
		globals.Sem_m_plani_largo.Unlock()

		//le pido a memoria que inicialice el primer proceso de la cola de NEW
		//SEMAFORO QUE LE AVISA A PEDIDO_A_MEMORIA_PARA_INICIALIZACION QUE PUEDE HACER LA PETICION A MEMORIA PARA INICIALIZAR UN PROCESO
	}

}

// ////////////////////// CREAR HILO PETICION ////////////////////////
// Body de la peticion a memoria para que esta cree un hilo
type TCBRequest struct {
	TID                   uint32 `json:"tid"`
	Archivo_instrucciones string `json:"archivo_instrucciones"` //archivo de instrucciones el hilo a crear en memoria
	PID                   uint32 `json:"pid"`
}

func Peticion_a_memoria_para_crear_hilo(tcb *pcb.T_TCB) {
	body, err := json.Marshal(TCBRequest{
		TID:                   tcb.TID,
		PID:                   tcb.PID,
		Archivo_instrucciones: tcb.Instrucciones,
	})
	if err != nil {
		log.Printf("Error al enviar peticion a memoria para crear hilo")
		return
	}
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/hilos", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error al enviar peticion a memoria al crear la solicitud HTTP %s", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Printf("Error al enviar peticion a memoria %s ", err.Error())
		return
	}
	if respuesta.StatusCode != http.StatusOK {
		log.Println("Error al enviar peticion a memoria (respuesta del servidor distinta de 200) ")
		return
	}

	color.Log_resaltado(color.LightOrange, "Memoria aceptó la creacion del hilo <TID:%d> del <PID:%d>", tcb.TID, tcb.PID)

}

//////////////////////// FINALIZAR HILO ////////////////////////

func Peticion_a_memoria_para_finalizar_hilo() {
	for {
		tcb_finalizado := <-globals.Sem_tcb_finalizar_hilo
		cliente := &http.Client{}
		//Paso por query path el tid del hilo y el pid del proceso al que pertenece
		url := fmt.Sprintf("http://%s:%d/finalizarHilo/%d/%d", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory, tcb_finalizado.TID, tcb_finalizado.PID)
		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Printf("Error al enviar peticion a memoria")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		respuesta, err := cliente.Do(req)
		if err != nil {
			return
		}
		if respuesta.StatusCode != http.StatusOK {
			color.Log_resaltado(color.Red, "Memoria no devuelve codigo http 200 OK para finalizar hilo")
			return
		}
		color.Log_obligatorio("## Memoria confirma la finalizacion del hilo <TID: %d > del proceso <PID: %d >", tcb_finalizado.TID, tcb_finalizado.PID)

	}
}

// ////////////////////// DUMP MEMORY ////////////////////////
type T_DumpMemoryRequest struct {
	TID       uint32 `json:"tid"`
	PID       uint32 `json:"pid"`
	Contenido []byte `json:"contenido"`
	Tamanio   int    `json:"tamanio"`
}

func Solicitar_dump_memory(PID uint32, TID uint32) {

	var solicitud_dump_memory T_DumpMemoryRequest
	solicitud_dump_memory.TID = TID
	solicitud_dump_memory.PID = PID

	body, err := json.Marshal(solicitud_dump_memory)
	if err != nil {
		log.Printf("Error codificando solicitud de dump memory: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/dumpMemoryMem", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_kernel.Ip_memory, globals.Config_kernel.Port_memory, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Petición a memoria  para dump memory enviada correctamente")
	}
}

func Recibir_respuesta_dump_memory(w http.ResponseWriter, r *http.Request) {

	//log.Printf("Recibi una respuesta del dump memory")

	decoder := json.NewDecoder(r.Body)

	var respuesta int
	err := decoder.Decode(&respuesta)

	if err != nil {
		fmt.Printf("Error al decodificar respuesta: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar respuesta"))
		return
	}

	//log.Println("Me llego un respuesta de un cliente")

	if respuesta == 1 {
		globals.Sem_int_respuesta_dump_memory <- respuesta
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	} else {
		globals.Sem_int_respuesta_dump_memory <- respuesta
		log.Println("Fallo el down memory")
	}
}
