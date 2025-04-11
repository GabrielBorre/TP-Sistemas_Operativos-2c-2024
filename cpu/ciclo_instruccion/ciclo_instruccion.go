package ciclo_instruccion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	mmu "github.com/sisoputnfrba/tp-golang/cpu/MMU"
	"github.com/sisoputnfrba/tp-golang/cpu/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/**
 * Delimitador: Función que separa la instrucción en sus partes
 * @param instActual: Instrucción a separar
 * @return instruccion_decodificada: Instrucción separada
**/

func Delimitador(instActual string) []string {
	delimitador := " "
	i := 0

	instruccion_decodificada_con_comillas := strings.Split(instActual, delimitador)
	instruccion_decodificada := instruccion_decodificada_con_comillas

	largo_instruccion := len(instruccion_decodificada_con_comillas)
	for i < largo_instruccion {
		instruccion_decodificada[i] = strings.Trim(instruccion_decodificada_con_comillas[i], `"`)
		i++
	}

	return instruccion_decodificada
}

// --------------------------------------------------FETCH seguro sufra cambios para comtemplar el TCB(quien posee el PC)-----------------------------------------------------------

// var a  (desactualizada) t_contexto_ejecucion
// PODRIAMOS USAR SEMAFOROS
func convertir(registros_CPU map[string]interface{}) {
	for key, value := range registros_CPU {
		switch key {
		case "AX":
			globals.CurrentTCB.Contexto_ejecucion.AX = value.(uint32)
		case "BX":
			globals.CurrentTCB.Contexto_ejecucion.BX = value.(uint32)
		case "CX":
			globals.CurrentTCB.Contexto_ejecucion.CX = value.(uint32)
		case "DX":
			globals.CurrentTCB.Contexto_ejecucion.DX = value.(uint32)
		case "EX":
			globals.CurrentTCB.Contexto_ejecucion.EX = value.(uint32)
		case "FX":
			globals.CurrentTCB.Contexto_ejecucion.FX = value.(uint32)
		case "GX":
			globals.CurrentTCB.Contexto_ejecucion.GX = value.(uint32)
		case "HX":
			globals.CurrentTCB.Contexto_ejecucion.HX = value.(uint32)
		case "PC":
			globals.CurrentTCB.Contexto_ejecucion.PC = value.(uint32)
		case "BASE":
			globals.CurrentTCB.Contexto_ejecucion.BASE = value.(uint32)
		case "LIMITE":
			globals.CurrentTCB.Contexto_ejecucion.LIMITE = value.(uint32)
		}
	}
}

var hubo_syscall bool
var hubo_interrupcion bool
var hubo_set_del_PC bool

func Ciclo_de_instruccion() {

	for {
		hubo_syscall = false
		hubo_interrupcion = false
		hubo_set_del_PC = false
		globals.Hubo_seg_fault = false
		//log.Printf("FETCH: Comenzando ciclo")
		<-globals.Sem_b_recepcion_contexto_ejecucion

		if globals.Peticion_instruccion == nil {
			globals.Peticion_instruccion = &globals.T_peticion_instruccion{
				PID: globals.Tid_y_pid.PID,
				TID: globals.Tid_y_pid.TID,
				PC:  globals.Contexto_de_ejecucion.PC,
			}
		} else {
			globals.Peticion_instruccion.PID = globals.Tid_y_pid.PID
			globals.Peticion_instruccion.TID = globals.Tid_y_pid.TID

		}

		if globals.CurrentTCB == nil {
			globals.CurrentTCB = &pcb.T_TCB{
				PID:           globals.Tid_y_pid.PID,
				TID:           globals.Tid_y_pid.TID,
				Registros_CPU: Convertir_contexto_a_registros(),
			}
		} else {
			globals.CurrentTCB.TID = globals.Tid_y_pid.TID
			globals.CurrentTCB.PID = globals.Tid_y_pid.PID
			globals.CurrentTCB.Registros_CPU = Convertir_contexto_a_registros()
			globals.Peticion_instruccion.PC = globals.Contexto_de_ejecucion.PC
			log.Printf("PC de la proxima instruccion a ejeuctar %d ", globals.Contexto_de_ejecucion.PC)

		}

		Fetch()
		Decode_and_execute()
		if hubo_syscall {
			<-globals.Sem_b_recepcion_contexto_ejecucion
			Check_interrupt()
			if !hubo_interrupcion {
				globals.Sem_b_recepcion_contexto_ejecucion <- true
			}
		} else {
			Check_interrupt()
			if !hubo_interrupcion {
				globals.Sem_b_recepcion_contexto_ejecucion <- true
			}
		}
		aumentar_PC()

	}
}

