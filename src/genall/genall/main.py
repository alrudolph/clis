from pathlib import Path

import click

from .genall import GenAll


@click.command()
@click.option(
    "--path",
    prompt="Path to the base directory",
    help="The base directory path.",
)
def main(path: str) -> None:
    base_path = Path(path).expanduser().resolve()
    #
    # TODO: path could just be file
    #
    ga = GenAll(base_path)
    ga.write_to_file()
