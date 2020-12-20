# MinAXNT

MinAXNT is a Go implementation of the Axentro mining protocol.

You can find more information about Axentro on the website: <https://axentro.io/>.

MinAXNT is well tested on GNU/Linux (x86_64), MacOs (x86_64) and Windows (x86_64).

## Install

Installing MinAXNT is really simple:

* Download your platform archive: <https://github.com/fenicks/minaxnt/releases/>
* Uncompress the archive in a directory of your choice
* Open a terminal and go the MinAXNT directory
* Run the program as described in the section bellow

## Usage

    ./minaxnt -n http://mainnet.axentro.io -a TTAyNmVjNGU4MTBlYzc1ZWVmNDMyYzc4MjI5NGRmZWNhYzUwMzJjN2UyYzlhNzE3 -p 2

OR

    ./minaxnt --node http://mainnet.axentro.io --address TTAyNmVjNGU4MTBlYzc1ZWVmNDMyYzc4MjI5NGRmZWNhYzUwMzJjN2UyYzlhNzE3 --process 2

OR

    minaxnt.exe --node http://mainnet.axentro.io --address TTAyNmVjNGU4MTBlYzc1ZWVmNDMyYzc4MjI5NGRmZWNhYzUwMzJjN2UyYzlhNzE3 --process 2

Have fun !

## Tunning

### CPU Affinity : Hyper-Threaded Architecture

[WIP]

## License

This project is under the MIT License. See the [LICENSE](https://github.com/fenicks/minaxnt/blob/main/LICENSE) file for the full license text.
