#include <Python.h>

/* Will come from Go. */
PyObject* identify(PyObject*, PyObject*);
PyObject* identify_detailed(PyObject*, PyObject*);
PyObject* identify_many_detailed(PyObject*, PyObject*);
PyObject* identify_dir_detailed(PyObject*, PyObject*);
PyObject* version(PyObject*);

/* To shim Go's missing variadic function support. */
int Pygfried_PyArg_ParseTuple_U(PyObject* args, PyObject** s) {
    return PyArg_ParseTuple(args, "U", s);
}

int Pygfried_PyArg_ParseTuple_Oi(PyObject* args, PyObject** o, int* i) {
    return PyArg_ParseTuple(args, "Oi", o, i);
}

int Pygfried_PyArg_ParseTuple_Uiii(PyObject* args, PyObject** s, int* recursive, int* workers, int* follow_symlinks) {
    return PyArg_ParseTuple(args, "Uiii", s, recursive, workers, follow_symlinks);
}

int Pygfried_PyBytes_AsStringAndSize(PyObject* obj, char** buffer, Py_ssize_t* length) {
    return PyBytes_AsStringAndSize(obj, buffer, length);
}

/* Go cannot access C macros. */
PyObject* Pygfried_Py_RETURN_NONE() {
    Py_RETURN_NONE;
}

/* Exception types. */
PyObject* Pygfried_GoError;

static int validate_workers(int workers) {
    if (workers < 1 || workers > 1024) {
        PyErr_SetString(PyExc_ValueError, "workers must be between 1 and 1024");
        return 0;
    }
    return 1;
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
        result = identify_detailed(self, go_func_args);
    } else {
        result = identify(self, go_func_args);
    }

    Py_DECREF(go_func_args);
    return result;
}

static PyObject* pygfried_identify_many_wrapper(PyObject* self, PyObject* args, PyObject* kwargs) {
    PyObject* paths_obj = NULL;
    int workers = 1;
    static char* kwlist[] = {"paths", "workers", NULL};

    if (PyTuple_Size(args) > 1) {
        PyErr_SetString(PyExc_TypeError, "identify_many() takes 1 positional argument");
        return NULL;
    }
    if (!PyArg_ParseTupleAndKeywords(args, kwargs, "O|i", kwlist, &paths_obj, &workers)) {
        return NULL;
    }
    if (PyUnicode_Check(paths_obj)) {
        PyErr_SetString(PyExc_TypeError, "identify_many() paths must be an iterable of strings, not a string");
        return NULL;
    }
    if (!validate_workers(workers)) {
        return NULL;
    }

    PyObject* paths_list = PySequence_List(paths_obj);
    if (!paths_list) {
        return NULL;
    }

    PyObject* go_func_args = PyTuple_New(2);
    if (!go_func_args) {
        Py_DECREF(paths_list);
        return NULL;
    }
    PyObject* py_workers = PyLong_FromLong(workers);
    if (!py_workers) {
        Py_DECREF(paths_list);
        Py_DECREF(go_func_args);
        return NULL;
    }
    PyTuple_SetItem(go_func_args, 0, paths_list);
    PyTuple_SetItem(go_func_args, 1, py_workers);

    PyObject* result = identify_many_detailed(self, go_func_args);

    Py_DECREF(go_func_args);
    return result;
}

static PyObject* pygfried_identify_dir_wrapper(PyObject* self, PyObject* args, PyObject* kwargs) {
    char* path_str = NULL;
    PyObject* recursive_obj = Py_True;
    int workers = 1;
    PyObject* follow_symlinks_obj = Py_False;
    static char* kwlist[] = {"path", "recursive", "workers", "follow_symlinks", NULL};

    if (PyTuple_Size(args) > 1) {
        PyErr_SetString(PyExc_TypeError, "identify_dir() takes 1 positional argument");
        return NULL;
    }
    if (!PyArg_ParseTupleAndKeywords(args, kwargs, "s|OiO", kwlist, &path_str, &recursive_obj, &workers, &follow_symlinks_obj)) {
        return NULL;
    }
    if (!validate_workers(workers)) {
        return NULL;
    }

    int recursive = PyObject_IsTrue(recursive_obj);
    if (recursive == -1) {
        return NULL;
    }
    int follow_symlinks = PyObject_IsTrue(follow_symlinks_obj);
    if (follow_symlinks == -1) {
        return NULL;
    }

    PyObject* go_func_args = PyTuple_New(4);
    if (!go_func_args) {
        return NULL;
    }
    PyObject* py_path_str = PyUnicode_FromString(path_str);
    PyObject* py_recursive = PyLong_FromLong(recursive);
    PyObject* py_workers = PyLong_FromLong(workers);
    PyObject* py_follow_symlinks = PyLong_FromLong(follow_symlinks);
    if (!py_path_str || !py_recursive || !py_workers || !py_follow_symlinks) {
        Py_XDECREF(py_path_str);
        Py_XDECREF(py_recursive);
        Py_XDECREF(py_workers);
        Py_XDECREF(py_follow_symlinks);
        Py_DECREF(go_func_args);
        return NULL;
    }
    PyTuple_SetItem(go_func_args, 0, py_path_str);
    PyTuple_SetItem(go_func_args, 1, py_recursive);
    PyTuple_SetItem(go_func_args, 2, py_workers);
    PyTuple_SetItem(go_func_args, 3, py_follow_symlinks);

    PyObject* result = identify_dir_detailed(self, go_func_args);

    Py_DECREF(go_func_args);
    return result;
}

static const char identify_doc[] =
    "identify(path, detailed=False)\n"
    "--\n"
    "\n"
    "Identify one file. Return the first PRONOM identifier by default, or a\n"
    "detailed siegfried-style result dictionary when detailed is true.";

static const char identify_many_doc[] =
    "identify_many(paths, *, workers=1)\n"
    "--\n"
    "\n"
    "Identify an iterable of file paths and return a detailed siegfried-style\n"
    "result dictionary. The workers argument controls Go-side concurrency.";

static const char identify_dir_doc[] =
    "identify_dir(path, *, recursive=True, workers=1, follow_symlinks=False)\n"
    "--\n"
    "\n"
    "Identify regular files under a directory and return a detailed\n"
    "siegfried-style result dictionary. The workers argument controls Go-side concurrency.";

static const char version_doc[] =
    "version()\n"
    "--\n"
    "\n"
    "Return the embedded siegfried version.";

static struct PyMethodDef methods[] = {
    {
        "identify",
        (PyCFunction)pygfried_identify_wrapper,
        METH_VARARGS | METH_KEYWORDS,
        identify_doc,
    },
    {
        "identify_many",
        (PyCFunction)pygfried_identify_many_wrapper,
        METH_VARARGS | METH_KEYWORDS,
        identify_many_doc,
    },
    {
        "identify_dir",
        (PyCFunction)pygfried_identify_dir_wrapper,
        METH_VARARGS | METH_KEYWORDS,
        identify_dir_doc,
    },
    {"version", (PyCFunction)version, METH_NOARGS, version_doc},
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
