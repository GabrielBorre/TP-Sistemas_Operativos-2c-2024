package api

import (
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	particiones "github.com/sisoputnfrba/tp-golang/memoria/particiones"
	//color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	//"github.com/sisoputnfrba/tp-golang/utils/slice"
)

func Enviar_dump_memory_a_fs(solicitud globals.T_DumpMemoryRequest) {

	j := 0

	pidInt := int(solicitud.PID) // Convierte `uint32` a `int`
	pid := &pidInt

	particion := particiones.Buscar_particion(pid)
	
	solicitud.Tamanio = particion.Limite - particion.Base + 1

	solicitud.Contenido=make([]byte, solicitud.Tamanio)

	for i := particion.Base; i <= particion.Limite; i++ {
		solicitud.Contenido[j] = globals.Memoria_usuario[i]
		j++

	}
	
	body, err := json.Marshal(solicitud)
	if err != nil {
		log.Printf("Error codificando respuesta a dump memory: %s", err.Error())
		return
	}

	//log.Printf("Enviando respuesta dump memory: %s", string(body))

	url := fmt.Sprintf("http://%s:%d/dumpMemoryFs", globals.Config_memoria.Ip_filesystem, globals.Config_memoria.Port_filesystem)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_memoria.Ip_filesystem, globals.Config_memoria.Port_filesystem, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a FS falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("SOlicitud de dump memory enviados a FS  correctamente")
	}

}

func Recibir_respuesta_dump_memory(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var respuesta uint32
	err := decoder.Decode(&respuesta)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	globals.Sem_b_finalizo_dump_memory <- respuesta
}
