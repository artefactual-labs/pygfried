package main

// #include <stdlib.h>
// #include <Python.h>
// extern int Pygfried_PyArg_ParseTuple_U(PyObject*, PyObject**);
// extern int Pygfried_PyArg_ParseTuple_Oi(PyObject*, PyObject**, int*);
// extern int Pygfried_PyArg_ParseTuple_Uiii(PyObject*, PyObject**, int*, int*, int*);
// extern int Pygfried_PyBytes_AsStringAndSize(PyObject*, char**, Py_ssize_t*);
// extern PyObject* Pygfried_Py_RETURN_NONE();
// extern PyObject* Pygfried_GoError;
// extern PyObject* Pygfried_json_loads(PyObject*);
import "C"

import (
	"fmt"
	"strings"
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

func goPathStringFromPyUnicode(obj *C.PyObject) (string, error) {
	bytes := C.PyUnicode_AsUTF8String(obj)
	if bytes == nil {
		C.PyErr_Clear()
		return "", fmt.Errorf("paths must contain only strings")
	}
	defer C.Py_DecRef(bytes)

	var data *C.char
	var size C.Py_ssize_t
	if C.Pygfried_PyBytes_AsStringAndSize(bytes, &data, &size) != 0 {
		C.PyErr_Clear()
		return "", fmt.Errorf("paths must contain only strings")
	}

	ret := string(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size)))
	if strings.ContainsRune(ret, '\x00') {
		return "", fmt.Errorf("paths must not contain null bytes")
	}

	return ret, nil
}

func goStringListAndIntFromArgs(args *C.PyObject) ([]string, int, error) {
	var obj *C.PyObject
	var workers C.int
	if C.Pygfried_PyArg_ParseTuple_Oi(args, &obj, &workers) == 0 {
		return nil, 0, fmt.Errorf("Failed to parse arguments")
	}

	size := C.PyList_Size(obj)
	if size < 0 {
		C.PyErr_Clear()
		return nil, 0, fmt.Errorf("Failed to parse paths")
	}

	paths := make([]string, int(size))
	for idx := C.Py_ssize_t(0); idx < size; idx++ {
		item := C.PyList_GetItem(obj, idx)
		path, err := goPathStringFromPyUnicode(item)
		if err != nil {
			return nil, 0, err
		}
		paths[int(idx)] = path
	}

	return paths, int(workers), nil
}

func goDirArgs(args *C.PyObject) (string, bool, int, bool, error) {
	var obj *C.PyObject
	var recursive C.int
	var workers C.int
	var followSymlinks C.int
	if C.Pygfried_PyArg_ParseTuple_Uiii(args, &obj, &recursive, &workers, &followSymlinks) == 0 {
		return "", false, 0, false, fmt.Errorf("Failed to parse arguments")
	}

	bytes := C.PyUnicode_AsUTF8String(obj)
	ret := C.GoString(C.PyBytes_AsString(bytes))
	C.Py_DecRef(bytes)
	return ret, recursive != 0, int(workers), followSymlinks != 0, nil
}

func jsonStringToPyObject(jsonResult string) *C.PyObject {
	pyStr := stringToPy(jsonResult)
	if pyStr == nil {
		return raise(fmt.Errorf("Failed to convert string to Python object"))
	}

	result := C.Pygfried_json_loads(pyStr)
	C.Py_DecRef(pyStr)

	return result
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

//export identify_with_json
func identify_with_json(self *C.PyObject, args *C.PyObject) *C.PyObject {
	path, err := goStringFromArgs(args)
	if err != nil {
		return raise(err)
	}

	jsonResult, err := pygfried.IdentifyWithJSON(path)
	if err != nil {
		return raise(err)
	}

	return jsonStringToPyObject(jsonResult)
}

//export identify_many_with_json
func identify_many_with_json(self *C.PyObject, args *C.PyObject) *C.PyObject {
	paths, workers, err := goStringListAndIntFromArgs(args)
	if err != nil {
		return raise(err)
	}

	jsonResult, err := pygfried.IdentifyAllWithJSONOptions(paths, pygfried.IdentifyOptions{
		Workers: workers,
	})
	if err != nil {
		return raise(err)
	}

	return jsonStringToPyObject(jsonResult)
}

//export identify_dir_with_json
func identify_dir_with_json(self *C.PyObject, args *C.PyObject) *C.PyObject {
	path, recursive, workers, followSymlinks, err := goDirArgs(args)
	if err != nil {
		return raise(err)
	}

	jsonResult, err := pygfried.IdentifyDirWithJSON(path, pygfried.IdentifyDirOptions{
		Recursive:      recursive,
		Workers:        workers,
		FollowSymlinks: followSymlinks,
	})
	if err != nil {
		return raise(err)
	}

	return jsonStringToPyObject(jsonResult)
}

//export version
func version(self *C.PyObject) *C.PyObject {
	v := pygfried.Version()

	return stringToPyOrNone(v)
}

func main() {}
