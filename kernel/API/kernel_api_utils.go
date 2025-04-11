package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
	//"github.com/sisoputnfrba/tp-golang/utils/slice"
)


////////////////////////////////// PROCESOS //////////////////////////////////

/*
- getProcessList: Devuelve una lista de todos los procesos en el sistema (LTS, STS, Blocked, STS_Priority, CurrentJob)
- @return []pcb.T_PCB: Lista de procesos
*/
func Get_lista_procesos() []*pcb.T_PCB {
	var procesos []*pcb.T_PCB

	procesos = append(procesos, globals.Lista_plani_corto...)
	procesos = append(procesos, globals.Lista_plani_largo...)
	procesos = append(procesos, globals.Lista_bloqueados...)

	// Por ahora estos no (nico)
	//procesos = append(procesos, globals.STS_Priority...)
	//procesos = append(procesos, globals.Terminated...)

	if globals.Proceso_actual.PID != 0 && Pid_no_esta_en_lista(globals.Proceso_actual.PID, procesos) {
		procesos = append(procesos, globals.Proceso_actual)
	}
	return procesos
}

func Pid_no_esta_en_lista(pid uint32, list []*pcb.T_PCB) bool {
	for _, process := range list {
		if process.PID == pid {
			return false
		}
	}
	return true
}

/*
- GetPIDList: Devuelve una lista de PID de todos los procesos en el sistema
- @param []pcb.T_PCB: Lista de procesos
- @return []uint32: Lista de PID
*/
func Get_lista_PIDs([]*pcb.T_PCB) []uint32 {
	var pidList []uint32

	for _, pcb := range Get_lista_procesos() {
		pidList = append(pidList, pcb.PID)
	}
	return pidList
}

////////////////////////////////// INTERRUPCIONES //////////////////////////////////

type InterruptionRequest struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid                uint32 `json:"pid"`
	ExecutionNumber    int    `json:"execution_number"`
}

/*
- SendInterrupt: Envia una interrupción a CPU
- @param reason: Motivo de la interrupción
- @param pid: PID del proceso a interrumpir
- @param executionNumber: Número de ejecución del proceso
*/
func SendInterrupt(reason string, pid uint32, executionNumber int) {
	url := fmt.Sprintf("http://%s:%d/interrupt", globals.Config_kernel.Ip_cpu, globals.Config_kernel.Port_cpu)

	bodyInt, err := json.Marshal(InterruptionRequest{
		InterruptionReason: reason,
		Pid:                pid,
		ExecutionNumber:    executionNumber,
	})
	if err != nil {
		return
	}

	enviarInterrupcion, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyInt))
	if err != nil {
		log.Printf("POST request failed (No se puede enviar interrupción): %v", err)
	}

	cliente := &http.Client{}
	enviarInterrupcion.Header.Set("Content-Type", "application/json")
	recibirRta, err := cliente.Do(enviarInterrupcion)
	if err != nil || recibirRta.StatusCode != http.StatusOK {
		log.Println("Error al interrupir proceso", err)
	}
}