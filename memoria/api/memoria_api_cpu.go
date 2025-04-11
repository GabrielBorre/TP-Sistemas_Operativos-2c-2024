package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	//"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	operaciones "github.com/sisoputnfrba/tp-golang/memoria/operaciones"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

////////////////////////////// OBTENER CONTEXTO DE EJECUCION //////////////////////////////

func Recibir_solicitud_contexto_ejecucion(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&globals.Nueva_solicitud_Tid_y_pid)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	//log.Printf("Me llego un solicitud de contexto de ejecucion para:%+v\n", globals.Nueva_solicitud_Tid_y_pid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	Enviar_contexto_ejecucion()

	// Aquí iría el manejo del semáforo para enviar el contexto de ejecucion segun PID y TID
	// log.Printf("Llamamos al semáforo para procesar el contexto de ejecución")

}

func Enviar_contexto_ejecucion() {
	//arrancan inicializados en cero jij
	//Buscamos contexto de ejecucion (falta hacer para ch3)
	Retardo_peticion()
	var tcb *pcb.T_TCB = Buscar_en_lista_hilos_creados(globals.Nueva_solicitud_Tid_y_pid.PID, globals.Nueva_solicitud_Tid_y_pid.TID)

	body, err := json.Marshal(tcb.Contexto_ejecucion)
	if err != nil {
		log.Printf("Error codificando contexto de ejecucion: %s", err.Error())

	}

	url := fmt.Sprintf("http://%s:%d/contextoEjecucion", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		color.Log_resaltado(color.Cyan, "Contexto de ejecucion enviado correctamente a:%+v\n", globals.Nueva_solicitud_Tid_y_pid)
	}

}

////////////////////////////// INSTRUCCION //////////////////////////////

func LeerArchivoTexto(rutaArchivo string) (string, error) {
	// Abre el archivo
	file, err := os.Open(rutaArchivo)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Lee todo el contenido del archivo
	contenido, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(contenido), nil
}

func Recibir_solicitud_instruccion(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&globals.Peticion_instruccion)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}
	log.Printf("Me llego una solicitud de instruccion para %+v", globals.Peticion_instruccion)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	//buscamos en la lista el hilo correspondiente al TID y PID que nos pasaron

	//instruccion_a_mandar=tcb.instrucciones[PC]

	//enviamos instruccion_a_mandar
	Enviar_instrucion_a_cpu()

	// Aquí iría el manejo del semáforo para enviar el contexto de ejecucion segun PID y TID
	// log.Printf("Llamamos al semáforo para procesar el contexto de ejecución")

}

func Buscar_en_lista_hilos_creados(pid_buscado uint32, tid_buscado uint32) *pcb.T_TCB {
	for i := 0; i < len(globals.Lista_hilos_creados); i++ {
		if ( (globals.Lista_hilos_creados[i].PID == pid_buscado) && (globals.Lista_hilos_creados[i].TID == tid_buscado) ) {
			return globals.Lista_hilos_creados[i]
		}
	}
	log.Printf("No se encontro el hilo PID:%d y TID:%d xD :^)", pid_buscado, tid_buscado)
	return nil
}

// CH3 Y PRUEBAS HARDCODEAR ARCHIVO CONFIG
func Enviar_instrucion_a_cpu() {
	Retardo_peticion()
	tcb := Buscar_en_lista_hilos_creados(globals.Peticion_instruccion.PID, globals.Peticion_instruccion.TID)
	pc_buscado := globals.Peticion_instruccion.PC

	if globals.Peticion_instruccion.PC >= uint32(len(tcb.Lista_instrucciones_a_ejecutar)) {
		color.Log_resaltado(color.Red, "Ya no quedan más instrucciones")
		return
	}

	instruccion := tcb.Lista_instrucciones_a_ejecutar[pc_buscado]

	body, err := json.Marshal(instruccion)
	if err != nil {
		log.Printf("Error codificando instruccion: %s", err.Error())
	}

	url := fmt.Sprintf("http://%s:%d/instruccion", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		color.Log_resaltado(color.Yellow, "Enviada correctamente:%+s\nenviada a %+v\n",
			instruccion,
			globals.Peticion_instruccion) // llega hasta acá la ejecucion, ver si cpu tiene una recepcion o algo
	}
}

