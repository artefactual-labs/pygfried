clean:
	rm -fr build/
	rm -fr dist/
	rm -fr .eggs/
	find . -name '*.egg-info' -exec rm -rf {} +
	find . -name '*.egg' -exec rm -rf {} +
	find . -name '*.pyc' -exec rm -f {} +
	find . -name '*.pyo' -exec rm -f {} +
	find . -name '*~' -exec rm -f {} +
	find . -name '__pycache__' -exec rm -fr {} +
	rm -fr .pytest_cache

release-linux: clean
	setuptools-golang-build-manylinux-wheels --golang="1.20.2" --pythons="cp37-cp37m cp38-cp38 cp39-cp39"
	python setup.py sdist
	twine upload dist/*
