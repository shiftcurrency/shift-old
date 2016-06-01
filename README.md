## SHIFT Blocktime

Fixed at 50 seconds (average target 35-40 seconds)

## SHIFT Blockreward

Static reward of 2 Shift per block.


## Building SHIFT

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/shiftcurrency/shift/wiki/Building-and-Installing-Shift)
on the wiki.

Building gshift requires both a Go and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make gshift

## Executables

Go Shift comes with several wrappers/executables found in
[the `cmd` directory](https://github.com/shiftcurrency/shift/tree/develop/cmd):

* `gshift` Shift CLI (shift command line interface client)
* `bootnode` runs a bootstrap node for the Discovery Protocol
* `shifttest` test tool which runs with the [tests](https://github.com/shiftcurrency/tests) suite:
  `/path/to/test.json > shftest --test BlockTests --stdin`.
* `evm` is a generic Shift Virtual Machine: `evm -code 60ff60ff -gas
  10000 -price 0 -dump`. See `-h` for a detailed description.
* `disasm` disassembles EVM code: `echo "6001" | disasm`
* `rlpdump` prints RLP structures

## Command line options

`gshift` can be configured via command line options, environment variables and config files.

To get the options available:

    gshift help

## WIKI (to be added on readthedocs)

## Contribution

If you'd like to contribute to shift please fork, fix, commit and
send a pull request. Commits who do not comply with the coding standards
are ignored (use gofmt!). If you send pull requests make absolute sure that you
commit on the `develop` branch and that you do not merge to master.
Commits that are directly based on master are simply ignored.

See [Developers' Guide](https://github.com/shiftcurrency/shift/wiki/Developers'-Guide)
for more details on configuring your environment, testing, and
dependency management.

