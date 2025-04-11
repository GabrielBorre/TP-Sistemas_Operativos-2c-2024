package particiones

import (
	"log"

	"github.com/sisoputnfrba/tp-golang/memoria/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
	slice "github.com/sisoputnfrba/tp-golang/utils/slice"
)

func Buscar_particion(pid *int) *globals.Particion {

	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].ProcesoID != nil && *(globals.Particiones[i].ProcesoID) == *(pid) {
			log.Printf("La particion %d fue asignada al PID %d", i, *globals.Particiones[i].ProcesoID)
			return globals.Particiones[i]
		}
	}

	/*for i := 0; i < len(globals.Particiones); i++ {
		if *(globals.Particiones[i].ProcesoID) == *(pid) {
			log.Printf("se encontro %+v", globals.Particiones[i])

		}
	}*/
	log.Printf("La particion no se encontro en globals.Particiones, devolvemos nul")
	return nil
}

func Liberar_particion(pid *int) {
	particion := Buscar_particion(pid)
	particion.Libre = true
	particion.ProcesoID = nil
}

/*
- Diferenciar esquema
- Diferenciar estrategia y buscar hueco libre
	- fijas: el proceso no puede iniciarse
	- dinamicas: compactacion
		- si sigue sin poder, el proceso no puede iniciarse
*/

// @annotation Cuando creemos un proceso, llamamos a esta
func Recibir_pedido_creacion_particion(tamanio int, pid *int) { // CHEQUEAR estoy muy quemada
	//var valor_confirmacion int
	valor_confirmacion := -1
	switch globals.Config_memoria.Schema {
	case "FIJAS":
		particion := Buscar_segun_estrategia(tamanio)
		if particion != nil {
			color.Log_obligatorio("## Proceso <Creado> - PID: <PID: %d> - Tamaño: <Tamaño: %d> ", *pid, tamanio)
			color.Log_resaltado(color.Yellow, "La particion elegida es de tamaño: %d", particion.Tamanio)
			Asignar_particion_fija(particion, pid)
			valor_confirmacion = 1
			globals.Sem_int_inicializar_pcb <- valor_confirmacion

		} else {
			color.Log_resaltado(color.Red, "De %s al mundo. No hay memoria suficiente para inicializar el proceso.", globals.Config_memoria.Schema)
			valor_confirmacion = 0
			color.Log_resaltado(color.Yellow, "AA EL VALOR DE CONFIRMACION ENVIADO DESPUES ES %d ", valor_confirmacion)
			globals.Sem_int_inicializar_pcb <- valor_confirmacion
		}

	case "DINAMICAS":
		particion := Buscar_segun_estrategia(tamanio)
		if particion != nil {
			color.Log_obligatorio("## Proceso <Creado> - PID: <PID: %d> - Tamaño: <Tamaño: %d> ", *pid, tamanio)
			//color.Log_resaltado(color.Yellow, "El tamaño disponible en memoria es: %d", particion.Tamanio)

			Crear_particion_dinamica(tamanio, pid, particion)
			valor_confirmacion = 1
			globals.Sem_int_inicializar_pcb <- valor_confirmacion

		} else {
			color.Log_resaltado(color.Yellow, "Vamos a verificar si hay espacio disponible para compactar")
			tamanio_total_disponible := Tamanio_disponible()

			if tamanio_total_disponible >= tamanio {
				color.Log_resaltado(color.Cyan, "Soy recursante, me voy a compactar")
				valor_confirmacion = 2

				globals.Sem_int_inicializar_pcb <- valor_confirmacion // le  mandamos a kernel el 2 para que sepa que tiene que preparase para la compactacion   // esperamos que nos diga que está listo

				<-globals.Sem_b_compactacion_recibida_y_realizada

				color.Log_resaltado(color.Cyan, "Soy recursante, he sido compactado")
				Recibir_pedido_creacion_particion(tamanio, pid)

			} else {
				valor_confirmacion = 0
				color.Log_resaltado(color.Red, "De %s al mundo. No hay memoria suficiente para inicializar el proceso. Lo mato!", globals.Config_memoria.Schema)
				globals.Sem_int_inicializar_pcb <- valor_confirmacion
			}
		}
	}
}

