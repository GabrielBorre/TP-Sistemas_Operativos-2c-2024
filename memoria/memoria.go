package main

import (
	"fmt"
	"log"
	"net/http"

	memoria_api "github.com/sisoputnfrba/tp-golang/memoria/api"

	globals "github.com/sisoputnfrba/tp-golang/memoria/globals"
	inicializacion "github.com/sisoputnfrba/tp-golang/memoria/particiones"
	datasend "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
)

func main() {

	//configura el logger
	logger.ConfigurarLogger("memoria.log")
	logger.LogfileCreate("memoria_debug.log")

	//cfg.VEnvCpu(nil, &globals.Configcpu.Port)
	//cfg.VEnvKernel(&globals.Configcpu.IP_kernel, &globals.Configcpu.Port_kernel)

	//configura el archivo de configuracion
	err := cfg.ConfigInit("config-memoria.json", &globals.Config_memoria)
	if err != nil {
		fmt.Printf("Error al cargar la configuracion %v", err)
	}
	log.Println("Archivo de config de MEMORIA cargado!!!!!!!!!!!")

	Inicializar_memoria()

	// Levanta el servidor
	go ServerStart(globals.Config_memoria.Port_memoria)

	var mensaje string = "hola soy memoria :-)"
	datasend.EnviarMensaje(globals.Config_memoria.Ip_filesystem, globals.Config_memoria.Port_filesystem, mensaje)

	go memoria_api.Enviar_respuesta_solicitud_iniciar_proceso(globals.Config_memoria.Ip_kernel, globals.Config_memoria.Port_kernel)

	// Pausa indefinida para evitar que el programa termine
	select {}

}

func ServerStart(port int) {

	mux := http.NewServeMux()

	/* KERNEL */
	mux.HandleFunc("PUT /procesos", memoria_api.Recibir_solicitud_iniciar_proceso)
	mux.HandleFunc("PUT /hilos", memoria_api.Recibir_solicitud_iniciar_hilo)
	mux.HandleFunc("DELETE /finalizarProceso/{PID}/{Tamanio}", memoria_api.Recibir_solicitud_finalizar_proceso)
	mux.HandleFunc("DELETE /finalizarHilo/{TID}/{PID}", memoria_api.Recibir_solicitud_finalizar_hilo)
	mux.HandleFunc("/dumpMemoryMem", memoria_api.Recibir_solicitud_dump_memory)
	mux.HandleFunc("/dumpMemoryFM", memoria_api.Recibir_respuesta_dump_memory)
	mux.HandleFunc("PUT /compactar", memoria_api.Recibir_solicitud_y_compactar)

	/* CPU */
	mux.HandleFunc("/contextoEjecucion", memoria_api.Recibir_solicitud_contexto_ejecucion)
	mux.HandleFunc("/instruccion", memoria_api.Recibir_solicitud_instruccion)
	mux.HandleFunc("/recuperacionTCB", memoria_api.Recibir_solicitud_tcb)
	mux.HandleFunc("/recuperacionPCB", memoria_api.Recibir_solicitud_pcb)
	mux.HandleFunc("/actualizarContextoEjecucion", memoria_api.Recibir_actualizacion_contexto_ejecucion)
	mux.HandleFunc("/baseYlimite", memoria_api.Recibir_solicitud_base_y_limite)
	mux.HandleFunc("/writeMem", memoria_api.Escribir_valor)

	//mux.HandleFunc("GET /instrucciones ", memoria_api.Enviar_instrucciones)
	mux.HandleFunc("GET /instrucciones ", memoria_api.Recibir_solicitud_contexto_ejecucion)

	mux.HandleFunc("/read", memoria_api.Leer_valor)

	log.Printf("Servidor escuchando en puerto: %d", port)

	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), mux)
	if err != nil {
		panic(err)
	}
}

func Inicializar_memoria() {
	inicializacion.Declarar_memoria()
	inicializacion.Aplicar_esquema()
}