func Fetch() {

	//SEMAFORO (SE ACTIVA CUANDO LE LLEGA UN HILO A EJECUTAR O EN LOS CASE QUE NO IMPLIQUEN)

	cpu_api.Pedir_instruccion_a_memoria()

	<-globals.Sem_b_recepcion_instruccion_de_memoria

	color.Log_resaltado(color.Yellow, "FETCH: Memoria nos dio las instruciones, %+v", globals.Instruccion_de_memoria)

	/*if globals.CurrentTCB == nil {
		globals.CurrentTCB = &pcb.T_TCB{
			PID:           globals.Tid_y_pid.PID,
			TID:           globals.Tid_y_pid.TID,
			Registros_CPU: Convertir_contexto_a_registros(*globals.Contexto_de_ejecucion),
		}
	}*/
}

func aumentar_PC() {
	//si no hubo syscall, al final del ciclo aumento el PC.Cuando hay syscall el PC aumenta antes de enviarle a memoria el contexto de ejecucion
	if !hubo_syscall && !hubo_interrupcion && !hubo_set_del_PC {
		globals.Peticion_instruccion.PC = globals.Peticion_instruccion.PC + 1
		globals.CurrentTCB.Registros_CPU["PC"] = globals.Peticion_instruccion.PC
		globals.CurrentTCB.Contexto_ejecucion.PC = globals.Peticion_instruccion.PC
		globals.Contexto_de_ejecucion.PC = globals.Peticion_instruccion.PC
	}
}

func Convertir_contexto_a_registros() map[string]interface{} {
	return map[string]interface{}{
		"AX":     globals.Contexto_de_ejecucion.AX,
		"BX":     globals.Contexto_de_ejecucion.BX,
		"CX":     globals.Contexto_de_ejecucion.CX,
		"DX":     globals.Contexto_de_ejecucion.DX,
		"EX":     globals.Contexto_de_ejecucion.EX,
		"FX":     globals.Contexto_de_ejecucion.FX,
		"GX":     globals.Contexto_de_ejecucion.GX,
		"HX":     globals.Contexto_de_ejecucion.HX,
		"PC":     globals.Contexto_de_ejecucion.PC,
		"BASE":   globals.Contexto_de_ejecucion.BASE,
		"LIMITE": globals.Contexto_de_ejecucion.LIMITE,
	}
}

