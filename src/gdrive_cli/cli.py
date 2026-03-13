import click

from gdrive_cli.commands import auth, fetch, search


@click.group()
@click.option("--debug", is_flag=True, default=False, help="Enable verbose debug logging.")
@click.pass_context
def cli(ctx: click.Context, debug: bool):
    """Google Drive CLI — fetch and search docs, sheets, and slides."""
    ctx.ensure_object(dict)
    ctx.obj["debug"] = debug


cli.add_command(auth)
cli.add_command(fetch)
cli.add_command(search)


def main():
    cli()
