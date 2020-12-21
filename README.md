# MinAXNT

MinAXNT is a Go implementation of the Axentro mining protocol.

You can find more information about Axentro on the website: <https://axentro.io/>.

MinAXNT is well tested on GNU/Linux (x86_64), MacOs (x86_64) and Windows (x86_64).

## Install

Installing MinAXNT is really simple:

First: _You need an Axentro wallet address to use the miner, please visite the [website](https://axentro.io/) or the [Youtube channel](https://www.youtube.com/channel/UC8itve8eHunJkfuyJYwMy6g)_

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

## Device Performance (Argon2id)

| **Device** | **Type** | **CPU Model** | **Operating System** | **MinAXNT process param** | **Work/s** |
|------------|----------|---------------|----------------------|---------------------------|------------|
| Lenovo: Y510P | Laptop | Intel(R) Core(TM) i7-4700MQ CPU @ 2.40GHz Rev. 3 (32-bit, 64-bit) | Ununtu 18.04 | 8 | **48 Work/s** (45 Work/s when using only the 4 physical CPU: `taskset -c 0,2,4,6 ./minaxnt -a xxx -p8`) |
| MinisForum: DMAF5 | Mini-PC | AMD Ryzen 5 3550H with Radeon Vega Mobile Gfx 2.10 GHz (64-bit) | Windows 10 Family | 8 | **68 Work/s** |
| PINE64: ROCKPro64 2Gio RAM | SBC (ARM) | RK3399 SOC with Mali T860 MP4 GPU (aarch64) | Ubuntu 18.04 | 6 | **10 Work/s** |

## License

This project is under the MIT License. See the [LICENSE](https://github.com/fenicks/minaxnt/blob/main/LICENSE) file for the full license text.