/*
Se deberá devolver el valor correspondiente a los primeros 4 bytes a partir del byte enviado como dirección física
dentro de la Memoria de Usuario.
*/
func Leer_valor(w http.ResponseWriter, r *http.Request) {
	Retardo_peticion()
	queryParams := r.URL.Query()
	direccion_fisica := queryParams.Get("direccion_fisica")
	PID := queryParams.Get("pid")

	direccion_uint32, err := strconv.ParseUint(direccion_fisica, 10, 32)
	if err != nil {
		http.Error(w, "Error: La dirección física debe ser un número", http.StatusBadRequest)
		return
	}

	pidInt, err := strconv.Atoi(PID)
	if err != nil {
		log.Printf("Error al convertir PID a entero: %v", err)
		return // O maneja el error de otra forma, según lo necesites
	}

	pidPointer := &pidInt
	color.Log_resaltado(color.Yellow, "Recibí una solicitud para leer en memoria: %d\n", direccion_uint32)
	valor := operaciones.Read_mem(uint32(direccion_uint32), pidPointer)

	log.Printf("EL valor leido fue %d ", valor)

	respuesta, err := json.Marshal(valor)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// //////////////////////////// WRITE_MEM //////////////////////////////
// CAMBIO DE []byte a string ver de cambiar en chk3 (nico)
type T_body_request_escritura struct {
	Direccion_fisica uint32 `json:"direccion_fisica"`
	Valor            uint32 `json:"valor"`
	PID              int    `json:"pid"`
}

/*
Se escribirán los 4 bytes enviados
a partir del byte enviado como dirección física dentro de la Memoria de Usuario y se responderá como OK.
*/

func Escribir_valor(w http.ResponseWriter, r *http.Request) {
	Retardo_peticion()
	var request T_body_request_escritura
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("El templo de agua es facil, pasa que vos sos imbecil")
		return
	}
	color.Log_resaltado(color.Yellow, "Recibí una solicitud para escribir en memoria: %+v\n", request)
	confirmacion := operaciones.Write_mem(request.Direccion_fisica, request.Valor, &request.PID)

	respuesta, err := json.Marshal(confirmacion)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	/*
		if confirmacion == 0 {
			log.Printf("No se pudo realizar la escritura en memoria")
		} else {
			Confirmar_escritura_memoria()
		}
	*/
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func Confirmar_escritura_memoria() {
	var respuesta = 1

	body, err := json.Marshal(respuesta)
	if err != nil {
		log.Printf("Error codificando byl: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/writeMem", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Enviamos confirmación de escritura")
	}
}

////////////////////////////// RECUPERACION TCB Y PCB //////////////////////////////

//ACA EL TCB DEBERIA VENIR CON ALGO MAS PORQUE HAY VARIOS TCB SON EL MISMO TID PORQUE SON DE VARIOS PROCESOS

func Recibir_solicitud_tcb(w http.ResponseWriter, r *http.Request) {

	var TCB uint32

	//log.Printf("Recibi una solicitud para enviar una TCB")

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&TCB)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	//log.Printf("Me llego un solicitud de TCB para:%+v\n", TCB)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	Devolver_solicitud_tcb()

}

func Recibir_solicitud_pcb(w http.ResponseWriter, r *http.Request) {

	var PCB uint32

	//log.Printf("Recibi una solicitud para enviar una PCB")

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&PCB)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	//log.Printf("Me llego un solicitud de PCB para:%+v\n", PCB)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	Devolver_solicitud_pcb()

}

func Devolver_solicitud_tcb() {

	var tcb pcb.T_TCB

	body, err := json.Marshal(tcb)
	if err != nil {
		log.Printf("Error codificando tcb: %s", err.Error())

	}

	url := fmt.Sprintf("http://%s:%d/recuperacionTCB", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("tcb enviada correctamente, enviamos %+v", tcb) // llega hasta acá la ejecucion, ver si cpu tiene una recepcion o algo
	}
}

func Devolver_solicitud_pcb() {

	var pcb pcb.T_PCB

	body, err := json.Marshal(pcb)
	if err != nil {
		log.Printf("Error codificando pcb: %s", err.Error())

	}

	url := fmt.Sprintf("http://%s:%d/recuperacionPCB", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("pcb enviada correctamente, enviamos %+v", pcb) // llega hasta acá la ejecucion, ver si cpu tiene una recepcion o algo
	}
}

////////////////////////////// ACTUALIZACION CONTEXTO EJECUCION //////////////////////////////

func Recibir_actualizacion_contexto_ejecucion(w http.ResponseWriter, r *http.Request) {
	Retardo_peticion()
	var nuevo_contexto_ejecucion *pcb.T_TCB
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&nuevo_contexto_ejecucion)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	color.Log_resaltado(color.Orange, "Nuevo contexto de ejecucion tiene:\n%+v\n", nuevo_contexto_ejecucion.Registros_CPU)
	
	tcb := Buscar_en_lista_hilos_creados(nuevo_contexto_ejecucion.PID, nuevo_contexto_ejecucion.TID)
	//log.Printf("Voy a entrar ;)")
	
	if tcb != nil {

		tcb.Registros_CPU = nuevo_contexto_ejecucion.Registros_CPU
		tcb.Contexto_ejecucion = nuevo_contexto_ejecucion.Contexto_ejecucion
	} else {
		log.Printf("No tengo creado el TID %d del PID %d, amigo", nuevo_contexto_ejecucion.TID, nuevo_contexto_ejecucion.PID)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	//Devolver_solicitud_pcb()
}

////////////////////////////// READ_MEM DE CPU //////////////////////////////

func Recibir_solicitud_base_y_limite(w http.ResponseWriter, r *http.Request) {

	var tid_y_pid globals.T_Tid_y_pid

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&tid_y_pid)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	log.Printf("Recibi una solicitud para base y limite: %+v\n", tid_y_pid)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	Enviar_base_y_limite(tid_y_pid)
}

func Enviar_base_y_limite(tid_y_pid globals.T_Tid_y_pid) {

	var byl globals.T_base_y_limite
	byl.Base = 1
	byl.Limite = 5

	body, err := json.Marshal(byl)
	if err != nil {
		log.Printf("Error codificando byl: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/baseYlimite", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_cpu, globals.Config_memoria.Port_cpu, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		color.Log_resaltado(color.Yellow, "Base y Limite enviados correctamente, enviamos:%+v para %+v\n ", byl, tid_y_pid)
	}
}

////////////////////////////// WRITE_MEM DE CPU //////////////////////////////

func Retardo_peticion() {
	retardo := globals.Config_memoria.Delay_response
	time.Sleep(time.Duration(retardo) * time.Millisecond)
}
