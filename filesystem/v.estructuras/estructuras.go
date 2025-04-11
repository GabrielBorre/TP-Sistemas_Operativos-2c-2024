package vestructuras

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sisoputnfrba/tp-golang/filesystem/globals"
	color "github.com/sisoputnfrba/tp-golang/utils/color-log"
)

func Verifica_o_crea_archivos() error {

	mount_dir := globals.Config_filesystem.Mount_dir
	if _, err := os.Stat(mount_dir); os.IsNotExist(err) {
		log.Printf("El directorio %s no existe. Creando...\n", mount_dir)
		if err := os.MkdirAll(mount_dir, 0755); err != nil {
			return fmt.Errorf("Error al crear el directorio mount_dir: %v", err)
		}
	}

	bitmap_path := mount_dir + "/bitmap.dat"
	block_path := mount_dir + "/bloques.dat"

	if _, err := os.Stat(bitmap_path); os.IsNotExist(err) {
		log.Println("Creando archivo bitmap.dat...")
		if err := Crear_archivo_bitmap(bitmap_path); err != nil {
			return err
		}
	} else {
		log.Println("El archivo bitmap.dat ya existe.")
	}

	if _, err := os.Stat(block_path); os.IsNotExist(err) {
		log.Println("Creando archivo bloques.dat...")
		if err := Crear_archivo_bloques(block_path); err != nil {
			return err
		}
	} else {
		log.Println("El archivo bloques.dat ya existe.")
	}

	return nil
}

////////////////////////////////////////////////////// BITMAP /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func Crear_archivo_bitmap(path string) error {

	block_count := globals.Config_filesystem.Block_count
	// El tamaño del bitmap es BLOCK_COUNT / 8 bytes, redondeado hacia arriba
	bitmapSize := (block_count + 7) / 8
	data := make([]byte, bitmapSize)

	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error al crear bitmap.dat: %v", err)
	}
	return nil
}

func Leer_bitmap() ([]byte, error) {
	mount_dir := globals.Config_filesystem.Mount_dir
	bitmap_path := mount_dir + "/bitmap.dat"

	data, err := ioutil.ReadFile(bitmap_path)
	if err != nil {
		return nil, fmt.Errorf("Error al leer bitmap.dat: %v", err)
	}
	return data, nil
}

// después de que algún bloque pase de ocupado (1) a libre (0) o viceversa
func Actualizar_bitmap(block_num uint32, ocupado bool) error {
	mount_dir := globals.Config_filesystem.Mount_dir
	bitmap_path := mount_dir + "/bitmap.dat"

	//leemos el contenido actual del bitmap
	data, err := Leer_bitmap()
	if err != nil {
		return err
	}

	byteIndex := block_num / 8
	bitIndex := block_num % 8

	//cambiamos el bit correspondiente
	if ocupado {
		//marca el bloque como ocupado
		data[byteIndex] |= (1 << bitIndex)
	} else { //marca como libre, lo pone en 0
		data[byteIndex] &^= (1 << bitIndex)
	}

	err = ioutil.WriteFile(bitmap_path, data, 0644)
	if err != nil {
		return fmt.Errorf("Error al actualizar el bitmap.dat: %v", err)
	}

	log.Printf("Bloque %d marcado como %s en el bitmap.\n", block_num, map[bool]string{true: "ocupado", false: "libre"}[ocupado])
	return nil
}

func Bloque_ocupado(block_num uint32) (bool, error) {
	data, err := Leer_bitmap()
	if err != nil {
		return false, err
	}
	//se fija que byte y que bit dentro del byte hay que modificar
	byteIndex := block_num / 8
	bitIndex := block_num % 8

	ocupado := (data[byteIndex] & (1 << bitIndex)) != 0 // mira el estado del bit

	return ocupado, nil
}

////////////////////////////////////////////////////// BLOQUES /////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//Faltaría definir cuales van a ser los bloques de índice

func Crear_archivo_bloques(path string) error {
	block_count := globals.Config_filesystem.Block_count
	block_size := globals.Config_filesystem.Block_size
	// El tamaño del archivo bloques.dat es BLOCK_COUNT * BLOCK_SIZE
	blocksSize := block_count * block_size
	data := make([]byte, blocksSize)

	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error al crear bloques.dat: %v", err)
	}
	return nil
}