func Tamanio_disponible() int {
	tamanio_disponible := 0
	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].Libre {
			tamanio_disponible += globals.Particiones[i].Tamanio
		}
	}
	return tamanio_disponible
}

func Asignar_particion_fija(particion_elegida *globals.Particion, pid *int) {
	// no hace falta pasarle el tamanio porque estaríamos modificando el tamanio de la partición fija
	// consigna mem dump: solicitar al módulo FileSystem la creación de un nuevo archivo con el *tamaño total de la memoria reservada por el proceso*
	// => no es necesario guardar hasta dónde ocupó de la partición
	log.Printf("Asignando particion fija para proceso: %d", *pid)
	particion_elegida.Libre = false
	particion_elegida.ProcesoID = pid
}

func Compactar() {
	lista_particiones_ocupadas := make([]*globals.Particion, 0) // 0???? si no tira index nil
	lista_particiones_libres := make([]*globals.Particion, 0)

	j := 0
	k := 0
	primera_particion := true
	proxima_base := 0
	tamanio_libre := 0

	//Creo una lista nueva con las particiones ocupadas de la lista de particiones globales y les asigno una nueva base y limite
	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].Libre {
			lista_particiones_libres[j] = slice.RemoveAtIndex(&globals.Particiones, i)
			tamanio_libre += lista_particiones_libres[j].Tamanio
			j++

		} else {
			lista_particiones_ocupadas[k] = slice.RemoveAtIndex(&globals.Particiones, i)

			if primera_particion {
				lista_particiones_ocupadas[k].Base = 0
				lista_particiones_ocupadas[k].Limite = lista_particiones_ocupadas[k].Tamanio - 1
				primera_particion = false
			} else {
				lista_particiones_ocupadas[k].Base = proxima_base
				lista_particiones_ocupadas[k].Limite = proxima_base + lista_particiones_ocupadas[k].Tamanio - 1
			}
			proxima_base += lista_particiones_ocupadas[k].Tamanio
			k++
		}
	}

	ultima_posicion_ocupadas := len(lista_particiones_ocupadas) - 1
	lista_particiones_libres[0].Base = lista_particiones_ocupadas[ultima_posicion_ocupadas].Limite + 1
	lista_particiones_libres[0].Limite = lista_particiones_libres[0].Base + tamanio_libre - 1
	lista_particiones_libres[0].Tamanio = tamanio_libre
	lista_particiones_libres[0].Libre = true
	lista_particiones_libres[0].ProcesoID = nil

	globals.Particiones = lista_particiones_ocupadas
	globals.Particiones = append(globals.Particiones, lista_particiones_libres[0])
}

