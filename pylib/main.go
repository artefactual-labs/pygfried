package main

// #include <stdlib.h>
// #include <Python.h>
// extern int Pygfried_PyArg_ParseTuple_U(PyObject*, PyObject**);
// extern int Pygfried_PyArg_ParseTuple_Oi(PyObject*, PyObject**, int*);
// extern int Pygfried_PyArg_ParseTuple_Uiii(PyObject*, PyObject**, int*, int*, int*);
// extern int Pygfried_PyBytes_AsStringAndSize(PyObject*, char**, Py_ssize_t*);
// extern PyObject* Pygfried_Py_RETURN_NONE();
// extern PyObject* Pygfried_GoError;
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
	pystr := C.PyUnicode_FromStringAndSize(cstr, C.Py_ssize_t(len(s)))
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

//export identify_detailed
func identify_detailed(self *C.PyObject, args *C.PyObject) *C.PyObject {
	path, err := goStringFromArgs(args)
	if err != nil {
		return raise(err)
	}

	detailedResult, err := pygfried.IdentifyWithDetailedResult(path)
	if err != nil {
		return raise(err)
	}

	return detailedResultToPyObject(detailedResult)
}

//export identify_many_detailed
func identify_many_detailed(self *C.PyObject, args *C.PyObject) *C.PyObject {
	paths, workers, err := goStringListAndIntFromArgs(args)
	if err != nil {
		return raise(err)
	}

	detailedResult, err := pygfried.IdentifyAllWithDetailedOptions(paths, pygfried.IdentifyOptions{
		Workers: workers,
	})
	if err != nil {
		return raise(err)
	}

	return detailedResultToPyObject(detailedResult)
}

//export identify_dir_detailed
func identify_dir_detailed(self *C.PyObject, args *C.PyObject) *C.PyObject {
	path, recursive, workers, followSymlinks, err := goDirArgs(args)
	if err != nil {
		return raise(err)
	}

	detailedResult, err := pygfried.IdentifyDirWithDetailedResult(path, pygfried.IdentifyDirOptions{
		Recursive:      recursive,
		Workers:        workers,
		FollowSymlinks: followSymlinks,
	})
	if err != nil {
		return raise(err)
	}

	return detailedResultToPyObject(detailedResult)
}

//export version
func version(self *C.PyObject) *C.PyObject {
	v := pygfried.Version()

	return stringToPyOrNone(v)
}

func detailedResultToPyObject(result *pygfried.DetailedResult) *C.PyObject {
	obj := C.PyDict_New()
	if obj == nil {
		return nil
	}

	if !pyDictSetString(obj, "siegfried", result.Siegfried) ||
		!pyDictSetString(obj, "scandate", result.ScanDate) ||
		!pyDictSetString(obj, "signature", result.Signature) ||
		!pyDictSetString(obj, "created", result.Created) ||
		!pyDictSetItem(obj, "identifiers", identifiersToPyList(result.Identifiers)) ||
		!pyDictSetItem(obj, "files", filesToPyList(result.Files)) {
		C.Py_DecRef(obj)
		return nil
	}

	return obj
}

func identifiersToPyList(identifiers []pygfried.DetailedIdentifier) *C.PyObject {
	list := C.PyList_New(C.Py_ssize_t(len(identifiers)))
	if list == nil {
		return nil
	}

	for idx, identifier := range identifiers {
		obj := C.PyDict_New()
		if obj == nil {
			C.Py_DecRef(list)
			return nil
		}
		if !pyDictSetString(obj, "name", identifier.Name) ||
			!pyDictSetString(obj, "details", identifier.Details) {
			C.Py_DecRef(obj)
			C.Py_DecRef(list)
			return nil
		}
		if !pyListSetItem(list, idx, obj) {
			C.Py_DecRef(list)
			return nil
		}
	}

	return list
}

func filesToPyList(files []pygfried.DetailedFile) *C.PyObject {
	list := C.PyList_New(C.Py_ssize_t(len(files)))
	if list == nil {
		return nil
	}

	for idx, file := range files {
		obj := detailedFileToPyObject(file)
		if !pyListSetItem(list, idx, obj) {
			C.Py_DecRef(list)
			return nil
		}
	}

	return list
}

func detailedFileToPyObject(file pygfried.DetailedFile) *C.PyObject {
	obj := C.PyDict_New()
	if obj == nil {
		return nil
	}

	if !pyDictSetString(obj, "filename", file.Filename) ||
		!pyDictSetItem(obj, "filesize", C.PyLong_FromLongLong(C.longlong(file.FileSize))) ||
		!pyDictSetString(obj, "modified", file.Modified) ||
		!pyDictSetString(obj, "errors", file.Errors) ||
		!pyDictSetItem(obj, "matches", matchesToPyList(file.Matches)) {
		C.Py_DecRef(obj)
		return nil
	}

	return obj
}

func matchesToPyList(matches []pygfried.DetailedMatch) *C.PyObject {
	list := C.PyList_New(C.Py_ssize_t(len(matches)))
	if list == nil {
		return nil
	}

	for idx, match := range matches {
		obj := detailedMatchToPyObject(match)
		if !pyListSetItem(list, idx, obj) {
			C.Py_DecRef(list)
			return nil
		}
	}

	return list
}

func detailedMatchToPyObject(match pygfried.DetailedMatch) *C.PyObject {
	obj := C.PyDict_New()
	if obj == nil {
		return nil
	}

	for _, field := range match.Fields {
		if !pyDictSetString(obj, field.Name, field.Value) {
			C.Py_DecRef(obj)
			return nil
		}
	}

	return obj
}

func pyDictSetString(dict *C.PyObject, key string, value string) bool {
	return pyDictSetItem(dict, key, stringToPy(value))
}

func pyDictSetItem(dict *C.PyObject, key string, value *C.PyObject) bool {
	if value == nil {
		return false
	}

	cKey := C.CString(key)
	result := C.PyDict_SetItemString(dict, cKey, value)
	C.free(unsafe.Pointer(cKey))
	C.Py_DecRef(value)

	return result == 0
}

func pyListSetItem(list *C.PyObject, idx int, value *C.PyObject) bool {
	if value == nil {
		return false
	}
	if C.PyList_SetItem(list, C.Py_ssize_t(idx), value) != 0 {
		C.Py_DecRef(value)
		return false
	}
	return true
}

func main() {}