func Decode_and_execute() {

	instruccion_decodificada := Delimitador(globals.Instruccion_de_memoria)

	if instruccion_decodificada[0] == "EXIT" {

		globals.CurrentTCB.MotivoDesalojo = "EXIT"
		pcb.FlagDesalojo = true

		color.Log_resaltado(color.LightPurple, "MOTIVO EXIT DE TID: %d - Ejecutando: %s", globals.CurrentTCB.TID, instruccion_decodificada[0])

		//ACA VOLVEMOS AL KERNEL
		//TERMINA EL HILO

	} else {
		color.Log_resaltado(color.LightPurple, "TID: %d - Ejecutando: %s - %s", globals.CurrentTCB.TID, instruccion_decodificada[0], instruccion_decodificada[1:])
	}

	switch instruccion_decodificada[0] {
	case "SET":

		valor := instruccion_decodificada[2]

		globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]] = Convertir_Uint32(valor)

		if instruccion_decodificada[1] == "PC" {

			hubo_set_del_PC = true
			globals.CurrentTCB.Contexto_ejecucion.PC = Convertir_Uint32(valor)

		}
		convertir(globals.CurrentTCB.Registros_CPU)
		globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

	case "SUB":
		//SUB (Registro Destino, Registro Origen): Resta al Registro Destino
		//el Registro Origen y deja el resultado en el Registro Destino.
		valorReg1 := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]]
		valorReg2 := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[2]]

		tipoActualReg1 := reflect.TypeOf(valorReg1).String()
		tipoActualReg2 := reflect.TypeOf(valorReg2).String()

		globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]] = Convertir[uint32](tipoActualReg1, valorReg1) - Convertir[uint32](tipoActualReg2, valorReg2)

		if instruccion_decodificada[1] == "PC" {
			globals.CurrentTCB.Contexto_ejecucion.PC = globals.CurrentTCB.Registros_CPU["PC"].(uint32)
		}
		convertir(globals.CurrentTCB.Registros_CPU)
		globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

	case "SUM":
		//tipoReg1 := pcb.Tipo_reg(instruccion_decodificada[1]) // instruccion_decodificada[1] es siempre uint32
		valorReg1 := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]]
		valorReg2 := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[2]]

		tipoActualReg1 := reflect.TypeOf(valorReg1).String()
		tipoActualReg2 := reflect.TypeOf(valorReg2).String()

		globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]] = Convertir[uint32](tipoActualReg1, valorReg1) + Convertir[uint32](tipoActualReg2, valorReg2)

		if instruccion_decodificada[1] == "PC" {
			globals.CurrentTCB.Contexto_ejecucion.PC = globals.CurrentTCB.Registros_CPU["PC"].(uint32)
		}
		convertir(globals.CurrentTCB.Registros_CPU)
		globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

	case "JNZ":
		if globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]] != 0 {
			globals.CurrentTCB.Contexto_ejecucion.PC = Convertir_Uint32(instruccion_decodificada[2])
			globals.CurrentTCB.Registros_CPU["PC"] = uint32(globals.CurrentTCB.Contexto_ejecucion.PC)
		}
		convertir(globals.CurrentTCB.Registros_CPU)
		globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

	case "LOG":
		//ACA TENEMOS QUE LOGUEAR EN EL ARCHIVO DE LOGS (CHK3)
		registro := instruccion_decodificada[1]
		log.Printf("Registro %s : %v", registro, globals.CurrentTCB.Registros_CPU[registro])
		convertir(globals.CurrentTCB.Registros_CPU)
		globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

	case "READ_MEM":

		// Registros Datos, Registros Direccion
		// Lee el valor de memoria correspondiente a la Dirección Lógica
		// que se encuentra en el Registro Dirección
		// Y lo almacena en el Registro Datos.

		reg_direccion, exists := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[2]]
		if !exists {
			log.Printf("Registro no encontrado: %s", instruccion_decodificada[2])
			return
		}

		if reg_direccion == nil {
			log.Printf("Este registro es nil, lo ponemos en 1 porque no puede ser nil")
			reg_direccion = uint32(1)
		} else if reg_direccion == 0 {
			log.Printf("Este registro es cero, lo ponemos en 1 porque no puede ser cero")
			reg_direccion = uint32(1)
		}

		// Convierte a uint32 de forma segura
		direccion_fisica, err := mmu.Traducir_direccion_logica(reg_direccion.(uint32), globals.CurrentTCB.TID, globals.CurrentTCB.PID)
		if err != nil {
			log.Fatal("No se pudo traducir la direccion logica", err)
		}

		if !globals.Hubo_seg_fault {
			valor := Leer_memoria(direccion_fisica)

			//log.Printf("El valor que nos devolvio memoria es: %d", valor)

			globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]] = uint32(valor)

			convertir(globals.CurrentTCB.Registros_CPU)
			globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

			log.Printf("TID: %d - LEER - DIRECCION FISICA: %d", globals.CurrentTCB.TID, direccion_fisica) //LOG OBLIGATORIO
		}
	case "WRITE_MEM":
		//(Registro Dirección, Registro Datos): Lee el valor del Registro Datos y
		//lo escribe en la dirección física de memoria obtenida a partir de la Dirección Lógica almacenada
		//en el Registro Dirección.
		// Obtener reg_direccion y manejar el error de conversión
		reg_direccion_raw, exists := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[1]]
		if !exists {
			log.Printf("Registro no encontrado: %s", instruccion_decodificada[1])
			return // Manejar el caso en que el registro no existe
		}

		reg_direccion_uint32, ok := reg_direccion_raw.(uint32)
		if !ok {
			log.Fatalf("Error de tipo: se esperaba uint32, pero se encontró %T", reg_direccion_raw)
		}

		reg_datos_raw, exists := globals.CurrentTCB.Registros_CPU[instruccion_decodificada[2]]
		if !exists {
			log.Printf("Registro no encontrado: %s", instruccion_decodificada[2])
			return // Manejar el caso en que el registro no existe
		}

		// Afirmar que reg_datos_raw es uint32
		reg_datos_uint32, ok := reg_datos_raw.(uint32)
		if !ok {
			log.Fatalf("Error de tipo: se esperaba uint32, pero se encontró %T", reg_datos_raw)
		}

		// Traducir la dirección lógica
		direccion_fisica, err := mmu.Traducir_direccion_logica(reg_direccion_uint32, globals.CurrentTCB.TID, globals.CurrentTCB.PID)
		if err != nil {
			log.Fatal("No se pudo traducir la direccion logica: ", err)
		}

		if !globals.Hubo_seg_fault {
			Escribir_memoria(direccion_fisica, reg_datos_uint32)
			convertir(globals.CurrentTCB.Registros_CPU)
			globals.Contexto_de_ejecucion = (*globals.T_contexto_ejecucion)(&globals.CurrentTCB.Contexto_ejecucion)

			log.Printf("TID: %d - ESCRIBIR - DIRECCION FISICA: %d", globals.CurrentTCB.TID, direccion_fisica) //LOG OBLIGATORIO
		}
	case "DUMP_MEMORY":
		aumentar_PC()
		hubo_syscall = true
		// Actualizamos el contexto de ejecución en Memoria
		cpu_api.Actualizar_contexto()
		//Devolvemos ejecución a kernel
		cpu_api.Enviar_syscall_a_kernel("DUMP_MEMORY", nil)

	case "IO":
		aumentar_PC()
		hubo_syscall = true
		tiempoIO := instruccion_decodificada[1]
		// Actualizamos el contexto en Memoria
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"tiempo": tiempoIO,
		}
		// Enviamos el TID y el tiempo de IO al Kernel
		cpu_api.Enviar_syscall_a_kernel("IO", params)

	case "PROCESS_CREATE":
		aumentar_PC()
		hubo_syscall = true
		archivoInstrucciones := instruccion_decodificada[1]
		tamanio := instruccion_decodificada[2]
		prioridad := instruccion_decodificada[3]

		cpu_api.Actualizar_contexto()

		params := map[string]interface{}{
			"archivo":   archivoInstrucciones,
			"tamanio":   tamanio,
			"prioridad": prioridad,
		}

		cpu_api.Enviar_syscall_a_kernel("PROCESS_CREATE", params)

	case "THREAD_CREATE":
		aumentar_PC()
		hubo_syscall = true
		archivoInstrucciones := instruccion_decodificada[1]
		prioridad := instruccion_decodificada[2]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"archivo":   archivoInstrucciones,
			"prioridad": prioridad,
		}
		cpu_api.Enviar_syscall_a_kernel("THREAD_CREATE", params)

	case "THREAD_JOIN":
		aumentar_PC()
		hubo_syscall = true
		tid := instruccion_decodificada[1]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"tid": tid,
		}
		cpu_api.Enviar_syscall_a_kernel("THREAD_JOIN", params)
	case "THREAD_CANCEL":
		aumentar_PC()
		hubo_syscall = true
		tid_a_cancelar := instruccion_decodificada[1]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"tid_a_cancelar": tid_a_cancelar,
		}
		cpu_api.Enviar_syscall_a_kernel("THREAD_CANCEL", params)

	case "MUTEX_CREATE":
		aumentar_PC()
		hubo_syscall = true
		recurso := instruccion_decodificada[1]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"recurso": recurso,
		}
		cpu_api.Enviar_syscall_a_kernel("MUTEX_CREATE", params)

	case "MUTEX_LOCK":
		aumentar_PC()
		hubo_syscall = true
		recurso := instruccion_decodificada[1]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"recurso": recurso,
		}
		cpu_api.Enviar_syscall_a_kernel("MUTEX_LOCK", params)

	case "MUTEX_UNLOCK":
		aumentar_PC()
		hubo_syscall = true
		recurso := instruccion_decodificada[1]
		cpu_api.Actualizar_contexto()
		params := map[string]interface{}{
			"recurso": recurso,
		}
		cpu_api.Enviar_syscall_a_kernel("MUTEX_UNLOCK", params)

	case "THREAD_EXIT":
		aumentar_PC()
		hubo_syscall = true
		cpu_api.Actualizar_contexto()
		cpu_api.Enviar_syscall_a_kernel("THREAD_EXIT", nil)

	case "PROCESS_EXIT":
		aumentar_PC()
		hubo_syscall = true
		cpu_api.Actualizar_contexto()
		cpu_api.Enviar_syscall_a_kernel("PROCESS_EXIT", nil)

	}
}