func Compactarr() {

	var lista_particiones_ocupadas []*globals.Particion
	var suma_memoria_ocupada int = 0
	j := 0
	var h int = 0
	//Creo una lista nueva con las particiones ocupadas de la lista de particiones globales y les asigno una nueva base y limite
	for i := 0; i < len(globals.Particiones); i++ {
		if !globals.Particiones[i].Libre {
			suma_memoria_ocupada = globals.Particiones[i].Tamanio + suma_memoria_ocupada
			lista_particiones_ocupadas = append(lista_particiones_ocupadas, globals.Particiones[i])

		}

	}
	for ; j < len(lista_particiones_ocupadas); j++ {
		//me guardo la base y el limite de la particion antes de ser modificada
		base_anterior := lista_particiones_ocupadas[j].Base
		limite_anterior := lista_particiones_ocupadas[j].Limite
		if j == 0 {
			lista_particiones_ocupadas[j].Base = 0

		} else {
			lista_particiones_ocupadas[j].Base = lista_particiones_ocupadas[j-1].Limite + 1
		}
		lista_particiones_ocupadas[j].Limite = lista_particiones_ocupadas[j].Base + lista_particiones_ocupadas[j].Tamanio - 1

		//traslado la memoria de usuario a las particiones hacia el lugar que ahora le corresponde
		for k := base_anterior; k <= limite_anterior; k++ {
			globals.Memoria_usuario[h] = globals.Memoria_usuario[k]
			h++
		}

	}
	//creo la particion libre final que va a estar detras de todas las particiones ocupadas
	var particion_libre_final = &globals.Particion{
		Libre:  true,
		Base:   lista_particiones_ocupadas[j].Limite + 1,
		Limite: globals.Config_memoria.Memory_size - suma_memoria_ocupada - 1,
	}
	//creo la lista final de particiones (todas las particiones ocupadaas fueron compactadaas y al final hay una particion final libre)
	lista_particiones_compactadas := append(lista_particiones_ocupadas, particion_libre_final)
	globals.Particiones = lista_particiones_compactadas
}

////////////////////CHAQUEÑO
//EN el caso en que exista espacio contiguo en memoria para almacenar el proceso

func Crear_particion_dinamica(tamanio int, pid *int, particion_elegida *globals.Particion) {

	color.Log_resaltado(color.Yellow, "Creando particion de tamanio %d en base %d", tamanio, particion_elegida.Base)
	limite_proceso := particion_elegida.Base + tamanio - 1

	// CREAMOS UNA particion restante libre a partir de los valores originales de la particion que elegimos
	if tamanio < particion_elegida.Tamanio {
		nuevo_tamanio := particion_elegida.Tamanio - tamanio
		color.Log_resaltado(color.LightRed, "Fragmentacion externa. Creando partición libre de tamanio %d en base %d", nuevo_tamanio, limite_proceso)

		particion_libre := &globals.Particion{
			Base:      limite_proceso + 1,
			Limite:    limite_proceso + nuevo_tamanio, // se supone que coincide con particion_elegida.Limite
			Tamanio:   nuevo_tamanio,
			Libre:     true,
			ProcesoID: nil,
		}
		globals.Particiones = append(globals.Particiones, particion_libre)
		color.Log_resaltado(color.Blue, "Los limites deberían coincidir: %d y %d", particion_libre.Limite, particion_elegida.Limite)
	}

	//PISAMOS LA PARTICION ELEGIDA CON SUS NUEVOS VALORES
	particion_elegida.Limite = limite_proceso
	particion_elegida.Tamanio = tamanio
	particion_elegida.Libre = false
	particion_elegida.ProcesoID = pid
}

// ----------------------- ESTRATEGIAS DE BÚSQUEDA
func Buscar_segun_estrategia(tamanio_proceso int) *globals.Particion {
	//aca buscamos en las particiones dependiendo la estrategia
	estrategia := globals.Config_memoria.Search_algorithm
	//log.Printf("Buscando partición según estrategia %s", estrategia)
	switch estrategia {
	case "FIRST":
		//log.Printf("Se entró a First")
		particion := Buscar_first(tamanio_proceso)
		return particion
	case "BEST":
		//log.Printf("Se entró a Best")
		particion := Buscar_best(tamanio_proceso)
		return particion
	case "WORST":
		//log.Printf("Se entró a Worst")
		particion := Buscar_worst(tamanio_proceso)
		return particion
	default:
		color.Log_resaltado(color.Red, "No se encontro la estrategia %s", estrategia)
		return nil
	}
}

func Buscar_first(tamanio_proceso int) *globals.Particion {

	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].Libre && globals.Particiones[i].Tamanio >= tamanio_proceso {
			log.Printf("Se encontró una partición libre de tamaño %d en base %d", globals.Particiones[i].Tamanio, globals.Particiones[i].Base)
			return globals.Particiones[i]
		}
	}

	log.Printf("No se encontró partición libre de tamaño %d", tamanio_proceso)
	return nil
}

