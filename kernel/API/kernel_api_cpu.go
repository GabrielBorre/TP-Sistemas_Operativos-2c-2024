package API

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"strings"
	"io/ioutil"
	"time"

	color "github.com/sisoputnfrba/tp-golang/utils/color-log"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	//"github.com/sisoputnfrba/tp-golang/utils/pcb"
	//"github.com/sisoputnfrba/tp-golang/utils/slice"
)

/////////////////////////// ENVIAR TID Y PID A CPU  ///////////////////////////

type T_Tid_y_pid struct {
	PID uint32 `json:"pid"`
	TID uint32 `json:"tid"`
}

func Enviar_tid_a_ejecutar() {
	//log.Printf("TID y PID enviando a ejecutar")

	globals.Sem_m_hilo_actual.Lock()

	if globals.Hilo_actual == nil {
		log.Printf("Cuidado: el hilo que queres mandar a ejecutar esta en nul")
		return
	}
	Tid_y_pid := T_Tid_y_pid{
		TID: globals.Hilo_actual.TID,
		PID: globals.Hilo_actual.PID,
	}
	globals.Sem_m_hilo_actual.Unlock()

	body, err := json.Marshal(Tid_y_pid)
	if err != nil {
		log.Printf("Error codificando TID_y_PID: %s", err.Error())
		return
	}

	color.Log_resaltado(color.LightRed, "TID y PID enviando a ejecutar: %s", string(body))

	url := fmt.Sprintf("http://%s:%d/hilos", globals.Config_kernel.Ip_cpu, globals.Config_kernel.Port_cpu)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_kernel.Ip_cpu, globals.Config_kernel.Port_cpu, err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a CPU falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("TID y PID enviados a CPU correctamente")
	}

}

func Confirmacion_recepcion_TID_y_PID(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var respuesta int

	err := json.NewDecoder(r.Body).Decode(&respuesta)

	if err != nil {
		http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Respuesta recibidos correctamente"))

	if respuesta == 1 {
		//globals.Sem_b_tcb_y_pid_recibido <- true
		log.Printf("CPU recibio TID y PId correctamente, empezamos a esperar interrupciones")

	} else {
		log.Printf("Error wacho, no te llamo una mierda al semaforo")
	}

}

//////////////////////////// Funcion que envia la interrupcion a CPU para desalojar un hilo ////////////////////////////

func Enviar_interrupcion_a_CPU() {
	for {
		//semaforo que espera a que se meta en Ready algun proceso con mayor prioridad al que se esta ejecutando  o que se termine el QUANTUM del hilo

		//el String que nos devuelve es QUANTUM si el hilo se desalojo por fin de QUANTUM
		//o PRIORIDAD si el hilo se desalojo porque llego otro hilo de mayor prioridad
		motivo := <-globals.Sem_string_interrumpir_hilo

		globals.Sem_m_round_robin_corriendo.Lock()
		if motivo == "PRIORIDADES" && globals.Config_kernel.Algoritmo_planificacion == "CMN" && globals.Round_robin_corriendo {
			globals.Sem_b_cancelar_Round_Robin <- true
		}
		globals.Sem_m_round_robin_corriendo.Unlock()

		// si el motivo fura distinto a "QUANTUM" deberiamos matar el hilo de round robin para que no siga contando
		cliente := &http.Client{}

		globals.Sem_m_hilo_actual.Lock()
		if globals.Hilo_actual != nil {
			tid := globals.Hilo_actual.TID
			pid := globals.Hilo_actual.PID
			globals.Sem_m_hilo_actual.Unlock()
			//Primer parametro de la ruta:TID
			//Segundo parametro de la ruta:PID
			//Tercer parametro de la ruta: motivo de desalojo (Quantum o Prioridades)

			url := fmt.Sprintf("http://%s:%d/interrupciones/%d/%d/%s", globals.Config_kernel.Ip_cpu, globals.Config_kernel.Port_cpu, tid, pid, motivo)

			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
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

			log.Printf("Se envio una interrupcion del <TID : %d > del <PID: %d > con motivo de %s", tid, pid, motivo)
			// en realidad, el semaforo se deberia activar cuando la API recibe el hilo desalojado
			//semaforo que activa el planificador de corto plazo (para mandar hilos a ejecutar)
		} else {

			globals.Sem_m_hilo_actual.Unlock()
		}

	}
}
func Round_robin() {
	for {
		globals.Sem_m_round_robin_corriendo.Lock()
		globals.Round_robin_corriendo = false
		globals.Sem_m_round_robin_corriendo.Unlock()
		if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
			<-globals.Sem_b_iniciar_round_Robin // Este canal bloqueará hasta que haya una señal
		}
		globals.Sem_m_round_robin_corriendo.Lock()
		globals.Round_robin_corriendo = true
		globals.Sem_m_round_robin_corriendo.Unlock()

		log.Printf("Empezo a contar el Round Robin")

		select {
		case <-time.After(time.Duration(globals.Config_kernel.Quantum) * time.Millisecond):
			globals.Sem_string_interrumpir_hilo <- "QUANTUM"
			log.Printf("Paso el semaforo de RR y le aviso a la funcion para interrumpir")
			continue
		case <-globals.Sem_b_cancelar_Round_Robin:
			log.Println("Round Robin cancelado por prioridad o replanificación")
			continue
			// Sale del bucle interno, vuelve a esperar una señal de inicio
		}

	}
}