func Escribir_bloque(block_num uint32, data []byte) error {
	mount_dir := globals.Config_filesystem.Mount_dir
	block_path := mount_dir + "/bloques.dat"

	block_size := globals.Config_filesystem.Block_size
	block_count := uint32(globals.Config_filesystem.Block_count)

	// Verificar si el block_num está dentro de los límites
	if block_num >= block_count {
		return fmt.Errorf("Error: El número de bloque %d está fuera de los límites. El número máximo de bloques es %d.", block_num, block_count-1)
	}

	if len(data) < int(block_size) {
		paddedData := make([]byte, block_size)
		copy(paddedData, data)
		data = paddedData
	} else if len(data) > int(block_size) {
		data = data[:block_size]
	}
	file, err := os.OpenFile(block_path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("ERROR al abrir bloques.dat: %v", err)
	}
	defer file.Close()

	//se calcula el offset en el archivo
	offset := int64(block_num) * int64(block_size)

	//muevo el puntero al lugar adecuado del archivo
	_, err = file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("ERROR al mover el puntero de bloques.dat: %v", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("ERROR al escribir en bloques.dat: %v", err)
	}

	log.Printf("Bloque %d actualizado correctamente en bloques.dat.\n", block_num)
	return nil
}

func Leer_bloque(block_num uint32) ([]byte, error) {
	mount_dir := globals.Config_filesystem.Mount_dir
	block_path := mount_dir + "/bloques.dat"

	block_size := globals.Config_filesystem.Block_size
	block_count := uint32(globals.Config_filesystem.Block_count)

	if block_num >= block_count {
		return nil, fmt.Errorf("ERROR: el número de bloque %d está fuera de los límites. El número máximo de bloques es %d.", block_num, block_count-1)
	}

	file, err := os.Open(block_path)
	if err != nil {
		return nil, fmt.Errorf("ERROR al abrir bloques.dat: %v", err)

	}
	defer file.Close()

	offset := int64(block_num) * int64(block_size)

	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, fmt.Errorf("ERROR al mover el puntero de bloques.dat: %v", err)
	}

	//leemos los datos del bloque
	data := make([]byte, block_size)
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("ERROR al leer el bloque %d de bloques.dat: %v", block_num, err)
	}

	log.Printf("Bloque %d leído correctamente de bloques.dat.\n", block_num)
	return data, nil

}

func Acceso_bloque_con_delay(block_num uint32, block_type string, archivo string) {
	//simulo el retardo del acceso
	block_access_delay := time.Duration(globals.Config_filesystem.Block_access_delay) * time.Millisecond
	log.Printf("Acceso al bloque %d - Tipo: %s\n", block_num, block_type)
	bloques_libres, err := Contar_bloques_libres()
	if err != nil {
		log.Printf("Error al contar bloques libres: %v", err)
		bloques_libres = 0 // En caso de error, se asigna 0 para evitar que falle el log
	}
	color.Log_obligatorio("## Bloque asignado: %d - Archivo: %s - Bloques Libres: %d", block_num, archivo, bloques_libres)
	time.Sleep(block_access_delay)
	log.Println("Acceso completado.")
}

// ///////////////////////////////////// ARCHIVO METADATA /////////////////////////////////////////////////
func Crear_archivo_metadata(nombre_archivo string, indexBlock uint32, size uint32) error {
	mount_dir := globals.Config_filesystem.Mount_dir
	ruta_files := mount_dir + "/files/"
	metadataPath := ruta_files + nombre_archivo + ".metadata"

	if _, err := os.Stat(ruta_files); os.IsNotExist(err) {
		err = os.MkdirAll(ruta_files, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Error al crear el directorio %s: %v", ruta_files, err)
		}
		log.Printf("Directorio %s creado.\n", ruta_files)
	}

	metadata := map[string]interface{}{
		"index_block": indexBlock,
		"size":        size,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("ERROR al convertir el metadata a JSON: %v", err)
	}

	err = os.WriteFile(metadataPath, metadataJSON, 0644)
	if err != nil {
		return fmt.Errorf("ERROR al escribir el metadata: %v", err)
	}

	log.Printf("Archivo de metadata para %s creado correctamente.\n", nombre_archivo)
	return nil
}

