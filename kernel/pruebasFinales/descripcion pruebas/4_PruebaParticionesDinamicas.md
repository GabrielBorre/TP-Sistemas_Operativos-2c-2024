# Prueba Particiones Dinamicas
Actividades
    1. Iniciar los módulos.
        a.Parámetros del Kernel
            archivo_pseudocodigo: MEM_DINAMICA_BASE
            tamanio_proceso: 128
    2.Esperar a que todos los procesos que ingresaron al sistema pasen al estado READY.

Resultados Esperados
    -Los procesos ingresan respetando la planificación de largo plazo y creando la partición correspondiente.

Configuración del sistema
--------------------------
    Kernel.config

    ALGORITMO_PLANIFICACION=CMN
    QUANTUM=500
--------------------------
    Memoria.config

    TAM_MEMORIA=1024
    RETARDO_RESPUESTA=200
    ESQUEMA=DINAMICAS
    ALGORITMO_BUSQUEDA=BEST
--------------------------
    FileSystem.config

    BLOCK_SIZE=32
    BLOCK_COUNT=4096
    RETARDO_ACCESO_BLOQUE=2500

    



  
