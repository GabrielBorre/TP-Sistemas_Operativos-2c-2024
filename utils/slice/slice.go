package slice

import "sync"

/*
*

  - RemoveAtIndex: Remueve un elemento de un slice en base a su índice.

  - @param slice: Slice de cualquier tipo.

  - @param index: Índice del elemento a remover.
*/
func RemoveAtIndex[T any](slice *[]T, index int) T {
	element := (*slice)[index]
	*slice = append((*slice)[:index], (*slice)[index+1:]...)
	return element
}

/*
*

  - InsertAtIndex: Inserta un elemento en un slice en el índice proporcionado.

  - @param slice: Slice de cualquier tipo.

  - @param index: Índice donde se va a ingresar el elemento.

  - @param elem:  Elemento a ingresar.
*/
func InsertAtIndex[T any](slice *[]T, index int, elem T) {
	*slice = append((*slice)[:index], append([]T{elem}, (*slice)[index:]...)...)
}

func CopyAtIndex[T any](slice *[]T, index int) T {
	elem := (*slice)[index]
	return elem
}

/*
*

  - Pop: Remueve el último elemento de un slice

  - @param slice: Slice de cualquier tipo.

  - @return T: Último elemento del slice.
*/
func Pop[T any](slice *[]T) T {
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	return last
}

/*
*

  - Shift: Remueve el primer elemento de un slice

  - @param slice: Slice de cualquier tipo.

  - @return T: Primer elemento del slice.

  - ! Si el slice está vacío, devuelve un valor por defecto.
*/
func Shift[T any](slice *[]T) T {
	if len(*slice) == 0 {
		var zero T
		return zero
	}

	first := (*slice)[0]
	*slice = (*slice)[1:]
	return first
}

/*
*

  - Push: Agrega un elemento al final de un slice

  - @param slice: Slice de cualquier tipo.

  - @param elem: Elemento a agregar.
*/
func Push[T any](slice *[]T, elem T) {
	*slice = append(*slice, elem)
}

// ----------------------------- CON MUTEX INCLUIDOS

/*
*

  - RemoveAtIndex: Remueve un elemento de un slice en base a su índice.

  - @param slice: Slice de cualquier tipo.

  - @param sem:   Semáforo a utilizar.

  - @param index: Índice del elemento a remover.
*/
func RemoveAtIndexMutex[T any](slice *[]T, index int, sem *sync.Mutex) T {
	sem.Lock()
	element := (*slice)[index]
	*slice = append((*slice)[:index], (*slice)[index+1:]...)
	sem.Unlock()
	return element
}

/*
*

  - InsertAtIndex: Inserta un elemento en un slice en el índice proporcionado.

  - @param slice: Slice de cualquier tipo.

  - @param index: Índice donde se va a ingresar el elemento.

  - @param elem:  Elemento a ingresar.

  - @param sem:   Semáforo a utilizar.
*/
func InsertAtIndexMutex[T any](slice *[]T, index int, elem T, sem *sync.Mutex) {
	sem.Lock()
	*slice = append((*slice)[:index], append([]T{elem}, (*slice)[index:]...)...)
	sem.Unlock()
}

/*
*

  - PopMutex: Remueve el último elemento de un slice

  - @param slice: Slice de cualquier tipo.

  - @param sem:   Semáforo a utilizar.

  - @return T:    Último elemento del slice.
*/
func PopMutex[T any](slice *[]T, sem *sync.Mutex) T {
	sem.Lock()
	last := (*slice)[len(*slice)-1]
	*slice = (*slice)[:len(*slice)-1]
	sem.Unlock()
	return last
}

/*
*

  - Shift: Remueve el primer elemento de un slice

  - @param slice: Slice de cualquier tipo.

  - @param sem: Semáforo a utilizar.

  - @return T: Primer elemento del slice.

  - ! Si el slice está vacío, devuelve un valor por defecto.
*/
func ShiftMutex[T any](slice *[]T, sem *sync.Mutex) T {
	sem.Lock()
	if len(*slice) == 0 {
		var zero T
		sem.Unlock()
		return zero
	}

	first := (*slice)[0]
	*slice = (*slice)[1:]
	sem.Unlock()
	return first
}

/*
*

  - Push: Agrega un elemento al final de un slice

  - @param slice: Slice de cualquier tipo.

  - @param elem: Elemento a agregar.

  - @param sem: Semáforo a utilizar.
*/
func PushMutex[T any](slice *[]T, elem T, sem *sync.Mutex) {
	sem.Lock()
	*slice = append(*slice, elem)
	sem.Unlock()
}

/*
*

  - CopyFirst: Devuelve el primer elemento de un slice

  - @param slice: Slice de cualquier tipo.

  - @param sem: Semáforo a utilizar.

  - @return T: Primer elemento del slice. Retorna una direccion de memoria (&)

  - ! Si el slice está vacío, devuelve un valor por defecto.
*/
func CopyFirstMutex[T any](slice *[]T, sem *sync.Mutex) T {
	sem.Lock()
	if len(*slice) == 0 {
		var zero T
		sem.Unlock()
		return zero
	}

	first := (*slice)[0]
	// En lugar de modificar el slice, simplemente lo dejamos igual
	sem.Unlock()
	return first
}

/*
*

  - CopyAtIndex: Devuelve el elemento de cierto indice de un slice

  - @param slice: Slice de cualquier tipo.

  - @param index: Posicion requerida.

  - @param sem: Semáforo a utilizar.

  - @return T: Elemento en la posición índice del slice. Retorna una direccion de memoria (&)

  - ! Si el slice está vacío, devuelve un valor por defecto. ??
*/

func CopyAtIndexMutex[T any](slice *[]T, index int, sem *sync.Mutex) T {
	sem.Lock()
	elem := (*slice)[index]
	sem.Unlock()

	return elem
}
