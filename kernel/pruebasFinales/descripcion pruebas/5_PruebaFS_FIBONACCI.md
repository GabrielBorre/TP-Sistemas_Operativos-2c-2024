# Prueba FS - Fibonacci Sequence
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: PRUEBA_FS
            tamanio_proceso: 0
    2.Esperar a que todos los procesos finalicen.
    3.Volver a iniciar la prueba.

Resultados Esperados
    -Llega un momento en el cual no se pueden hacer mas DUMP porque se llena el FS.

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=CMN
    QUANTUM=500
--------------------------
    Memoria.config

    TAM_MEMORIA=2048
    RETARDO_RESPUESTA=200
    ESQUEMA=DINAMICAS
    ALGORITMO_BUSQUEDA=BEST
--------------------------
    FileSystem.config

    BLOCK_SIZE=32
    BLOCK_COUNT=200
    RETARDO_ACCESO_BLOQUE=500

    



  