func Buscar_best(tamanio_proceso int) *globals.Particion {
	var mejor_particion *globals.Particion
	mejor_particion = Buscar_primera_particion_valida(tamanio_proceso)

	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].Libre && globals.Particiones[i].Tamanio < mejor_particion.Tamanio && globals.Particiones[i].Tamanio >= tamanio_proceso {
			mejor_particion = globals.Particiones[i]
		}
	}
	return mejor_particion
}

func Buscar_worst(tamanio_proceso int) *globals.Particion {
	var peor_particion *globals.Particion

	for i := 0; i < len(globals.Particiones); i++ {
		if globals.Particiones[i].Libre && globals.Particiones[i].Tamanio > tamanio_proceso {
			peor_particion = globals.Particiones[i]
			break

		}
	}
	return peor_particion

}

func Buscar_primera_particion_valida(tamanio_proceso int) *globals.Particion {
	var particion *globals.Particion = nil
	encontrado := false
	log.Printf("Buscando la partición inicial")
	for i := 0; i < len(globals.Particiones) && !encontrado; i++ {

		if globals.Particiones[i].Libre && globals.Particiones[i].Tamanio >= tamanio_proceso {
			particion = globals.Particiones[i]
			encontrado = true
		}
	}
	if particion == nil {
		return nil
	}
	return particion
}

func Liberar_particion_asociada(pid int) {
	var particion_encontrada *globals.Particion
	var indice int = 0
	//recorro la lista de particiones y busco la particion asociada por eel pid
	for ; indice < len(globals.Particiones); indice++ {
		if !globals.Particiones[indice].Libre && *globals.Particiones[indice].ProcesoID == pid {
			particion_encontrada = globals.Particiones[indice]
			break
		}
	}
	particion_encontrada.Libre = true
	particion_encontrada.ProcesoID = nil

	//esto solo ocurre si las particiones son dinamicas, lo de arriba para ambos tipos de particiones
	if globals.Config_memoria.Schema == "DINAMICAS" {
		//si la particion elegida no es la ultima y a la derecha hay una particion libre, las uno
		if indice != len(globals.Particiones)-1 && globals.Particiones[indice+1].Libre {
			//el limite de la particion va a ser el limite de la particion a la derecha
			globals.Particiones[indice].Limite = globals.Particiones[indice+1].Limite
			//el tamaño de la particion nueva va a ser la suma del tamaño de ambas particiones
			globals.Particiones[indice].Tamanio = globals.Particiones[indice].Tamanio + globals.Particiones[indice+1].Tamanio
			globals.Particiones[indice+1] = nil
			//remuevo la particion de la derecha
			slice.RemoveAtIndex(&globals.Particiones, indice+1)

		}
		if indice != 0 && globals.Particiones[indice-1].Libre {
			globals.Particiones[indice].Base = globals.Particiones[indice-1].Base
			globals.Particiones[indice].Tamanio += globals.Particiones[indice-1].Tamanio
			globals.Particiones[indice-1] = nil
			slice.RemoveAtIndex(&globals.Particiones, indice-1)

		}

	}
}

func Finalizar_hilos_asociados_al_proceso(pid int) {

	for i := 0; i < len(globals.Lista_hilos_creados); i++ {
		if globals.Lista_hilos_creados[i].PID == uint32(pid) {
			globals.Lista_hilos_creados[i] = nil
			slice.RemoveAtIndex(&globals.Lista_hilos_creados, i)
		}
	}

}

func Devolver_particion_asociada_al_proceso(pid int) *globals.Particion {
	for i := 0; i < len(globals.Particiones); i++ {
		if !globals.Particiones[i].Libre && *globals.Particiones[i].ProcesoID == pid {
			return globals.Particiones[i]
		}
	}
	return nil

}
