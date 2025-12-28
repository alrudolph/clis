from __future__ import annotations

from collections.abc import Sequence
from pathlib import Path
from typing import Optional

import pathspec
from rich import print

from .codegen import generate_all, generate_import
from .filters.file_filter import ConfigFileFilter
from .parsing import File, PythonObject


def load_gitignore(root: Path) -> pathspec.PathSpec:
    gitignore = root / ".gitignore"

    if not gitignore.exists():
        return pathspec.PathSpec.from_lines("gitwildmatch", [])

    with gitignore.open() as f:
        return pathspec.PathSpec.from_lines("gitwildmatch", f)


def check_ignored(specs: Sequence[pathspec.PathSpec], root: Path, path: Path) -> bool:
    rel_path = path.relative_to(root)

    for spec in reversed(specs):
        if spec.match_file(str(rel_path)):
            return True

    return False


class GenAll:

    def __init__(
        self,
        base_path: Path,
        pathspecs: Optional[list[pathspec.PathSpec]] = None,
    ) -> None:
        self._base_path = base_path
        self._all_objs: list[PythonObject] = []
        self._initialized = False

        if pathspecs is not None:
            self._pathspecs = [
                *pathspecs,
                load_gitignore(base_path),
            ]
        else:
            self._pathspecs = [
                pathspec.PathSpec(
                    [
                        pathspec.RegexPattern(".git"),
                        pathspec.RegexPattern("__init__.py"),
                    ],
                ),
                load_gitignore(base_path),
            ]

    def write_to_file(self) -> None:
        for dir in self._sub_dirs:
            dir.write_to_file()

        if not self._filter_file_path.exists():
            return

        items: list[tuple[str, str]] = []

        for obj in self.all_objs:
            rel = obj._file._path.relative_to(self._base_path)
            parts = rel.parts

            if len(parts) == 1:
                p = rel.stem
            else:
                p = parts[0]

            items.append((p, obj._name))

        imports = [generate_import(*item) for item in items]
        imp = "\n".join(imports)

        code = generate_all([i[1] for i in items])

        file_contents = f"{imp}\n\n{code}"

        with open(self._init_path, "w") as file:
            file.write(file_contents)

        rel = self._init_path.relative_to(self._base_path)
        print(f"[green]+[/green] [bold]Creating file[/bold] [purple]{rel}[/purple]")

    @property
    def all_objs(self) -> list[PythonObject]:
        if not self._initialized:
            self._generate()
            self._initialized = True

        return self._all_objs

    def _generate(self) -> None:
        all_items: list[PythonObject] = []
        output: list[PythonObject] = []

        for file in self._sub_files:
            all_items.extend(file.get_all_objs())

        for dir in self._sub_dirs:
            all_items.extend(dir.all_objs)

        for item in all_items:
            for filter in self._filters:
                if not filter.keep(item):
                    continue

                output.append(item)

        self._all_objs = output

    @property
    def _sub_dirs(self) -> list[GenAll]:
        # TODO: only containing python files
        return [
            GenAll(p, pathspecs=self._pathspecs)
            for p in self._base_path.iterdir()
            if not p.is_file()
            and not check_ignored(self._pathspecs, self._base_path, p)
        ]

    @property
    def _sub_files(self) -> list[File]:
        return [
            File(p)
            for p in self._base_path.iterdir()
            if p.is_file()
            and p.suffix == ".py"
            and not check_ignored(self._pathspecs, self._base_path, p)
        ]

    @property
    def _init_path(self) -> Path:
        return self._base_path / "__init__.py"

    @property
    def _filters(self) -> list[ConfigFileFilter]:
        if not self._filter_file_path.exists():
            return []

        return [ConfigFileFilter.from_file(self._filter_file_path)]

    @property
    def _filter_file_path(self) -> Path:
        return self._base_path / ".genall.yaml"