/////////////////////////// RECEPCION SYSCALLS / INTERRUPCIONES CPU ///////////////////////////

/// RECIBIR DESDE LA CPU UN BODY QUE CONTENGA EL NOMBRE DE LA SYSCALL QUE EL HILO SOLICITO
//CON ESE NOMBRE HAGO UN SWITCH DEL NOMBRE DE LA SYSCALL EN LA FUNCION RecibirSyscalls PARA ATENDER LA SYSCALL CON SU FUNCION CORRESPONDIENTE

type CPURequest struct {
	NombreSyscall             string `json:"nombreSyscall"`
	PID                       int    `json:"PID"` //EN REALIDAD, PCB Y PID QUIZAS NO SON NECESARIOS PORQUE EL KERNEL CONOCE EL HILO QUE ESTA EJECUTANDO A TRAVES DE LA COLA DE RUNNING
	TID                       int    `json:"TID"`
	NombreArchivoPseudocodigo string `json:"archivoPseudocodigo"` //este parametro lo necesita thread_create
	Prioridad                 int    `json:"Prioridad"`           //este parametro lo necesita la syscall thread_create
}

type T_syscall struct {
	Syscall string                 `json:"syscall"`
	PID     uint32                 `json:"pid"`
	TID     uint32                 `json:"tid"`
	Params  map[string]interface{} `json:"params"`
}

func Recibir_syscall_CPU(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "No se pudo leer el body de la solicitud", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &globals.Nueva_syscall); err == nil {
		if globals.Nueva_syscall.Syscall != "PRIORIDADES" && globals.Nueva_syscall.Syscall != "QUANTUM" {
			color.Log_obligatorio("##(<PID>:<%d>) (<TID>: <%d>) - Solicitó syscall: <%s> ", globals.Nueva_syscall.PID, globals.Nueva_syscall.TID, globals.Nueva_syscall.Syscall)
		} else {
			color.Log_obligatorio("##(<PID>:<%d>) (<TID>: <%d>) - Fue DESALOJADO con motivo de %s", globals.Nueva_syscall.PID, globals.Nueva_syscall.TID, globals.Nueva_syscall.Syscall)

		}
		globals.Sem_b_llego_syscall <- true

	} else {
		http.Error(w, "Error al parsear la solicitud", http.StatusBadRequest)
	}
}

/////////////////////////// INTERRUPCIONES DE CPU ///////////////////////////

func Recibir_interrupcion_CPU(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "No se pudo leer el body de la solicitud", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &globals.Nueva_interrupcion); err == nil {

		color.Log_obligatorio("##(<PID>:<%d>) - Interrupción recibida: <%s> ", globals.Nueva_interrupcion.Tid, globals.Nueva_interrupcion.InterruptionReason)
		Tratar_interrupcion()

		w.WriteHeader(http.StatusOK)
		return
	}
}
