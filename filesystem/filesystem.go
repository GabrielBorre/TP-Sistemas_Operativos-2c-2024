package main

import (
	"fmt"
	"log"
	"net/http"
	//"strconv"

	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	estructuras "github.com/sisoputnfrba/tp-golang/filesystem/v.estructuras"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	fs_api "github.com/sisoputnfrba/tp-golang/filesystem/api"
		
)

func main() {

	// Configura el logger
	logger.ConfigurarLogger("filesystem.log")
	logger.LogfileCreate("fs_debug.log")

	// Configura el archivo de configuración
	err := cfg.ConfigInit("config-filesystem.json", &globals.Config_filesystem)
	if err != nil {
		fmt.Printf("Error al cargar la configuracion %v", err)
	}
	log.Println("Archivo de config de FILESYSTEM cargado!!!!!!!!!!!")
	
	globals.Tamanio_bloque = uint32(globals.Config_filesystem.Block_size)

	// Verifica si existen los archivos bitmap.dat y bloques.dat, si no los crea
	err = estructuras.Verifica_o_crea_archivos()
	if err != nil {
		log.Fatalf("Error al verificar/crear archivos de estructura: %v", err)
	}
	log.Println("Archivos de estructura verificados/creados correctamente.")

	// Ejemplo de prueba: Creación de archivo dump y manejo de bloques
	//ejecutarPruebaFilesystem()

	// Levanta el servidor
	go ServerStart(globals.Config_filesystem.Port_filesystem)

	// Pausa indefinida para evitar que el programa termine
	select {}
}

func ejecutarPruebaFilesystem() {
	// Valores hardcodeados para la prueba
	pid := uint32(1234) // Ejemplo: ID de proceso
	tid := uint32(1)    // Ejemplo: ID de hilo
	nombreArchivo := "archivo_test"
	tamanioRequerido := uint32(128) // Ejemplo: 128 bytes de tamaño de archivo

	// Crear contenido dummy para escribir en el archivo (tamaño = tamanioRequerido)
	contenido := make([]byte, tamanioRequerido)
	for i := 0; i < int(tamanioRequerido); i++ {
		contenido[i] = byte(i % 256) // Solo un ejemplo de contenido
	}

	// Crear el archivo dump con bloques reservados
	err := estructuras.Crear_archivo_DUMP(pid, tid, nombreArchivo, tamanioRequerido, contenido)
	if err != nil {
		log.Fatalf("Error al crear archivo dump: %v", err)
	}
	log.Println("Archivo dump creado correctamente.")

	// Leer y mostrar el contenido del bloque
	bloque_num := uint32(0) // Ejemplo: leer el bloque 0
	bloque_data, err := estructuras.Leer_bloque(bloque_num)
	if err != nil {
		log.Fatalf("Error al leer el bloque %d: %v", bloque_num, err)
	}
	fmt.Printf("Contenido del bloque %d: %v\n", bloque_num, bloque_data)

	// Marcar un bloque como ocupado
	err = estructuras.Actualizar_bitmap(bloque_num, true)
	if err != nil {
		log.Fatalf("Error al actualizar el bitmap: %v", err)
	}
	log.Printf("Bloque %d marcado como ocupado.", bloque_num)

	// Verificar si el bloque está ocupado
	ocupado, err := estructuras.Bloque_ocupado(bloque_num)
	if err != nil {
		log.Fatalf("Error al verificar el estado del bloque %d: %v", bloque_num, err)
	}
	log.Printf("El bloque %d está ocupado: %v", bloque_num, ocupado)
}

func ServerStart(port int) {

	mux := http.NewServeMux()

	/* MEMORIA */
	mux.HandleFunc("/dumpMemoryFs", fs_api.Recibir_solicitud_de_mem_dump_memory)

	log.Printf("Servidor escuchando en puerto: %d", port)

	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), mux)
	if err != nil {
		panic(err)
	}
}
