package globals

import (
	"sync"

	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

////////////////////////////////// VARIABLES //////////////////////////////////

// Ver si además de la estructura del pcb, decidimos cambiar/agregar al TCB y modificando todas estas variables
var (

	//Procesos y Planificacion
	Estado_planificacion string // "DETENIDA" O "CORRIENDO"
	Siguiente_PID        uint32 = 0
	Siguiente_TID        uint32 = 0 //Esto no me parece bien xq el TID no es global, sino que es relativo al proceso
	Proceso_actual       *pcb.T_PCB
	Config_kernel        T_config_kernel

	//Hilos
	Nuevo_hilo  *pcb.T_TCB
	Hilo_actual *pcb.T_TCB

	//Interrupciones
	Nueva_interrupcion T_interrupcion
	Interrumpir_RR     bool = false //variable que me dice si tengo que interrumpir el RR porque el proceso fue desalojado por prioridad y no por fin de quantum

	//Syscalls
	Nueva_syscall T_syscall

	//Colas
	Cola_Ready   []*pcb.T_TCB // Cola de READY donde van a estar los hilos que ya estan inicializados en memoria
	Cola_Running []*pcb.T_TCB // Cola de RUNNING: cola de un solo hilo (el que en ese momento esté ejecutando en CPU)
	Cola_Exit    []*pcb.T_TCB
	Cola_Blocked []*pcb.T_TCB
	//Listas
	Lista_plani_largo      []*pcb.T_PCB // LTS - Esta es la cola de NEW
	Lista_plani_corto      []*pcb.T_PCB // STS
	Lista_bloqueados       []*pcb.T_PCB
	Lista_de_Procesos      []*pcb.T_PCB             // Lista en donde se encontraran que el kernel administre
	Lista_de_Hilos         []*pcb.T_TCB             //Lista en donde se encontraran todos los que estén presentes en el sistema
	Lista_mutexs           []*pcb.T_MUTEX           //Lista en donde se encontraran los mutexes asignados a hilos
	Lista_colas_multinivel []*pcb.T_Cola_Multinivel //Lista en donde se encontraran todas las colas multinivel creadas

	//primer hilo
	Primera_planificacion bool = true //es true cuando se planifica por primera vez y pasa a false cuando se manda al hilo TID0 del PID0 a ejecutar por primera vez

	Round_robin_corriendo = true
	//variable que uso en thread_cancel y finalizar_hilo

)

////////////////////////////////// ESTRUCTURAS //////////////////////////////////

type T_config_kernel struct {
	IP_kernel               string   `json:"ip_kernel"`
	Port_kernel             int      `json:"port_kernel"`
	Ip_memory               string   `json:"ip_memory"`
	Port_memory             int      `json:"port_memory"`
	Ip_cpu                  string   `json:"ip_cpu"`
	Port_cpu                int      `json:"port_cpu"`
	Port_cpu_interrupt      int      `json:"port_cpu_interrupt"`
	Algoritmo_planificacion string   `json:"scheduler_algorithm"`
	Quantum                 int      `json:"quantum"`
	Resources               []string `json:"resources"`
	Resource_instances      []int    `json:"resource_instances"`
	Multiprogramming        int      `json:"multiprogramming"`
}

type T_interrupcion struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid                uint32 `json:"pid"`
	Tid                uint32 `json:"tid"`
	ExecutionNumber    int    `json:"execution_number"`
}

type T_syscall struct {
	Syscall string                 `json:"syscall"`
	PID     uint32                 `json:"pid"`
	TID     uint32                 `json:"tid"`
	Params  map[string]interface{} `json:"params"` //son los distintos parametros que llegan dependiendo de la syscall
}

////////////////////////////////// SEMAFOROS //////////////////////////////////

