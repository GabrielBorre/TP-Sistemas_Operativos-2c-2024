# Prueba Race Condition
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: RECURSOS_MUTEX_PROC
            tamanio_proceso: 32
    2.Esperar la finalización de los procesos.
    3.Cambiar el valor del QUANTUM del Kernel a 750 y volver a ejecutar la prueba.

Resultados Esperados
    -El valor final generado por los hilos sin mutex no es determinístico y puede no ser el esperado.
    -El valor final generado por los hilos con mutex es correcto y determinístico.

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=CMN
    QUANTUM=3750
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



    



  
