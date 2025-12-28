from typing import Optional, overload


def generate_all(items: list[str]) -> str:
    # TODO: sort by all lowercase? otherwise capital comes first?
    return Assign("__all__", List(sorted(items))).build()


def generate_import(from_: str, item: str) -> str:
    return Import(item, from_=f".{from_}").build()


class Buildable:

    def build(self) -> str:
        raise NotImplementedError


class Assign(Buildable):

    def __init__(self, name: str, value: Buildable) -> None:
        self._name = name
        self._value = value

    def build(self) -> str:
        return f"{self._name} = {self._value.build()}"


class List(Buildable):

    def __init__(self, values: list[str]) -> None:
        self._values = values

    def build(self) -> str:
        inner = "".join([f'"{i}", ' for i in self._values])
        return f"[{inner}]"


class Import(Buildable):

    @overload
    def __init__(self, item: str) -> None: ...

    @overload
    def __init__(self, item: str, *items: str, from_: str) -> None: ...

    def __init__(self, item: str, *items: str, from_: Optional[str] = None) -> None:
        self._item = item
        self._items = [item, *items]
        self._from = from_

    def build(self) -> str:
        if self._from is None:
            return f"import {self._item}"

        return f"from {self._from} import {','.join(self._items)}"
