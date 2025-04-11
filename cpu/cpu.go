package main

import (
	"fmt"
	"log"
	"net/http"

	cpu_api "github.com/sisoputnfrba/tp-golang/cpu/API"
	globals "github.com/sisoputnfrba/tp-golang/cpu/globals"

	//datasend "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"

	//server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
	ciclo "github.com/sisoputnfrba/tp-golang/cpu/ciclo_instruccion"
)

func main() {

	//configura el logger
	logger.ConfigurarLogger("cpu.log")
	logger.LogfileCreate("cpu_debug.log")

	//cfg.VEnvCpu(nil, &globals.Configcpu.Port)
	//cfg.VEnvKernel(&globals.Configcpu.IP_kernel, &globals.Configcpu.Port_kernel)

	//configura el archivo de configuracion
	err := cfg.ConfigInit("config-cpu.json", &globals.Config_cpu)
	if err != nil {
		fmt.Printf("Error al cargar la configuracion %v", err)
	}
	log.Printf("Archivo de configuracion de CPU cargado")

	// Levanta el servidor
	go ServerStart(globals.Config_cpu.Port_cpu)

	log.Printf("Servidor cargado en CPU")

	ciclo.Ciclo_de_instruccion()

	select {}

}

func ServerStart(port int) {

	mux := http.NewServeMux()

	/* KERNEL */
	mux.HandleFunc("POST /hilos", cpu_api.Recibir_pid_y_tid) // Recibir confirmacion de CPU respecto TID y PID
	mux.HandleFunc("DELETE /interrupciones/{TID}/{PID}/{MOTIVO}", cpu_api.Manejar_interrupcion)

	/* MEMORIA */
	mux.HandleFunc("/contextoEjecucion", cpu_api.Recibir_contexto_ejecucion_memoria)
	mux.HandleFunc("/instruccion", cpu_api.Recibir_instruccion_de_memoria)
	mux.HandleFunc("/recuperacionPCB", cpu_api.Recibir_pcb_de_memoria)
	mux.HandleFunc("/recuperacionTCB", cpu_api.Recibir_tcb_de_memoria)
	mux.HandleFunc("/baseYlimite", cpu_api.Recibir_base_y_limite)
	mux.HandleFunc("/writeMem", cpu_api.Recibir_confirmacion_escritura)

	log.Printf("Servidor escuchando en puerto: %d", port)

	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), mux)
	if err != nil {
		panic(err)
	}

}
