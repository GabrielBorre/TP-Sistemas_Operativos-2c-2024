package pcb

///////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////ESTRUCTURAS ADMINISTRATIVAS PRINCIPALES///////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

// Estructura PCB que comparten tanto el kernel como el CPU
// El PCB conoce el TID y la prioridad de cada uno de sus hilos, por eso tiene una lista de TCBs

type T_PCB struct {
	PID               uint32         `json:"pid"`
	Estado            string         `json:"state"` // "NEW","READY","RUNNING","BLOCKED","EXIT", en mayuscula
	TCBs              []*T_TCB       `json:"tids"`  // es todo un tema...
	Mutexs            []*T_MUTEX     `json:"mutexes"`
	Base              uint32         `json:"base"`
	Limite            uint32         `json:"limite"`
	Tamanio           uint32         `json:"tamanio"`
	Ejecuciones       int            `json:"executions"`
	Resources         map[string]int `json:"resources"`
	RequestedResource string         `json:"requested_resource"`
	Contador_TIDS     int            `json:"contador_tids"`
}

// DEJEN EL PC por Separado, gracias

type T_TCB struct {
	PCB                              *T_PCB                 `json:"pcb"` //esto para mi hay que sacarlo de aca pero me da miedo causar un hiroshima en alguna parte del codigo xD
	PID                              uint32                 `json:"pid"`
	TID                              uint32                 `json:"tid"`
	Prioridad                        int                    `json:"prioridad"`
	Estado                           string                 `json:"estado"` //"NEW","READY","RUNNING","BLOCKED","EXIT", en mayuscula
	//PC                               uint32                 `json:"pc"`     // el PID es el PID del proceso al que pertencen
	Registros_CPU                    map[string]interface{} `json:"cpu_reg"`
	Quantum                          uint32                 `json:"quantum"`
	MotivoDesalojo                   string                 `json:"motivo_desalojo"`
	Instrucciones                    string                 `json:"instrucciones"` //archivo de instrucciones, podriamos crear un campo nuevo que sea una lista de string que representen las instrucciones a ejecutar de un hilo
	Hilos_bloqueados_por_thread_join []*T_TCB
	Hilos_bloqueados_por_io          []*T_TCB
	Lista_instrucciones_a_ejecutar   []string
	Contexto_ejecucion               T_contexto_ejecucion `json:"contexto_ejecucion"`
}

type T_MUTEX struct {
	Nombre                 string
	Esta_asignado          bool
	Hilo_utilizador        *T_TCB
	Lista_hilos_bloqueados []*T_TCB // cola de hilos BLOQUEADOS dentro del MUTEX
}

///Estructura que nos puede servir para pensar el algoritmo de cola multinivel

type T_Cola_Multinivel struct {
	Cola      []*T_TCB
	Prioridad int
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

/////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////ESTRUCTURAS ADMINISTRATIVAS SECUNDARIAS (LAS PODEMOS DESESTIMAR)////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

type Estado int

// podríamos usar un enum para modelar a los Estados de los procesos e hilos, pero tambien el estado puede ser un string
const (
	NEW Estado = iota
	READY
	RUNNING
	BLOCKED
	EXIT
)

func (e Estado) Nombre() string {
	switch e {
	case NEW:
		return "NEW"
	case READY:
		return "READY"
	case RUNNING:
		return "RUNNING"
	case BLOCKED:
		return "BLOCKED"
	case EXIT:
		return "EXIT"
	default:
		return "UNKNOWN"
	}
}

// porque en SUM/SET/SUB/JNZ lo estamos diferenciando del PC que tenemos dentro de TCB
func Tipo_reg(reg string) string {
	if reg == "PC" || reg == "AX" || reg == "BX" || reg == "CX" || reg == "DX" || reg == "EX" || reg == "FX" || reg == "GX" || reg == "HX" {
		return "uint32"
	} else {
		return "El tipo de dato no corresponde a un registro"
	}
}

var FlagDesalojo bool = false
