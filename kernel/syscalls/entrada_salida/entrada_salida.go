package entrada_salida

import (

	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
)

// esta rutina se va a ejecutar en una go routine porque es una syscall asincronica
func Entrada_salida() {
	for {

		//espero a que desde la funcion TratarSyscall me llegue un tcb para realizar la operacion de i/o
		tcb := <-globals.Sem_tcb_realizar_operacion_io
		//espero a que me llegue el tiempo para simular la operacion de i/0
		tiempo := <-globals.Sem_int_tiempo_milisegundos

		color.Log_obligatorio("## <PID: %d > <TID: %d > -Bloqueado por IO", tcb.PID, tcb.TID)

		//cambio el estado del TCB
		tcb.Estado = "BLOCKED"

		//simulo la operacion de IO, durmiendo el tiempo que me pasan por el canal
		time.Sleep(time.Duration(tiempo) * time.Millisecond)

		color.Log_obligatorio("## <PID: %d > <TID: %d > -Desbloqueado por IO", tcb.PID, tcb.TID)

		//aviso a enviar_cola_de_ready que tiene que enviar a la cola de Ready al tcb que realizo la operacion de IO
		globals.Sem_tcb_enviar_a_cola_de_Ready <- tcb

	}

}