func Leer_archivo_metadata(nombre_archivo string) (map[string]interface{}, error) {
	mount_dir := globals.Config_filesystem.Mount_dir
	metadataPath := mount_dir + "/files/" + nombre_archivo + ".metadata"

	metadata_bytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("ERROR al leer el archivo de metadata: %v", err)
	}

	var metadata map[string]interface{}
	err = json.Unmarshal(metadata_bytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("ERROR al parsear el JSON de metadata: %v", err)
	}

	log.Printf("Archivo de metadata para %s leído correctamente.\n", nombre_archivo)
	return metadata, nil
}

// ///////////////////////////////////////// CREACION ARCHIVOS /////////////////////////////////////////////////////////////
func Verificar_espacio_disponible(tamanio_requerido uint32) (bool, error) {

	bloques_libres, err := Contar_bloques_libres() // Captura los dos valores
	if err != nil {
		return false, fmt.Errorf("Error al contar los bloques libres: %v", err) // Maneja el error si ocurre
	}

	// Calcular los bloques necesarios, incluyendo el bloque de índice
	bloques_necesarios := (tamanio_requerido / globals.Tamanio_bloque) + 1 // +1 por el bloque de índice

	// Comparar si hay suficientes bloques libres
	return bloques_libres >= int(bloques_necesarios), nil
}

func Reservar_bloques(tamanio_requerido uint32) ([]uint32, error) {
	var bloques_reservados []uint32

	bloque_indice, err := Reservar_bloque_libre()
	if err != nil {
		return nil, fmt.Errorf("No se pudo reservar el bloque de índice")
	}
	bloques_reservados = append(bloques_reservados, bloque_indice)

	bloques_necesarios := (tamanio_requerido + globals.Tamanio_bloque - 1) / globals.Tamanio_bloque
	for i := 0; i < int(bloques_necesarios); i++ {
		bloque_datos, err := Reservar_bloque_libre()
		if err != nil {
			return nil, fmt.Errorf("No se pudo reservar bloque de datos")
		}
		bloques_reservados = append(bloques_reservados, bloque_datos)
	}

	return bloques_reservados, nil
}

func Crear_archivo_DUMP(pid uint32, tid uint32, nombre_archivo string, tamanio_requerido uint32, contenido []byte) error {

	espacio_suficiente, err := Verificar_espacio_disponible(tamanio_requerido)
	if err != nil {
		return fmt.Errorf("Error al verificar el espacio disponible: %v", err)
	}

	if !espacio_suficiente {
		// Notificar a memoria sobre el error
		log.Printf("## (%d:%d) - Error: No hay suficiente espacio para crear el archivo %s\n", pid, tid, nombre_archivo)
		return fmt.Errorf("No hay suficiente espacio para crear el archivo %s", nombre_archivo)
	}

	bloques_necesarios := int(tamanio_requerido) / globals.Config_filesystem.Block_size
	if tamanio_requerido%uint32(globals.Config_filesystem.Block_size) != 0 {
		bloques_necesarios++ //redondea para arriba si hay restos
	}

	bloques_libres, err := Contar_bloques_libres()
	if err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return err
	}

	if bloques_libres < bloques_necesarios+1 { //+1 por el bloque indice
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("No hay espacio suficiente en el sistema de archivos")
	}

	//reservar bloque de indide
	indice_bloque, err := Reservar_bloque_libre()
	if err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("Error reservando el bloque de índice (PID: %d, TID: %d): %v", pid, tid, err)
	}

	bloques_reservados := make([]uint32, 0, bloques_necesarios)
	for i := 0; i < bloques_necesarios; i++ {
		bloque, err := Reservar_bloque_libre()
		if err != nil {
			globals.Sem_b_finalizo_dump_memory <- 0
			return fmt.Errorf("Error reservando bloque de datos (PID: %d, TID: %d): %v", pid, tid, err)
		}
		bloques_reservados = append(bloques_reservados, bloque)
	}

	timestamp := time.Now().Unix()
	metadata_filename := fmt.Sprintf("%s-%d.dmp", nombre_archivo, timestamp)
	if err := Crear_archivo_metadata(metadata_filename, indice_bloque, tamanio_requerido); err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("Error creando archivo de metadata (PID: %d, TID: %d): %v", pid, tid, err)
	}
	//escribe los punteros en el bloque de indice
	for _, bloque := range bloques_reservados {
		puntero_data := make([]byte, 4) //suponiendo que son punteros de 4 bytes
		copy(puntero_data, []byte{byte(bloque >> 24), byte(bloque >> 16), byte(bloque >> 8), byte(bloque)})
		if err := Escribir_bloque(indice_bloque, puntero_data); err != nil {
			globals.Sem_b_finalizo_dump_memory <- 0
			return fmt.Errorf("ERROR escribiendo en el bloque de índice (PID: %d, TID: %d): %v", pid, tid, err)
		}
	}

	//escribe los datos en los bloques reservados
	for i, bloque := range bloques_reservados {
		offset := i * int(globals.Config_filesystem.Block_size)
		fin := offset + int(globals.Config_filesystem.Block_size)
		if fin > len(contenido) {
			fin = len(contenido)
		}

		if err := Escribir_bloque(bloque, contenido[offset:fin]); err != nil {
			globals.Sem_b_finalizo_dump_memory <- 0
			return fmt.Errorf("ERROR escribiendo en el bloque %d (PID: %d, TID: %d): %v", bloque, pid, tid, err)
		}
		//simulacion del retardo de acceso a bloque
		Acceso_bloque_con_delay(bloque, "DATA", metadata_filename)
	}
	color.Log_obligatorio("## Archivo Creado: %s - Tamaño: %d", metadata_filename, tamanio_requerido)

	color.Log_resaltado(color.Pink, "Archivo DUMP %s creado con éxito (PID: %d, TID: %d).\n", metadata_filename, pid, tid)

	globals.Sem_b_finalizo_dump_memory <- uint32(1)

	return nil

}

