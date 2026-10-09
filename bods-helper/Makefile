PYTHON ?= python3

.PHONY: test build-deps build clean

test:
	$(PYTHON) -m unittest discover -s tests -v

build-deps:
	$(PYTHON) -m pip install -r requirements-build.txt

build: build-deps
	$(PYTHON) -m PyInstaller --clean --onefile --name bods-helper --distpath dist --workpath build --specpath build app.py

clean:
	rm -rf build dist