package api

import (
	//"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"os"
	"time"
	//"strconv"

	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	//color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	estructuras "github.com/sisoputnfrba/tp-golang/filesystem/v.estructuras"
)

func Recibir_solicitud_de_mem_dump_memory(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)

	var solicitud globals.T_DumpMemoryRequest

	err := decoder.Decode(&solicitud)

	if err != nil {
		log.Printf("Error al decodificar solicitud: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar solicitud"))
		return
	}

	log.Printf("Me llego un solicitud de un proceso PID: %d TID: %d para un dump memory", solicitud.TID, solicitud.PID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))

	timestamp := time.Now().Format("20060102-150405")
	nombre_archivo := fmt.Sprintf("<%d>-<%d>-<%s>.dmp", solicitud.PID, solicitud.TID, timestamp)

	//cosas cuando mergee con memoria, ver tema de PARTICIOES

	estructuras.Crear_archivo_DUMP(solicitud.PID, solicitud.TID, nombre_archivo, uint32(solicitud.Tamanio), solicitud.Contenido)

	respuesta := <-globals.Sem_b_finalizo_dump_memory

	Enviar_respuesta_de_fs_dump_memory(respuesta)
}

func Enviar_respuesta_de_fs_dump_memory(respuesta uint32) {

	body, err := json.Marshal(respuesta)
	if err != nil {
		log.Printf("Error codificando respuesta a dump memory: %s", err.Error())
		return
	}

	//log.Printf("Enviando respuesta dump memory: %s", string(body))

	url := fmt.Sprintf("http://%s:%d/dumpMemoryFM", globals.Config_filesystem.Ip_memory, globals.Config_filesystem.Port_memory)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creando solicitud HTTP: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error enviando solicitud a %s:%d: %s", globals.Config_filesystem.Ip_memory, globals.Config_filesystem.Port_memory, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Solicitud a FS falló con el código de estado: %d", resp.StatusCode)
	} else {
		//log.Printf("SOlicitud de dump memory enviados a FS  correctamente")
	}

}
