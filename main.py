import click
from configparser import ConfigParser
from pathlib import Path
from pyrogram import Client
from pyrogram.enums import ParseMode

APPLICATION_NAME = 'Telegram_qBittorrent_Notifier'
VERSION_NUMBER = '0.0.1'
WORKING_DIRECTORY = None
CONFIG_PATH = f"{str(Path.home())}/.config/{APPLICATION_NAME}.ini"


@click.group()
@click.version_option(version=VERSION_NUMBER, prog_name=APPLICATION_NAME)
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
@click.option(
    "-d",
    "--working-directory",
    metavar="WORKING_DIRECTORY",
    prompt=True,
    default=f"{Path.home() / '.cache'}",
    help="Specify the working directory, where the Telegram session file should be saved."
)
def init(api_id, api_hash, working_directory):
    directory = Path(working_directory)
    session_file = directory / (APPLICATION_NAME + '.session')
    directory.mkdir(parents=True, exist_ok=True)
    session_file.unlink(missing_ok=True)

    config['DEFAULT']['working_directory'] = working_directory
    with open(CONFIG_PATH, 'w') as config_file:
        config.write(config_file)

    app = Client(APPLICATION_NAME, api_id=api_id, api_hash=api_hash,
                 app_version=APPLICATION_NAME + ' ' + VERSION_NUMBER,
                 workdir=working_directory, parse_mode=ParseMode.MARKDOWN, hide_password=True)
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

    app = Client(APPLICATION_NAME, workdir=WORKING_DIRECTORY, parse_mode=ParseMode.MARKDOWN)
    app.run(send_message(app, message))


@cli.command()
@click.argument("message", required=False)
def send(message):
    app = Client(APPLICATION_NAME, workdir=WORKING_DIRECTORY, parse_mode=ParseMode.MARKDOWN)
    app.run(send_message(app, message))


async def send_message(app, message: str = None):
    async with app:
        if message is None:
            await app.send_message("self", f"Greetings from **{APPLICATION_NAME}**!")
        else:
            await app.send_message("self", message)


if __name__ == '__main__':
    config = ConfigParser()
    if Path(CONFIG_PATH).exists():
        config.read(CONFIG_PATH)
        WORKING_DIRECTORY = config['DEFAULT']['working_directory']

    cli()
