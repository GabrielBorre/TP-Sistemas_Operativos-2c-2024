# Prueba Planificación
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: PLANI_PROC
            tamanio_proceso: 32
    2.Esperar la finalización de los procesos.
    3.Cambiar el algoritmo de planificación a PRIORIDADES y volver a ejecutar.
    4.Cambiar el algoritmo de planificación a CMN y volver a ejecutar.

Resultados Esperados
    -Los procesos se ejecutan respetando el algoritmo elegido:
        FIFO: Los hilos se ejecutan secuencialmente en orden de creación.
        Prioridades: Los hilos se ejecutan secuencialmente según su prioridad.
        CMN: Los hilos se ejecutan alternadamente según su prioridad.

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=FIFO
    QUANTUM=1750
--------------------------
    Memoria.config

    TAM_MEMORIA=1024
    RETARDO_RESPUESTA=1000
    ESQUEMA=FIJAS
    ALGORITMO_BUSQUEDA=FIRST
    PARTICIONES=[32,32,32,32,32,32,32,32]
--------------------------
    FileSystem.config

    BLOCK_SIZE=16
    BLOCK_COUNT=1024
    RETARDO_ACCESO_BLOQUE=2500

    



  
