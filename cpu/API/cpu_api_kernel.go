package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/globals"

	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/slice"
	//ciclo_instruccion "github.com/sisoputnfrba/tp-golang/cpu/ciclo_instruccion"
)

/////////////////////////// TID Y PID de Kernel para ejectura hilo ///////////////////////////

/**
PID_y_TID_recv: Recibe un tid y un pid de kernel
Lo envia a memoria para volver a recibirlo como un pcb
Cumple con la funcionalidad principal de CPU.
Procesar = Fetch -> Decode -> Execute
*/

type T_Tid_y_pid struct {
	PID uint32 `json:"pid"`
	TID uint32 `json:"tid"`
}

type T_motivo_replanificacion struct {
	TID    uint32 `json:"tid"`
	Motivo string `json:"motivo"`
}

var nueva_solicitud_Tid_y_pid T_Tid_y_pid

func Recibir_pid_y_tid(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&nueva_solicitud_Tid_y_pid)
	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	color.Log_resaltado(color.Blue, "Recibido TID: %d, PID: %d", nueva_solicitud_Tid_y_pid.TID, nueva_solicitud_Tid_y_pid.PID)

	// Procesar los datos recibidos, por ejemplo, enviarlos al canal
	// cpuChan <- tidYPid

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("TID y PID recibidos correctamente"))

	globals.Tid_y_pid.PID = nueva_solicitud_Tid_y_pid.PID
	globals.Tid_y_pid.TID = nueva_solicitud_Tid_y_pid.TID

	//globals.Sem_b_recepcion_tcb_para_ejecutar <- true

	Pedir_contexto_ejecucion_memoria(nueva_solicitud_Tid_y_pid)
	//Confirmar_recepcion_nueva_solicitud_Tid_y_pid()

}

/////////////////////////// INTERRUPCIONES ///////////////////////////

func Manejar_interrupcion(w http.ResponseWriter, r *http.Request) {

	tid_string := r.PathValue("TID")
	pid_string := r.PathValue("PID")
	motivo := r.PathValue("MOTIVO")

	tid, err := strconv.Atoi(tid_string) //convirto el tid de string a int
	if err != nil {
		log.Printf("Error al convertir el TID de string a int")
	}
	pid, err := strconv.Atoi(pid_string) //convirto el pid de string a int
	if err != nil {
		log.Printf("Error al convertir el PID de string a int")
	}

	// esta linea la puse para confirmar que se enviaba correctamente el mensaje desde el kernel, luego habria que colocarla en la fase CheckInterrupt
	log.Printf("Llego una interrupcion para el <TID: %d > del <PID: %d > con MOTIVO: %s ", tid, pid, motivo)

	// creo la interrupcion que llega desde el kernel
	interrupcion := globals.T_interrupcion{
		Tid:                uint32(tid),
		Pid:                uint32(pid),
		InterruptionReason: motivo,
	}
	//agrego la interrupcion a una cola de interrupciones, la cual luego sera atendida en la fase checkInterrupt (nunca antes)
	slice.PushMutex(&globals.Lista_interrupciones, interrupcion, &globals.Sem_m_Lista_interrupciones_mutex)

}

func Enviar_motivo_check_interrupt(interrupcion globals.T_interrupcion) {

	body, err := json.Marshal(interrupcion)
	if err != nil {
		log.Printf("Error codificando motivo de replanificacion: %s", err.Error())
		return
	}

	url := fmt.Sprintf("http://%s:%d/dispatch", globals.Config_cpu.Ip_kernel, globals.Config_cpu.Port_kernel)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_cpu.Ip_kernel, globals.Config_cpu.Port_kernel, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a Kernel falló con el código de estado: %d", resp.StatusCode)
	} else {
		log.Printf("Confirmacion de motivo de replanificacion enviada correctamente: %+v", interrupcion)
	}

}

//////////////////// SYSCALLS A KERNEL ////////////////////

func Enviar_syscall_a_kernel(syscall string, params map[string]interface{}) {

	if globals.CurrentTCB == nil {
		log.Printf("El current TCB es nulo papa")
		return
	}
	Syscall := globals.T_syscall{
		Syscol: syscall,
		PID:    globals.CurrentTCB.PID,
		TID:    globals.CurrentTCB.TID,
		Params: params,
	}

	url := fmt.Sprintf("http://%s:%d/dispatch", globals.Config_cpu.Ip_kernel, globals.Config_cpu.Port_kernel)

	jsonData, _ := json.Marshal(Syscall)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error al crear la request:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error al mandar la syscall al Kernel:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("TID: %d - Syscall: %s", globals.CurrentTCB.TID, syscall)
	} else {
		log.Printf("No se ejecuto la syscall")
	}

}
