package globals

//"github.com/sisoputnfrba/tp-golang/utils/pcb"

type t_config_fs struct {
	Ip_filesystem      string `json:"ip_filesystem"`
	Port_filesystem    int    `json:"port_filesystem"`
	Ip_memory          string `json:"ip_memory"`
	Port_memory        int    `json:"port_memory"`
	Mount_dir          string `json:"mount_dir"`
	Block_size         int    `json:"block_size"`
	Block_count        int    `json:"block_count"`
	Block_access_delay int    `json:"block_access_delay"`
	Log_level          string `json:"log_level"`
}

type T_DumpMemoryRequest struct {
	TID       uint32 `json:"tid"`
	PID       uint32 `json:"pid"`
	Contenido []byte `json:"contenido"`
	Tamanio   int    `json:"tamanio"`
}

var Config_filesystem *t_config_fs
var Tamanio_bloque uint32

////////////////////////////////// SEMAFOROS //////////////////////////////////

var (
	// Enteros
	Sem_b_finalizo_dump_memory = make(chan uint32, 1)
)
