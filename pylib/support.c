#include <Python.h>

/* Will come from Go. */
PyObject* identify(PyObject*, PyObject*);
PyObject* identify_with_json(PyObject*, PyObject*);
PyObject* version(PyObject*);

/* To shim Go's missing variadic function support. */
int Pygfried_PyArg_ParseTuple_U(PyObject* args, PyObject** s) {
    return PyArg_ParseTuple(args, "U", s);
}

/* Go cannot access C macros. */
PyObject* Pygfried_Py_RETURN_NONE() {
    Py_RETURN_NONE;
}

/* Exception types. */
PyObject* Pygfried_GoError;

/* JSON loading function used by Go. */
PyObject* Pygfried_json_loads(PyObject* json_str) {
    PyObject* json_module = PyImport_ImportModule("json");
    if (!json_module) {
        return NULL;
    }

    PyObject* loads_func = PyObject_GetAttrString(json_module, "loads");
    Py_DECREF(json_module);
    if (!loads_func) {
        return NULL;
    }

    PyObject* args = PyTuple_Pack(1, json_str);
    PyObject* result = PyObject_CallObject(loads_func, args);
    Py_DECREF(loads_func);
    Py_DECREF(args);

    return result;
}

static PyObject* pygfried_identify_wrapper(PyObject* self, PyObject* args, PyObject* kwargs) {
    char* path_str = NULL;
    PyObject* detailed_kwarg_obj = Py_False;
    static char* kwlist[] = {"path", "detailed", NULL};

    if (!PyArg_ParseTupleAndKeywords(args, kwargs, "s|O", kwlist, &path_str, &detailed_kwarg_obj)) {
        return NULL;
    }

    int use_detailed = PyObject_IsTrue(detailed_kwarg_obj);
    if (use_detailed == -1) {
        return NULL;
    }

    PyObject* go_func_args = PyTuple_New(1);
    if (!go_func_args) {
        return NULL;
    }
    PyObject* py_path_str = PyUnicode_FromString(path_str);
    if (!py_path_str) {
        Py_DECREF(go_func_args);
        return NULL;
    }
    PyTuple_SetItem(go_func_args, 0, py_path_str);

    PyObject* result;
    if (use_detailed) {
        result = identify_with_json(self, go_func_args);
    } else {
        result = identify(self, go_func_args);
    }

    Py_DECREF(go_func_args);
    return result;
}

static struct PyMethodDef methods[] = {
    {"identify", (PyCFunction)pygfried_identify_wrapper, METH_VARARGS | METH_KEYWORDS},
    {"version", (PyCFunction)version, METH_NOARGS},
    {NULL, NULL}
};

static PyObject* _setup_module(PyObject* module) {
    if (module) {
        Pygfried_GoError = PyErr_NewException("pygfried.GoError", PyExc_OSError, NULL);
        PyModule_AddObject(module, "GoError", Pygfried_GoError);
    }
    return module;
}

static struct PyModuleDef module = {
    PyModuleDef_HEAD_INIT,
    "pygfried",
    NULL,
    -1,
    methods
};

PyMODINIT_FUNC PyInit_pygfried(void) {
    return _setup_module(PyModule_Create(&module));
}