//////////////////// MEMORIA ////////////////////
/*
- Leer_memoria: lee un valor guardado en memoria
- @param int: la direccion fisica donde esta guardado el valor
*/
func Leer_memoria(direccion_fisica uint32) int {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/read", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	var valor int

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error al crear la solicitud HTTP: %v", err)
		return 0
	}

	pid := strconv.FormatInt(int64(globals.CurrentTCB.PID), 10)
	q := req.URL.Query()
	q.Add("direccion_fisica", strconv.FormatUint(uint64(direccion_fisica), 10))
	q.Add("pid", pid)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Printf("Error al leer memoria")
		return 0
	}

	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		log.Printf("Error en la respuesta del servidor de memoria: %v", respuesta.Status)
		return 0
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Printf("Error al leer el cuerpo de la respuesta: %v", err)
		return 0
	}

	err = json.Unmarshal(bodyBytes, &valor)
	if err != nil {
		log.Printf("Error al decodificar el valor de la memoria: %v", err)
		return 0
	}

	log.Println(string(bodyBytes))
	log.Printf("Devolvemos lectura de memoria, %+v", valor)
	return valor
}

/*
- Escribir_memoria: escribe un valor en memoria
- @param int: la direccion fisica donde se va a guardar el valor
- @param string: el registro de donde lee el valor
*/
//////////////////// WRITE MEM ////////////////////

type T_body_request_escritura struct {
	Direccion_fisica uint32 `json:"direccion_fisica"`
	Valor            uint32 `json:"valor"`
	PID              int    `json:"pid"`
}

func Escribir_memoria(direccion_fisica uint32, registro_datos uint32) {

	payload := T_body_request_escritura{
		Direccion_fisica: direccion_fisica,
		Valor:            registro_datos,
		PID:              int(globals.CurrentTCB.PID),
	}

	jsonData, _ := json.Marshal(payload)

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/writeMem", globals.Config_cpu.Ip_memory, globals.Config_cpu.Port_memory)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error al crear la solicitud HTTP: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	respuesta, err := cliente.Do(req)
	if err != nil {
		log.Printf("Error al escribir en memoria")
	}

	defer respuesta.Body.Close()

	if respuesta.StatusCode != http.StatusOK {
		log.Printf("Error en la respuesta del servidor de memoria: %v", respuesta.Status)
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		log.Printf("Error al leer el cuerpo de la respuesta: %v", err)
	}

	fmt.Println(string(bodyBytes))
	fmt.Println("OK write_mem")

}
