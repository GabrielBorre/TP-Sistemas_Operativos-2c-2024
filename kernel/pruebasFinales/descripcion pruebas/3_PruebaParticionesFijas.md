# Prueba Particiones Fijas
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: MEM_FIJA_BASE
            tamanio_proceso: 12
    2.Esperar a que todos los procesos que ingresaron al sistema ejecute la instrucción de IO.
    3.Cambiar el valor del ALGORITMO_BUSQUEDA de la Memoria a BEST y volver a ejecutar la prueba.
    4.Cambiar el valor del ALGORITMO_BUSQUEDA de la Memoria a WORST y volver a ejecutar la prueba.

Resultados Esperados
    -Para cada algoritmo los procesos son asignados a la partición que corresponde según el mismo.

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=CMN
    QUANTUM=3750
--------------------------
    Memoria.config

    TAM_MEMORIA=256
    RETARDO_RESPUESTA=1000
    ESQUEMA=FIJAS
    ALGORITMO_BUSQUEDA=FIRST
    PARTICIONES=[32,16,64,128,16]
--------------------------
    FileSystem.config

    BLOCK_SIZE=16
    BLOCK_COUNT=1024
    RETARDO_ACCESO_BLOQUE=2500

    



  
