from __future__ import annotations

import re
from pathlib import Path
from typing import Optional

from yaml import safe_load

from genall.parsing import PythonObject


class ConfigFileFilter:

    def __init__(
        self,
        classes: Optional[list[Filter]] = None,
        functions: Optional[list[Filter]] = None,
        variables: Optional[list[Filter]] = None,
    ) -> None:
        self._classes = classes
        self._functions = functions
        self._variables = variables

    @classmethod
    def from_file(cls, path: Path) -> ConfigFileFilter:
        with open(path, "r") as f:
            data = safe_load(f)

        if data is None:
            data = {}

        classes: list[Filter] | None = None
        functions: list[Filter] | None = None
        variables: list[Filter] | None = None

        if data.get("classes") is not None:
            classes = []
            for item in data["classes"]:
                filter = Filter(
                    type="class",
                    name=item["name"],
                    include=item.get("include", True),
                )
                classes.append(filter)

        if data.get("functions") is not None:
            functions = []
            for item in data["functions"]:
                filter = Filter(
                    type="function",
                    name=item["name"],
                    include=item.get("include", True),
                )
                functions.append(filter)

        if data.get("variables") is not None:
            variables = []
            for item in data["variables"]:
                filter = Filter(
                    type="variable",
                    name=item["name"],
                    include=item.get("include", True),
                )
                variables.append(filter)

        return cls(
            classes=classes,
            functions=functions,
            variables=variables,
        )

    def keep(self, obj: PythonObject) -> bool:
        return self._keep_by_type(obj)

    def _keep_obj_filters(
        self,
        obj: PythonObject,
        filters: Optional[list[Filter]] = None,
    ) -> bool:
        if filters is None:
            return True

        for filter in filters:
            if not filter.matches(obj):
                if not filter._include:
                    return True

                continue

            return filter.keep(obj)

        return False

    def _keep_class(self, obj: PythonObject) -> bool:
        return self._keep_obj_filters(obj, self._classes)

    def _keep_function(self, obj: PythonObject) -> bool:
        return self._keep_obj_filters(obj, self._functions)

    def _keep_variable(self, obj: PythonObject) -> bool:
        return self._keep_obj_filters(obj, self._variables)

    def _keep_by_type(self, obj: PythonObject) -> bool:
        if obj._type == "class":
            return self._keep_class(obj)
        elif obj._type == "function":
            return self._keep_function(obj)
        elif obj._type == "variable":
            return self._keep_variable(obj)

        return True


class Filter:

    def __init__(self, type: str, name: str, include: bool) -> None:
        self._type = type
        self._name = name
        self._include = include

    def matches(self, obj: PythonObject) -> bool:
        if obj._type != self._type:
            return False

        if not self._name.startswith("/") or not self._name.endswith("/"):
            return obj._name == self._name

        pattern = self._name[1:-1]
        return re.match(pattern, obj._name) is not None

    def keep(self, obj: PythonObject) -> bool:
        if not self.matches(obj):
            return False

        return self._include

    def __repr__(self) -> str:
        return f"Filter(type={self._type!r}, name={self._name!r}, include={self._include!r})"
