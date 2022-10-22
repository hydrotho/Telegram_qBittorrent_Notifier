import os
import shutil
import subprocess
import sys
from pathlib import Path

import click
from pyrogram import Client

APPLICATION_NAME = 'Telegram_qBittorrent_Notifier'
VERSION_NUMBER = '0.0.2'
SAVE_DIRECTORY = f"/etc/{APPLICATION_NAME}"
SESSION_FILE = f"{SAVE_DIRECTORY}/{APPLICATION_NAME}.session"


@click.group()
@click.version_option(prog_name=APPLICATION_NAME, version=VERSION_NUMBER)
def cli():
    pass


@cli.command()
@click.option(
    "--api-id",
    metavar="API_ID",
    prompt=True,
    help="Provide your telegram api_id."
)
@click.option(
    "--api-hash",
    metavar="API_HASH",
    prompt=True,
    hide_input=True,
    help="Provide your telegram api_hash."
)
def init(api_id, api_hash):
    if os.geteuid() != 0:
        subprocess.check_call(['sudo', sys.argv[0], 'init', '--api-id', api_id, '--api-hash', api_hash])
        sys.exit()

    Path(SAVE_DIRECTORY).mkdir(parents=True, exist_ok=True)
    Path(SESSION_FILE).unlink(missing_ok=True)

    app = Client(APPLICATION_NAME, api_id=api_id, api_hash=api_hash, app_version=f"{APPLICATION_NAME} {VERSION_NUMBER}",
                 workdir=SAVE_DIRECTORY, hide_password=True)
    app.run(send_message(app))


@cli.command()
@click.option("-n", "torrent_name", metavar="TORRENT_NAME")
@click.option("-l", "category", metavar="CATEGORY")
@click.option("-g", "tags", metavar="TAGS")
@click.option("-f", "content_path", metavar="CONTENT_PATH")
@click.option("-r", "root_path", metavar="ROOT_PATH")
@click.option("-d", "save_path", metavar="SAVE_PATH")
@click.option("-c", "number_of_files", metavar="NUMBER_OF_FILES")
@click.option("-z", "torrent_size", metavar="TORRENT_SIZE")
@click.option("-t", "current_tracker", metavar="CURRENT_TRACKER")
@click.option("-i", "info_hash", metavar="INFO_HASH")
def notify(torrent_name, category, tags, content_path, root_path, save_path, number_of_files, torrent_size,
           current_tracker, info_hash):
    message = f"[Message From **{APPLICATION_NAME}**]"
    message += "\n\n✅ Download Complete!\n\n"

    if torrent_name is not None:
        message += f"Torrent Name: {torrent_name}\n"
    if info_hash is not None:
        message += f"Info Hash: {info_hash}\n"
    if current_tracker is not None:
        message += f"Current Tracker: {current_tracker}\n"
    if number_of_files is not None:
        message += f"Number of Files: {number_of_files}\n"
    if torrent_size is not None:
        message += f"Torrent Size: {torrent_size}\n"

    if content_path is not None:
        message += f"Content Path: {content_path}\n"
    if root_path is not None:
        message += f"Root Path: {root_path}\n"
    if save_path is not None:
        message += f"Save Path: {save_path}\n"

    if category is not None:
        message += f"Category: #{category}\n"
    if tags is not None:
        tag_list = tags.split(',')
        message += f"Tags: #" + " #".join(tag_list)

    app = Client(APPLICATION_NAME, app_version=f"{APPLICATION_NAME} {VERSION_NUMBER}", no_updates=True)
    app.run(send_message(app, message))


@cli.command()
@click.argument("message", required=False)
def send(message):
    app = Client(APPLICATION_NAME, app_version=f"{APPLICATION_NAME} {VERSION_NUMBER}", no_updates=True)
    app.run(send_message(app, message))


async def send_message(app, message: str = None):
    async with app:
        if message is None:
            await app.send_message("self", f"Greetings from **{APPLICATION_NAME}**!")
        else:
            await app.send_message("self", message)


if __name__ == '__main__':
    if Path(SESSION_FILE).exists() is True:
        shutil.copy(SESSION_FILE, Path(sys.argv[0]).parent)

    cli()
