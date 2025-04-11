package globals

import (
	"sync"

	"github.com/sisoputnfrba/tp-golang/utils/pcb"
)

////////////////////////////////// ESTRUCTURAS //////////////////////////////////

type T_config_memoria struct {
	Memory_size       int    `json:"memory_size"`
	Instructions_path string `json:"instruction_path"`
	Delay_response    int    `json:"response_delay"`
	Ip_kernel         string `json:"ip_kernel"`
	Port_kernel       int    `json:"port_kernel"`
	Ip_cpu            string `json:"ip_cpu"`
	Port_cpu          int    `json:"port_cpu"`
	Ip_memoria        string `json:"ip_memoria"`
	Port_memoria      int    `json:"port_memoria"`
	Ip_filesystem     string `json:"ip_filesystem"`
	Port_filesystem   int    `json:"port_filesystem"`
	Schema            string `json:"schema"`
	Search_algorithm  string `json:"search_algorithm"`
	Partitions        []int  `json:"partitions"`
	Log_level         string `json:"log_level"`
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

type T_base_y_limite struct {
	Base   uint32 `json:"base"`
	Limite uint32 `json:"limite"`
}

type Particion struct {
	Base      int
	Limite    int
	Tamanio   int
	Libre     bool
	ProcesoID *int // Puede ser nulo si no hay proceso asignado
}

type T_DumpMemoryRequest struct {
	TID       uint32 `json:"tid"`
	PID       uint32 `json:"pid"`
	Contenido []byte `json:"contenido"`
	Tamanio   int    `json:"tamanio"`
}

////////////////////////////////// VARIABLES //////////////////////////////////

var (
	Config_memoria *T_config_memoria
	//Contexto_de_ejecucion     T_contexto_ejecucion
	Nueva_solicitud_Tid_y_pid *T_Tid_y_pid
	Peticion_instruccion      *T_peticion_instruccion

	//Listas
	Lista_procesos_creados []*pcb.T_PCB //lista de todos los procesos creados en la memoria
	Lista_hilos_creados    []*pcb.T_TCB //lista de todos los hilos creados en la memoria

	Memoria_usuario []byte
	Particiones     []*Particion
)

////////////////////////////////// SEMAFOROS //////////////////////////////////

var (
	// Mutex

	// Canal que manda un 1 si el proceso se pudo inicializar, un 0 si no pudo o un 2 si podría pero debe compactar
	Sem_int_inicializar_pcb = make(chan int, 1)

	Sem_m_esperar_pedido_kernel_compactacion = make(chan bool, 1)

	//Mutex
	Sem_mutex_lista_hilos sync.Mutex

	// Enteros
	Sem_b_finalizo_dump_memory              = make(chan uint32, 1)
	Sem_b_compactacion_recibida_y_realizada = make(chan bool, 1)
)
