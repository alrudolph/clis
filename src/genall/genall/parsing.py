from __future__ import annotations

import ast
from pathlib import Path


class PythonObject:

    def __init__(self, file: File, type: str, name: str) -> None:
        self._file = file
        self._type = type
        self._name = name

    def __repr__(self) -> str:
        return f"PythonObject({self._type!r}, {self._name!r})"


class File:

    def __init__(self, path: Path) -> None:
        self._path = path

    @property
    def _source(self) -> str:
        with open(self._path, "r") as file:
            return file.read()

    @property
    def _ast(self) -> ast.Module:
        return ast.parse(self._source, self._path.name)

    def get_all_objs(self) -> list[PythonObject]:
        objs: list[PythonObject] = []

        for node in self._ast.body:
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
                name = node.name
                obj = PythonObject(self, "function", name)
            elif isinstance(node, ast.ClassDef):
                name = node.name
                obj = PythonObject(self, "class", name)
            elif isinstance(node, ast.Assign):
                # TODO: multiple assignments...
                name = node.targets[0].id  # type: ignore
                obj = PythonObject(self, "variable", name)
            else:
                continue

            if name.startswith("_"):
                continue

            objs.append(obj)

        return sorted(objs, key=lambda x: x._name)
