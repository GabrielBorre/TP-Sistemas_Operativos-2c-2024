package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/////////////////////////// Pregunta para actualizar contexto de ejecucion ///////////////////////////

func Pedir_contexto_ejecucion_memoria(nuevo_Tid_y_pid T_Tid_y_pid) {
	body, err := json.Marshal(nuevo_Tid_y_pid)
	if err != nil {
		log.Printf("Error codificando nuevo_Tid_y_pid: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/contextoEjecucion", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory, err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria de contexto de ejecucion falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Solicitud a memoria enviada correctamente de: %+v", nuevo_Tid_y_pid)
	}

}

func Recibir_contexto_ejecucion_memoria(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&globals.Contexto_de_ejecucion)

	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("nuevo_Tid_y_pid recibidos correctamente"))

	if globals.Contexto_de_ejecucion != nil {
		//log.Printf("Recibimos contexto de ejecucion+%v", globals.Contexto_de_ejecucion)
		globals.Sem_b_recepcion_contexto_ejecucion <- true
	} else {
		log.Printf("Recibimos contexto de ejecucion PERO ESTA VACIO")
	}
}

/////////////////////////// Pregunta para instruccion ///////////////////////////

func Pedir_instruccion_a_memoria() {

	body, err := json.Marshal(globals.Peticion_instruccion)
	if err != nil {
		log.Printf("Error codificando peticion de instruccion: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/instruccion", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria de instruccion falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("Solicitud a memoria de instruccion enviada correctamente")
		//log.Printf("+%v", globals.Peticion_instruccion)
	}
}
func Recibir_instruccion_de_memoria(w http.ResponseWriter, r *http.Request) {

	var instruccion_de_memoria string

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&instruccion_de_memoria)

	globals.Instruccion_de_memoria = instruccion_de_memoria

	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Instrucciones recibidas correctamente"))

	if globals.Instruccion_de_memoria != "" {
		//log.Printf("Recibimos %+v", instruccion_de_memoria)
		globals.Sem_b_recepcion_instruccion_de_memoria <- true
	} else {
		log.Printf("Recibimos instrucciones PERO ESTÁ VACÍO")
	}
}

/////////////////////////// ACTUALIZAR CONTEXTO DE EJECUCION ///////////////////////////

// ver que url usamos para el contexto
func Actualizar_contexto() {

	url := fmt.Sprintf("http://%s:%d/actualizarContextoEjecucion", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	jsonData, _ := json.Marshal(globals.CurrentTCB)

	if globals.CurrentTCB == nil {
		log.Printf("estas queriendo enviar un tcb que es nulo a memoria")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer((jsonData)))
	if err != nil {
		fmt.Println("Error al crear la request:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	cliente := &http.Client{}
	resp, err := cliente.Do(req)
	if err != nil {
		fmt.Println("Error al enviar el request a Memoria:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		color.Log_obligatorio("## TID: <%d> - Actualizo Contexto Ejecución", globals.CurrentTCB.TID)
		//log.Printf("+%v", globals.CurrentTCB.Registros_CPU)
	} else {
		log.Printf("Sonamos y no se actualizo nada")
	}
}

/////////////////////////// RECUPERACION TCB Y PCB ///////////////////////////

func Pedir_pcb_a_memoria(PID uint32) {

	body, err := json.Marshal(PID)
	if err != nil {
		log.Printf("Error codificando peticion de instruccion: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/recuperacionPCB", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria de PCB falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Solicitud a memoria de PCB enviada correctamente")
	}

}

func Pedir_tcb_a_memoria(TID uint32) {

	body, err := json.Marshal(TID)
	if err != nil {
		log.Printf("Error codificando peticion de instruccion: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/recuperacionTCB", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a memoria de TID falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("Solicitud a memoria de TID enviada correctamente")
	}

}

func Recibir_tcb_de_memoria(w http.ResponseWriter, r *http.Request) {

	var tcb *pcb.T_TCB

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&tcb)

	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TCB recibida correctamente"))

	if tcb != nil {
		//Printf("Recibimos %+v", tcb)
		globals.Sem_b_recepcion_tcb <- true
	} else {
		log.Printf("Recibimos tcb PERO ESTA VACIO")
	}
}

func Recibir_pcb_de_memoria(w http.ResponseWriter, r *http.Request) {

	var pcb *pcb.T_PCB

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&pcb)

	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PCB recibida correctamente"))

	if pcb != nil {
		//log.Printf("Recibimos %+v", pcb)
		globals.Sem_b_recepcion_pcb <- true
	} else {
		log.Printf("Recibimos pcb PERO ESTA VACIO")
	}
}

/////////////////////////// BASE Y LIMITE DE MEMORIA PARA READ/WRITE _MEM ///////////////////////////

func Solicitar_base_y_limite(tid uint32, pid uint32) {

	var tid_y_pid globals.T_Tid_y_pid

	tid_y_pid.PID = pid
	tid_y_pid.TID = tid

	body, err := json.Marshal(tid_y_pid)
	if err != nil {
		log.Printf("Error codificando Contexto ejecucion: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/baseYlimite", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a Memoria falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("Tid y Pid para base y limite enviado correctamente")
	}
}

func Recibir_base_y_limite(w http.ResponseWriter, r *http.Request) {

	var byl globals.T_base_y_limite

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&byl)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	} else {
		//log.Println("Me llego:")
		//log.Printf("%+v\n", byl)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	globals.Sem_byl_read_write_mem <- byl

}

func Recibir_confirmacion_escritura(w http.ResponseWriter, r *http.Request) {

	var respuesta uint32

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&respuesta)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	if respuesta == 1 {
		log.Printf("Se pudo realizar la escritura correctamente")
	} else {
		log.Println("todo mal... borra system 32")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

}