func Grabar_bloques(nombre_archivo string, bloques []uint32, contenido []byte) error {

	for i, bloque := range bloques {
		offset := i * int(globals.Tamanio_bloque)
		data := contenido[offset : offset+int(globals.Tamanio_bloque)]

		err := Escribir_bloque(bloque, data)
		if err != nil {
			globals.Sem_b_finalizo_dump_memory <- 0
			return fmt.Errorf("ERROR al escribir el bloque: %v", err)
		}
		//acá con el acceso tendría que entrar el delay
		color.Log_obligatorio("## Acceso Bloque - Archivo: %s - Tipo Bloque: DATOS - Bloque File System %d\n", nombre_archivo, bloque)
	}
	color.Log_obligatorio("## Fin de solicitud - Archivo: %s\n", nombre_archivo)
	return nil
}

func Crear_DUMP(pid uint32, tid uint32, nombre_archivo string, tamanio_requerido uint32, contenido []byte) error {
	espacio_disponible, err := Verificar_espacio_disponible(tamanio_requerido)
	if err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("ERROR al verificar espacio disponible: %v", err)
	}

	if !espacio_disponible {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("ERROR: Espacio insuficiente en el FileSystem")
	}

	err = Crear_archivo_DUMP(pid, tid, nombre_archivo, tamanio_requerido, contenido)
	if err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("ERROR al crear archivo DUMP: %v", err)
	}

	bloques_reservados, _ := Reservar_bloques(tamanio_requerido)
	err = Grabar_bloques(nombre_archivo, bloques_reservados, contenido)
	if err != nil {
		globals.Sem_b_finalizo_dump_memory <- 0
		return fmt.Errorf("ERROR al grabar bloques de archivo DUMP: %v", err)
	}

	return nil
}

// se encarga de contar los bloques libres del bitmap
func Contar_bloques_libres() (int, error) {
	data, err := Leer_bitmap()
	if err != nil {
		return 0, err
	}

	bloques_libres := 0
	for _, byte := range data {
		for i := 0; i < 8; i++ {
			if byte&(1<<i) == 0 {
				bloques_libres++
			}
		}
	}

	return bloques_libres, nil
}

//reserva el 1er bloque libre disponible

func Reservar_bloque_libre() (uint32, error) {
	data, err := Leer_bitmap()
	if err != nil {
		return 0, err
	}

	for byteIndex := 0; byteIndex < len(data); byteIndex++ {
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			if data[byteIndex]&(1<<bitIndex) == 0 {
				//encuentra el bloque y lo marca como ocupado
				block_num := uint32(byteIndex*8 + bitIndex)
				if err := Actualizar_bitmap(block_num, true); err != nil {
					return 0, err
				}
				return block_num, nil
			}
		}
	}
	return 0, fmt.Errorf("No hay bloques libres disponibles")
}
