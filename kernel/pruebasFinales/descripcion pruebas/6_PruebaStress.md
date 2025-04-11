# Prueba Stress
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: THE_EMPTINESS_MACHINE
            tamanio_proceso: 16

Resultados Esperados
    -No hay esperas activas ni memory leaks

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=CMN
    QUANTUM=250
--------------------------
    Memoria.config

    TAM_MEMORIA=8192
    RETARDO_RESPUESTA=100
    ESQUEMA=DINAMICAS
    ALGORITMO_BUSQUEDA=BEST
--------------------------
    FileSystem.config

    BLOCK_SIZE=64
    BLOCK_COUNT=1024
    RETARDO_ACCESO_BLOQUE=250

    



  

