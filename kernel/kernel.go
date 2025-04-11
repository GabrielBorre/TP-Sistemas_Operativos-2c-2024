package main

import (
	"fmt"
	"log"
	"net/http"

	kernel_api "github.com/sisoputnfrba/tp-golang/kernel/API"
	globals "github.com/sisoputnfrba/tp-golang/kernel/globals"
	kernel_io_syscalls "github.com/sisoputnfrba/tp-golang/kernel/syscalls/entrada_salida"
	kernel_utils "github.com/sisoputnfrba/tp-golang/kernel/utils"
	kernel_utils_syscalls "github.com/sisoputnfrba/tp-golang/kernel/utils/kernel_syscalls"

	//datasend "github.com/sisoputnfrba/tp-golang/utils/client-Functions"
	cfg "github.com/sisoputnfrba/tp-golang/utils/config"
	logger "github.com/sisoputnfrba/tp-golang/utils/log"
	//server "github.com/sisoputnfrba/tp-golang/utils/server-Functions"
)

func main() {

	//configura el logger
	logger.ConfigurarLogger("kernel.log")
	logger.LogfileCreate("kernel_debug.log")

	//conigura el archivo de configuracion
	err := cfg.ConfigInit("config-kernel.json", &globals.Config_kernel)
	if err != nil {
		log.Printf("Error al cargar la configuracion %v", err)
	}
	log.Println("Archivo de config de KERNEL cargado")

	// Levanta el servidor
	go ServerStart(globals.Config_kernel.Port_kernel)

	/*Dios bendiga a Beethoven y a Oscar Peterson*/

	// Definimos variables
	globals.Estado_planificacion = "DETENIDA"
	//globals.Sem_binario_plani_largo_plazo <- false
	kernel_utils.Crear_proceso_inicial("/home/utnso/tp-2024-2c-GOlazo/kernel/pruebasFinales/THE_EMPTINESS_MACHINE.txt", 16, 0)
	//log.Printf("LARGO PLAZO: Creamos el proceso inicial, Esperando nuevo proceso")

	// Arranca la planificacion
	//go kernel_utils.Largo_plazo()
	go kernel_utils.Corto_plazo() //prod.liveshare.vsengsaas.visualstudio.com/join?064932E116762C3568799FABAB63CE394F11

	go kernel_utils_syscalls.Tratar_syscall()

	//rutina que se activa cuando hay que interrumpir un hilo que se encuentra ejecutando en CPU (solo aplica para algoritmos PRIORIDADES Y COLAS MULTINIVEL)
	go kernel_api.Enviar_interrupcion_a_CPU()

	if globals.Config_kernel.Algoritmo_planificacion == "CMN" {
		go kernel_api.Round_robin()
	}

	//rutina que se activa cuando hay que pedirle a memoria que inicialice un proceso
	go kernel_utils.Pedido_a_memoria_para_inicializacion()

	//rutina que se activa cuando se pudo inicializar un proceso con su hilo main.
	go kernel_utils.Enviar_a_cola_de_ready()
	//envia un hilo a la cola de Ready de acuerdo al algoritmo que usemos

	//rutina que se activa cuando  hay un THREAD_EXIT o THREAD_CANCEL
	//se puede ejecutar concurrentemente con la logica de la finalizacion del hilo, por eso puede ir en una rutina
	go kernel_api.Peticion_a_memoria_para_finalizar_hilo()

	go kernel_api.Peticion_a_memoria_para_finalizar_proceso()

	go kernel_io_syscalls.Entrada_salida()

	// Pausa indefinida para evitar que el programa termine
	select {}

}

func ServerStart(port int) {

	mux := http.NewServeMux()

	/*
		GET: Para obtener información (consulta).
		POST: Para enviar datos y crear un nuevo recurso.
		PUT: Para actualizar o reemplazar un recurso existente.
		DELETE: Para eliminar un recurso.
	*/

	/* MEMORIA */
	mux.HandleFunc("/procesos", kernel_api.Recibir_solicitud_iniciar_proceso) // Respuesta de memoria para inicializar proceso
	mux.HandleFunc("/dumpMemoryMem", kernel_api.Recibir_respuesta_dump_memory)
	//mux.HandleFunc("/procesos", kernel_api.Recibir_solicitud_finalizar_proceso) //Respuesta de memoria para finalizar proceso

	/* CPU */
	mux.HandleFunc("/confirmaciones", kernel_api.Confirmacion_recepcion_TID_y_PID) // Recibir confirmacion de CPU respecto TID y PID. No hace falta crear un API por cada respuesta a una solicitud que vos haces, lo podes hacer secuencial
	mux.HandleFunc("/dispatch", kernel_api.Recibir_syscall_CPU)
	mux.HandleFunc("/interrupciones", kernel_api.Recibir_interrupcion_CPU)

	// Respuesta de CPU de PCB recibido
	// Respuesta de CPU para ejecutar proceso
	// Respuesta de CPU para desalojar proceso

	/* PLANIFICACION */
	// Cuando arranca la plani
	// Cuando se detiene le plani

	log.Printf("Servidor escuchando en puerto: %d\n", port)

	err := http.ListenAndServe(":"+fmt.Sprintf("%v", port), mux)
	if err != nil {
		panic(err)
	}
}
