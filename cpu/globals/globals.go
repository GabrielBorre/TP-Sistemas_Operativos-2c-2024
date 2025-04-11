package globals

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

/////////////////////////// ESTRUCTURAS ///////////////////////////

type t_config_cpu struct {
	IP_cpu             string `json:"ip_cpu"`
	Port_cpu           int    `json:"port_cpu"`
	Ip_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Ip_kernel          string `json:"ip_kernel"`
	Port_kernel        int    `json:"port_kernel"`
	Number_felling_tlb int    `json:"number_felling_tlb"`
	Algorithm_tlb      string `json:"algorithm_tlb"`
	Log_level          string `json:"log_level"`
}

type T_contexto_ejecucion struct {
	AX     uint32
	BX     uint32
	CX     uint32
	DX     uint32
	EX     uint32
	FX     uint32
	GX     uint32
	HX     uint32
	PC     uint32
	BASE   uint32
	LIMITE uint32
}

type T_Tid_y_pid struct {
	PID uint32 `json:"pid"`
	TID uint32 `json:"tid"`
}

type T_peticion_instruccion struct {
	PID uint32 `json:"pid"`
	TID uint32 `json:"tid"`
	PC  uint32 `json:"pc"`
}

// En teoria, Nahuel dijo que desde el kernel tenemos que mandar un verbo DELETE para desalojar un proceso, por lo que no podríamos usar el body para interrumpir al hilo
type T_interrupcion struct {
	InterruptionReason string `json:"InterruptionReason"`
	Pid                uint32 `json:"pid"`
	Tid                uint32 `json:"tid"`
	ExecutionNumber    int    `json:"execution_number"`
}

type T_base_y_limite struct {
	Base   uint32 `json:"base"`
	Limite uint32 `json:"limite"`
}

type T_syscall struct {
	Syscol string                 `json:"syscall"`
	PID    uint32                 `json:"pid"`
	TID    uint32                 `json:"tid"`
	Params map[string]interface{} `json:"params"`
}

// ///////////////////////// VARIABLES ///////////////////////////
var (
	Contexto_de_ejecucion  *T_contexto_ejecucion
	Config_cpu             *t_config_cpu
	CurrentJob             *pcb.T_PCB
	CurrentTCB             *pcb.T_TCB
	MemDelay               int
	Tid_y_pid              T_Tid_y_pid
	Peticion_instruccion   *T_peticion_instruccion
	Instruccion_de_memoria string
	Lista_interrupciones   []T_interrupcion
	Syscall_atendida       bool = true
	Hubo_seg_fault         bool
)

var MotivosDesalojo = map[string]struct{}{
	"EXIT":           {},
	"DUMP_MEMORY":    {},
	"IO":             {},
	"PROCESS_CREATE": {},
	"THREAD_CREATE":  {},
	"THREAD_JOIN":    {},
	"THREAD_CANCEL":  {},
	"MUTEX_CREATE":   {},
	"MUTEX_LOCK":     {},
	"MUTEX_UNLOCK":   {},
	"THREAD_EXIT":    {},
	"PROCESS_EXIT":   {},
}

// ///////////////////////// SEMAFOROS ///////////////////////////
var (
	// * Mutex

	MotivoDesalojoMutex              sync.Mutex
	OperationMutex                   sync.Mutex
	PCBMutex                         sync.Mutex
	TCB_Mutex                        sync.Mutex
	Sem_m_Lista_interrupciones_mutex sync.Mutex

	// * Binario
	PlanBinary                             = make(chan bool, 1)
	Sem_b_recepcion_contexto_ejecucion     = make(chan bool, 1)
	Sem_b_recepcion_instruccion_de_memoria = make(chan bool, 1)
	Sem_b_recepcion_tcb                    = make(chan bool, 1)
	Sem_b_recepcion_pcb                    = make(chan bool, 1)
	Sem_b_recuperacion_completa            = make(chan bool, 1)

	Sem_b_recepcion_tcb_para_ejecutar = make(chan bool, 1)

	//Base y Limite
	Sem_byl_read_write_mem = make(chan T_base_y_limite, 1)

	// * Contadores
	MultiprogrammingCounter = make(chan int, 10)
)
