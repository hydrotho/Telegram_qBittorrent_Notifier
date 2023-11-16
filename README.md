# Telegram_qBittorrent_Notifier

A simple CLI tool for qBittorrent that sends a notification to Telegram chat via bot on torrent finished.

## Quick Start

### Installation

Download prebuilt statically-linked binaries from [Releases](https://github.com/hydrotho/Telegram_qBittorrent_Notifier/releases/latest) for Linux, Windows and macOS.

### Configuration

Please refer to the [config example](config.example.yaml).

### Integrate with qBittorrent

In the qBittorrent Web UI:

1. Navigate to `Options` and select the `Downloads` tab.
2. At the bottom, locate `Run external program on torrent finished`.
   ![Run External Program](https://github.com/hydrotho/Telegram_qBittorrent_Notifier/assets/42911474/bb262e35-422f-4522-8530-4ad583d12798)

### Handling Empty Category or Tags

When using `%L` (Category) or `%G` (Tags), follow these formats to avoid errors:

```sh
❯ Telegram_qBittorrent_Notifier send -n "%N" -l "6д9%L"
❯ Telegram_qBittorrent_Notifier send -n "%N" -g "6д9%G"
❯ Telegram_qBittorrent_Notifier send -n "%N" -l "6д9%L" -g "6д9%G"
```

These formats use "6д9" as a prefix to ensure that the CLI does not encounter parsing errors from empty values in `%L` or `%G`.

You can customize this magic word in the configuration file or by using the `--magic-word` option.

## Support

If you encounter any issues or have any suggestions, please [open an issue](https://github.com/hydrotho/Telegram_qBittorrent_Notifier/issues).

## License

This project is licensed under the MIT License, see the [LICENSE](LICENSE) file for details.
