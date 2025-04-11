package main

import (
	"fmt"
	"os/exec"
	"time"
)

// ------- SI O SI HAY QUE HACER CTRL C en cada terminal cuando queremos cerrar sino se bloquea el VS--------
func ejecutarEnNuevaTerminal(directorio string, comando string, titulo string) {
	// Modificamos el comando para que incluya el título
	cmd := exec.Command("terminator", "--working-directory", directorio, "-e", comando, "--title", titulo)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error al ejecutar el comando en una nueva terminal: %v\n", err)
	}
}

func main() {
	// Llamamos a la función y le pasamos el título correspondiente
	ejecutarEnNuevaTerminal("/home/utnso/tp-2024-2c-GOlazo/filesystem", "go build filesystem.go && ./filesystem; exec bash", "Filesystem")
	time.Sleep(3 * time.Second)

	ejecutarEnNuevaTerminal("/home/utnso/tp-2024-2c-GOlazo/memoria", "go build memoria.go && ./memoria; exec bash", "Memoria")
	time.Sleep(3 * time.Second)

	ejecutarEnNuevaTerminal("/home/utnso/tp-2024-2c-GOlazo/cpu", "go build cpu.go && ./cpu; exec bash", "CPU")
	time.Sleep(3 * time.Second)

	ejecutarEnNuevaTerminal("/home/utnso/tp-2024-2c-GOlazo/kernel", "go build kernel.go && ./kernel; exec bash", "Kernel")
}
