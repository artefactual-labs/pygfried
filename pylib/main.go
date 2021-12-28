package main

// #include <stdlib.h>
// #include <Python.h>
// extern int Pygfried_PyArg_ParseTuple_U(PyObject*, PyObject**);
// extern PyObject* Pygfried_Py_RETURN_NONE();
// extern PyObject* Pygfried_GoError;
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/artefactual-labs/pygfried"
)

func raise(err error) *C.PyObject {
	tp := C.Pygfried_GoError
	cstr := C.CString(err.Error())
	C.PyErr_SetString(tp, cstr)
	C.free(unsafe.Pointer(cstr))
	return nil
}

func stringToPy(s string) *C.PyObject {
	cstr := C.CString(s)
	pystr := C.PyUnicode_FromString(cstr)
	C.free(unsafe.Pointer(cstr))
	return pystr
}

func stringToPyOrNone(s string) *C.PyObject {
	if s == "" {
		return C.Pygfried_Py_RETURN_NONE()
	} else {
		return stringToPy(s)
	}
}

func goStringFromArgs(args *C.PyObject) (string, error) {
	var obj *C.PyObject
	if C.Pygfried_PyArg_ParseTuple_U(args, &obj) == 0 {
		return "", fmt.Errorf("Failed to parse arguments")
	}
	bytes := C.PyUnicode_AsUTF8String(obj)
	ret := C.GoString(C.PyBytes_AsString(bytes))
	C.Py_DecRef(bytes)
	return ret, nil
}

//export identify
func identify(self *C.PyObject, args *C.PyObject) *C.PyObject {
	path, err := goStringFromArgs(args)
	if err != nil {
		return nil
	}

	res, err := pygfried.Identify(path)
	if err != nil {
		return raise(err)
	}

	if len(res.Identifiers) == 0 {
		return C.Pygfried_Py_RETURN_NONE()
	}

	return stringToPyOrNone(res.Identifiers[0])
}

//export version
func version(self *C.PyObject) *C.PyObject {
	v := pygfried.Version()

	return stringToPyOrNone(v)
}

func main() {}
