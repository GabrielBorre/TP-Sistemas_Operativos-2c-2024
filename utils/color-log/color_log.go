package color_log

import "log"

// Definir colores usando códigos de escape ANSI
const (
	Black     = "\033[30m"
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Purple    = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	Orange    = "\033[38;5;208m" // Color ANSI aproximado a naranja
	Pink      = "\033[38;5;205m" // Color ANSI aproximado a rosa
	LightBlue = "\033[38;5;81m"  // Color ANSI aproximado a azul claro

	// Titilantes
	LightPurple = "\033[35;5;" // a rezar que sea
	LightRed    = "\033[31;5;" // a rezar que sea
	LightOrange = "\033[38;5;" // Color ANSI aproximado a azul claro

	// Pruebas
	YellowPrueba = "\033[33;5;82m" // a rezar que sea

	// Negrita
	BoldBlack  = "\033[30;1m"
	BoldRed    = "\033[31;1m"
	BoldGreen  = "\033[32;1m"
	BoldYellow = "\033[33;1m"
	BoldBlue   = "\033[34;1m"
	BoldPurple = "\033[35;1m"
	BoldCyan   = "\033[36;1m"
	BoldWhite  = "\033[37;1m"

	// Colores de fondo
	BgBlack  = "\033[40m" + BoldYellow
	BgRed    = "\033[41m" + BoldBlack
	BgGreen  = "\033[42m" + BoldRed
	BgYellow = "\033[43m" + BoldPurple
	BgBlue   = "\033[44m" + BoldYellow
	BgPurple = "\033[45m" + BoldYellow
	BgCyan   = "\033[46m" + BoldBlack
	BgWhite  = "\033[47m" + BoldBlack
)

/*
Ejemplo: color.Log_obligatorio("##(<PID>:<%d>) (<TID>: <%d>) - Solicitó syscall: <%s> ", globals.Nueva_syscall.TID, globals.Nueva_syscall.TID, globals.Nueva_syscall.Syscall)
Siempre sale en verde
*/
func Log_obligatorio(mensaje string, args ...interface{}) {
	// Si no hay argumentos, simplemente usa log.Println
	formato := Green + mensaje + Reset
	if len(args) == 0 {
		log.Println(formato)
	} else {
		// Si hay argumentos, usa log.Printf
		log.Printf(formato, args...)
	}
}

/*
Ejemplo: color.Log_error("Error porque paso algo xD")
Siempre sale en rojo
*/
func Log_error(mensaje string, args ...interface{}) {
	// Si no hay argumentos, simplemente usa log.Println
	formato := Red + mensaje + Reset
	if len(args) == 0 {
		log.Printf(formato)
	} else {
		// Si hay argumentos, usa log.Printf
		log.Printf(formato, args...)
	}
}

/*
Ejemplo: color.Log_resaltado(color.Blue, "Petición para crear un proceso a memoria enviada correctamente")
Y se pone el color que quieras de la lista
*/
func Log_resaltado(color string, mensaje string, args ...interface{}) {
	// Si no hay argumentos, simplemente usa log.Println
	formato := color + mensaje + Reset
	if len(args) == 0 {
		log.Println(formato)
	} else {
		// Si hay argumentos, usa log.Printf
		log.Printf(formato, args...)
	}
}

/*
//EJEMPLOS DE COMO SE ESCRIBE EN COLORES:
	log.Println(Red + "Este es un mensaje en rojo" + Reset)
	log.Println(Green + "Este es un mensaje en verde" + Reset)
	log.Println(Yellow + "Este es un mensaje en amarillo" + Reset)
	log.Println(Blue + "Este es un mensaje en azul" + Reset)
	log.Println(Purple + "Este es un mensaje en púrpura" + Reset)
	log.Println(Cyan + "Este es un mensaje en cian" + Reset)
	log.Println(White + "Este es un mensaje en blanco" + Reset)
*/