var (
	// Mutex
	Sem_m_plani_corto sync.Mutex
	Sem_m_plani_largo sync.Mutex

	Sem_m_lista_procesos sync.Mutex
	Sem_m_lista_hilos    sync.Mutex
	Sem_m_lista_mutex    sync.Mutex

	Sem_m_cola_new              sync.Mutex
	Sem_m_cola_ready            sync.Mutex
	Sem_m_cola_running          sync.Mutex
	Sem_m_cola_blocked          sync.Mutex
	Sem_m_cola_exit             sync.Mutex
	Sem_m_round_robin_corriendo sync.Mutex

	Sem_m_hilo_actual            sync.Mutex
	Sem_m_lista_colas_multinivel sync.Mutex

	// Binarios
	Sem_b_plani_largo                  = make(chan bool, 1)
	Sem_b_plani_corto                  = make(chan bool, 1)
	Sem_b_plani_largo_vacia            = make(chan bool, 1)
	Sem_b_pcb_recibido                 = make(chan bool, 1)
	Sem_b_inicializar_pcb              = make(chan bool, 1)
	Sem_b_tcb_y_pid_recibido           = make(chan bool, 1)
	Sem_b_iniciar_round_Robin          = make(chan bool, 1) // semaforo que se activa luego de que se confirma la recepcion del tid y el pid por parte de la CPU
	Sem_b_llego_syscall                = make(chan bool, 1) // semaforo que se activa luego de que se confirma la recepcion del tid y el pid por parte de la CPU
	Sem_b_llego_interrupcion           = make(chan bool, 1) // semaforo que se activa luego de que se confirma la recepcion del tid y el pid por parte de la CPU
	Sem_b_cancelar_Round_Robin         = make(chan bool, 1)
	Sem_b_esperar_compactacion         = make(chan bool, 1)
	Sem_b_planificacion_detenida       = make(chan bool, 1)
	Sem_b_compactacion_finalizada      = make(chan bool, 1)
	Sem_b_round_robin_cancelado        = make(chan bool, 1)
	Sem_b_detener_planificacion   bool = false
	// Contadores
	Sem_c_grado_multiprog   chan int
	Sem_c_plani_corto_plazo chan int // Contador asociado a la cola de READY

	//PCB
	Sem_pcb_pedido_a_memoria_para_inicializar_proceso = make(chan *pcb.T_PCB, 1) /////Semaforo que le va a avisar a una funcion para que le avise a memoria que tiene que inicializar un proceso

	//TCB
	Sem_tcb_finalizar_hilo         = make(chan *pcb.T_TCB, 1) //Semaforo cuando se pide finalizar un hilo: se le pasa al TCB a una funcion para que esta lo finalice (se recibe la syscall en la API y se activa este semaforo)
	Sem_tcb_enviar_a_cola_de_Ready = make(chan *pcb.T_TCB, 1)
	Sem_tcb_realizar_operacion_io  = make(chan *pcb.T_TCB, 1) //canal que le envia a la rutina Entrada_salida el tcb que solicito esta syscall
	Sem_tcb_para_finalizar_proceso = make(chan *pcb.T_PCB, 1)
	Sem_tcb_para_crear_hilo        = make(chan *pcb.T_TCB, 1)

	//STRING
	Sem_string_interrumpir_hilo = make(chan string, 1) ////Semaforo que se activa hay que desalojar un hilo, ya sea por la llegada de un hilo con mayor prioridad al que se esta ejecutando o porque finalizo el quantum del hilo ejecutante)//el string que se le va a pasar es el motivo por el que se interrumpe: "QUANTUM" o "PRIORIDAD"

	//INT
	Sem_int_respuesta_dump_memory = make(chan int, 1)
	Sem_int_tiempo_milisegundos   = make(chan int, 1)
)

////////////////////////////////// FUNCIONES //////////////////////////////////

func Cambiar_estado_hilo(tcb *pcb.T_TCB, newState string) {

	estado_anterior := tcb.Estado
	tcb.Estado = newState
	color.Log_resaltado(color.BgWhite, "HILO TID: %d - Estado anterior: %s - Estado actual: %s ", tcb.TID, estado_anterior, tcb.Estado)
}

func Cambiar_estado_proceso(pcb *pcb.T_PCB, newState string) {
	//Sem_m_lista_procesos.Lock()  esto no se por que estaba aca jeje
	//defer Sem_m_lista_procesos.Unlock()

	estado_anterior := pcb.Estado
	pcb.Estado = newState
	color.Log_resaltado(color.BgWhite, "PROCESO PID: %d - Estado anterior: %s - Estado actual: %s ", pcb.PID, estado_anterior, pcb.Estado)
}
