# slow-stream
:rabbit2: pipe to throttle streams (bin + go lib)

![Slow Stream Logo](https://raw.githubusercontent.com/moul/slow-stream/master/contrib/assets/slow-stream.png)

## Usage

```console
$ slow-stream -h
NAME:
   slow-stream - Slow Stream

USAGE:
   slow-stream [global options] command [command options] [arguments...]

VERSION:
   1.2.0

AUTHOR(S):
   Manfred Touron <https://github.com/moul/slow-stream>

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose, -V			Enable verbose mode
   --raw, -r				Enable raw mode
   --stdout-passthrough			Do not slow stdout
   --max-write-interval, -i "100"	Max write interval (in millisecond)
   --buff-size, -b "1024"		Buffer size
   --help, -h				show help
   --version, -v			print the version
```

## Schema

```
          ┌───────┐
          │ Hello │
          └───────┘
              │
            stdin
              │
              ▼
┌───────────────────────────┐
│ $ slow-stream -b1 -i 100  │
└───────────────────────────┘
              │
            stdout
              │            ┌───┐
              ├────t=0────▶│ H │
              │            └───┘
              │            ┌───┐
              ├──t=100ms──▶│ e │
              │            └───┘
              │            ┌───┐
              ├──t=200ms──▶│ l │
              │            └───┘
              │            ┌───┐
              ├──t=300ms──▶│ l │
              │            └───┘
              │            ┌───┐
              └──t=400ms──▶│ o │
                           └───┘
```

## License

MIT
