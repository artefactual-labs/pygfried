#include <Python.h>

/* Will come from Go */
PyObject* identify(PyObject*, PyObject*);
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

static struct PyMethodDef methods[] = {
    {"identify", (PyCFunction)identify, METH_VARARGS},
    {"version", (PyCFunction)version, METH_VARARGS},
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
