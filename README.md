# MindMiner [![Version tag](https://img.shields.io/github/release/Quake4/MindMiner.svg)](https://github.com/Quake4/MindMiner/releases/latest) [![Version date tag](https://img.shields.io/github/release-date/Quake4/MindMiner.svg)](https://github.com/Quake4/MindMiner/releases/latest) [![Issues tag](https://img.shields.io/github/issues-raw/Quake4/MindMiner.svg)](https://github.com/Quake4/MindMiner/issues)
### *by Oleg Samsonov aka Quake4*

Multi-algorithm, multi-platform, miner manager program.
Mining supported on NVIDIA, AMD, and CPU platforms.

This is not another fork based on MultiPoolMiner (NemosMiner, SniffDogMiner, MegaMiner, NPlusMiner and etc).
This is a fully new, from scratch source code, written by myself with a smarter miner manager program as the goal.

If anyone has a claim to any of it post your case in the [Discord](https://discord.gg/HrKChEQ) or Bitcoin Talk Forum on [english](https://bitcointalk.org/index.php?topic=3022754) or [russian](https://bitcointalk.org/index.php?topic=3139620) or [create issue](https://github.com/Quake4/MindMiner/issues/new).

You may configure and further fine-tune any supported miner as well, by modifying its accompanying config.txt

## Features
* Online monitoring on [MindMiner site](http://mindminer.online/monitoring)
* Accounting power consumption
* Very small [![code size](https://img.shields.io/github/languages/code-size/Quake4/MindMiner.svg)](https://github.com/Quake4/MindMiner)
* Low memory use ~60Mb
* Self update
* User confirm for benchmarks and new pools
* No memory leak
* Small CPU usage
* Fast switch for most profit algo/coin
* Very configurable
* Different verbose level
* Pools actual balance
* Actual, up-to-date miners (if not, write me)
* Up to three currencies ([supported list](https://api.coinbase.com/v2/exchange-rates?currency=BTC))
* Api Pools proxy for more then 7 rigs (prevent blocking api pools request)
* Api/status server (http://127.0.0.1:5555)
* Switching resistance by percentage and/or timeout
* Dual mining on Claymore Ethereum & Bminer
* Run process before & after execution of miner
* Account accepted & rejected shares
* Failover pools support
* Rentals pool support (MiningRigRentals)

## Support
### Pools ([full list](https://github.com/Quake4/MindMiner/tree/master/Pools))
* AhashPool
* BlazePool
* BlockMasters
* MiningPoolHub
* MiningRigRentals
* NiceHash
* NLPool
* ZergPool
* Zpool

### Miners ([full list](https://github.com/Quake4/MindMiner/tree/master/Miners))
* ccminer (any)
* сlaymore ethereum (dual)
* сlaymore neoscrypt
* cpuminer (any)
* cpuminer-opt/bf
* cryptodredge
* ethminer
* gminer
* lolminer
* phoenix (ethash)
* sgminer (any)
* srbm-multi
* t-rex
* teamred
* wildrig
* xmrig all
* xmr-stak
* z-enemy
* archive (bminer, cast xmr, claymore cn/zcash, dstm, ewbf, gatelessgate, hsrminer, jce, nheqminer & other)

## Requirements

**PowerShell 5.1**
* Windows 10 x64 or Server 2016 already contain PowerShell 5.1
* For Windows 7 SP1, 8.1, Server 2008 R2 SP1, 2012, 2012 R2 install [PowerShell 5.1 by link](https://www.microsoft.com/download/details.aspx?id=54616).
* PowerShell requires [Microsoft .NET Framework 4.5.2](https://docs.microsoft.com/powershell/wmf/5.1/install-configure) or [above](https://msdn.microsoft.com/library/5a4x27ek(v=vs.110).aspx).

Windows 64-bit edition required as most of the miners releases are compiled as x64 and support only x64 platforms.

Please install both x64 and x86 versions:
* Visual C++ Redistributable 2015, 2017, 2019 [x64](https://aka.ms/vs/16/release/vc_redist.x64.exe) [x86](https://aka.ms/vs/16/release/vc_redist.x86.exe)
* [Visual C++ Redistributable 2013](https://www.microsoft.com/download/details.aspx?id=40784)
* ~~[Visual C++ Redistributable 2012](https://www.microsoft.com/download/details.aspx?id=30679)~~

If use CPU mining please [allow lock page support](https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/enable-the-lock-pages-in-memory-option-windows) in OS to double CryptoNight algorithm profit (XMR).

## Install
Download [![latest release](https://img.shields.io/github/release/Quake4/MindMiner.svg)](https://github.com/Quake4/MindMiner/releases/latest) to any folder. Unpack and may create `config.txt` (see config section) or just run "run.bat" and enter BTC wallet and other asked data.

## Config ([full manual](https://github.com/Quake4/MindMiner/blob/master/Config.md))
Place simple `config.txt` file into programm folder with json content
```json
{
    "Wallet": { "BTC":  "Your BTC wallet" },
    "Login": "Login for pool with registration (MPH)"
}
```
For details, see [full configuration manual](https://github.com/Quake4/MindMiner/blob/master/Config.md).

## Screenshot
### Monitoring
![MindMiner](https://github.com/Quake4/MindMinerPrerequisites/raw/master/monitoring.png "MindMiner online monitoring")
### nVidia GTX 1070
![MindMiner](https://github.com/Quake4/MindMinerPrerequisites/raw/master/MindMiner.png "MindMiner on nVidia GTX 1070")
### nVidia GTX 1060 3G
![MindMiner](https://github.com/Quake4/MindMinerPrerequisites/raw/master/GTX10603G.png "MindMiner on nVidia GTX 1060 3G")
### AMD RX 580 4G
![MindMiner](https://github.com/Quake4/MindMinerPrerequisites/raw/master/RX5804G.png "MindMiner on AMD RX 580 4G")
### Intel i3-6100
![MindMiner](https://github.com/Quake4/MindMinerPrerequisites/raw/master/i36100.png "MindMiner on Intel i3-6100")

## Fee
MindMiner has development fee 1%. Enabling MinigRigRentals will add 1% fee.

## Thanks
Thanks to aaronsace to the idea but weak implementation.
