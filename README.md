# MinAXNT

MinAXNT is a Go implementation of the Axentro mining protocol.

You can find more information about Axentro on the website: <https://axentro.io/>.

MinAXNT is well tested on:

* GNU/Linux
  * x86_64
  * ARM aarch64 (Raspberry Pi 3 Rev.3)
* MacOS (x86_64)
* Windows
  * x86_64

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

Sometime you have to use the same computer to mine $AXNT and do other stuff.

That's interesting to configure MinAXNT to use only a limited cores in order to use your computer in the same time.

To use MinAXNT on only 4 logical cores (the most powerful), you could use external tips described bellow:

#### Linux with `taskset`

On linux the most powerful cores in HT CPU are those who have an _even_ index (0, 2, 4, 6, 8, 10, ...).

Most distribution have `taskset` binary.

    taskset -c 0,2,4,6 ./minaxnt -a TTBjMjg1NWJmZTIwMTA3MDc3NTZkNjg1MTcxYjE1YmE5ZjU3N2EzODhkYWRiZWIy -n http://localhost:3000 -p4

This way only 4 cores are used for MinAXNT and the others for other stuff like programming, gaming or Internet, ...

#### Windows [WIP]

On Windows the most powerful cores in HT CPU are those who have an _odd_ index (1, 3, 5, 7, 9, 11, ...).

## Device Performance (Argon2id)

| **Device** | **Type** | **CPU Model** | **Operating System** | **MinAXNT process param** | **Work/s** | **MinAXNT version** |
|------------|----------|---------------|----------------------|---------------------------|------------|---------------------|
| **iMac Pro** | Desktop | Intel Xeon W-?? @3.00 GHz 10-core 64-bit | MacOs Catalina | 40 | **160 Work/s** | v0.7.0 |
| **MacBook Pro** | Laptop | Intel Core i9-?? U@2.3 GHz 8-Core 64-bit | MacOS Big sur | 30 | **108 Work/s** | v0.7.0 |
| MinisForum: DMAF5 | Mini-PC | AMD Ryzen 5 3550H with Radeon Vega Mobile Gfx @2.10 GHz 4-core HT (8 vcore) 64-bit | Windows 10 Family (x86_64) | 8 | **69 Work/s** | v0.9.0 |
| Lenovo: Y510P | Laptop | Intel(R) Core(TM) i7-4700MQ @2.40GHz Rev.3 4-core HT (8 vcore) 32-bit-64-bit | Ununtu 18.04 (x86_64) | 8 | **49 Work/s** (45 Work/s when using only the 4 physical CPU: `taskset -c 0,2,4,6 ./minaxnt -a xxx -p4`) | v0.9.0 |
| **IMB/PC** | Laptop | Intel i7-7820 HQ @2.90 GHz  4-core HT (8 vcore) 64-bit | Windows 10 Pro | ?? | **36 Work/s** | v0.7.0 |
| PINE64: ROCKPro64 2Gio RAM | SBC (ARM) | RK3399 Hexa-Core (2x ARM Cortex A72, 2.0GHz + 4x ARM Cortex A53, 1.5GHz)  | Ubuntu 18.04 (aarch64) | 6 | **10 Work/s** | v0.7.0 |
| Raspberry Pi 4, ?? Gio RAM | SBC (ARM) | 4x ARM Cortex-A72, 1.5GHz | ?? (aarch64) | ?? | **xx Work/s** | - |
| Raspberry Pi 3 Model B Rev 1.2, 1 Gio RAM | SBC (ARM) | 4x ARM Cortex-A53, 1.2GHz | Alpine Linux 5.4.84-0-rpi (aarch64) | 4 | **2.6 Work/s** | v0.9.0 |

## License

This project is under the MIT License. See the [LICENSE](https://github.com/fenicks/minaxnt/blob/main/LICENSE) file for the full license text.
