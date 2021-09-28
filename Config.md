# Сonfiguration manual of MindMiner
Any configuration stored in json format.

## MindMiner config
MindMiner settings placed in config.txt file into root application folder.

Main settings file is read only at the start of the MindMiner. If configuration file is absent or has wrong json format MindMiner ask your wallet and create default config.

```json
{
    "Region": "Europe",
    "SSL": true,
    "Wallet": { "BTC": "BTC Wallet", "LTC": "LTC Wallet", "NiceHashNew": "NiceHash New Wallet", "NiceHash": "NiceHash Old Wallet" },
    "WorkerName": "Worker name",
    "Login": "Login",
    "Password": "x",
    "CheckTimeout": 5,
    "LoopTimeout": 60,
    "NoHashTimeout": 10,
    "AverageCurrentHashSpeed": 180,
    "AverageHashSpeed": "12 hours",
    "Verbose": "Normal",
    "ShowBalance": true,
    "ShowExchangeRate": true,
    "AllowedTypes": [ "CPU", "nVidia", "AMD", "Intel" ],
    "Currencies": { "BTC": 8, "USD": 2, "EUR": 2 },
    "CoolDown": 0,
    "ApiServer": false,
    "ApiServerAllowWallets": false,
    "SwitchingResistance": { "Enabled": true, "Percent": 4, "Timeout": 15 },
    "BenchmarkSeconds": { "CPU": 60, "nVidia": 240 },
    "MinimumMiners": 25,
    "Switching": "Normal",
    "MinerWindowStyle": "Minimized",
    "ApiKey": "Api Key ID",
    "ConfirmMiner": false,
    "LowerFloor": { "CPU": 0.00001, "nVidia": { "USD": 3 }, "AMD": { "EUR": 2 } },
    "DevicesStatus": true,
    "ElectricityPrice": { "USD": { "7": 0.1, "23": 0.02 } },
    "ElectricityConsumption": false,
    "MaximumAllowedGrowth": 2,
    "DefaultCPUCores": 5,
    "DefaultCPUThreads": 10
}
```

* ***Region*** [enum] (**Europe**|Usa|China|Japan|Other) - pool region.
* ***SSL*** [bool] (**true**|false) - use secure protocol if possible.
* **Wallet** [key value collection] - coin wallet addresses (now support wallets: `BTC`, `LTC`, `NiceHashNew`, `NiceHash` and other. See specific pools option `Wallet`):
    * **Key** [string] - coin short name (if specified `"NiceHashNew"` or `"NiceHash"` wallet it use on NiceHash (New from 2019.07.01)).
    * **Value** [string] - coin wallet address.
* ***WorkerName*** [string] - worker name. If empty use machine name.
* **Login** [string] - login for pool with registration (MiningPoolHub).
* ***Password*** [string] - password. If empty default value `"x"`.
* ***CheckTimeout*** [int] - check timeout in seconds for read miner speed. Recomended value from 3 seconds to 15 secounds.
* ***LoopTimeout*** [int] - loop timeout in second. Recomended value from 60 seconds to five minute.
* ***NoHashTimeout*** [int] - timeout in minutes to disable miner after determining zero hash.
* ***ShowBalance*** [bool] - show balance if value equal true, else dont show.
* ***ShowExchangeRate*** [bool] - show exchage rate balance if value equal true, else dont show only if `ShowBalance` is enabled.
* ***AverageCurrentHashSpeed*** [int] - miner average current hash speed in seconds. Recomended value from 120 second to five minute.
* ***AverageHashSpeed*** [string] - miner average hash speed in [time interval](https://github.com/Quake4/HumanInterval/blob/master/README.md). Recomeded value from few hours to one day.
* ***Verbose*** [enum] (Full|**Normal**|Minimal) - verbose level.
* ***AllowedTypes*** [enum array] (CPU|nVidia|AMD|Intel) - allowed devices to mine.
* ***Currencies*** [key value collection] - currencies for output (maximum supported 3). If empty use by default `{ "BTC": 8, "USD": 2}`:
    * **Key** [string] - currency name from [supported list](https://api.coinbase.com/v2/exchange-rates?currency=BTC) + `mBTC`.
    * **Value** [int] - the number of digits after the decimal point.
* ***CoolDown*** [int] - the number of seconds to wait when switching miners.
* ***ApiServer*** [bool] - start local api server for get api pools info in proxy mode or show MindMiner status.
* ***ApiServerAllowWallets*** [bool] - allow publish wallets, login and password data on server page and api.
* ***SwitchingResistance*** [key value collection] - switching resistance. If it is enabled, the switching is performed if the percentage or timeout is exceeded.
    * **Enabled** [bool] (**true**|false) - enable or disable the switching resistance between miners.
    * **Percent** [decimal] (4) - the percentage of switching. Must be a greater then zero.
    * **Timeout** [int] (15) - the switching timeout in minutes. Must be a greater then **LoopTimeout** in munutes.
* ***BenchmarkSeconds*** [key value collection] - global default timeout in seconds of benchmark for device type. If set, it overrides the miner configuration:
    * **Key** [string] - (CPU|nVidia|AMD|Intel) device type.
    * **Value** [int] - timeout in seconds of benchmark.
* ***MinimumMiners*** [int] (25) - minimum number of miners on the pool algorithm to use. Only for yiimp like pools.
* ***Switching*** [enum] (**Normal**|Fast) - the mode of operation of the program in which either the profit averaging (Normal) is used or not (Fast).
* ***MinerWindowStyle*** [enum] (Hidden|Maximized|**Minimized**|Normal) - specifies the state of the window that is used for starting the miner.
* ***ApiKey*** [string] - Api Key ID for online monitoring the rigs on [MindMiner site](http://mindminer.online/monitoring).
* ***ConfirmMiner*** [bool] (true|**false**) - need user confirm for the miner without configuration file (false - auto download new miners).
* ***ConfirmBenchmark*** [bool] (**true**|false) - need user confirm for the miner benchmark (false - auto benchmark new miners).
* ***LowerFloor*** [key value collection] - the mining profitability lower floor: 
    * **Key** [string] - (CPU|nVidia|AMD|Intel) device type.
    * **Value** [decimal] or [key value] - if number it value in BTC or currency key value pair (`"XXX": 2`) where `XXX` is any [supported currency](https://api.coinbase.com/v2/exchange-rates?currency=BTC).
* ***DevicesStatus*** [bool] (**true**|false) - retreive and display devices status from `nvidia-smi` and `overdriven`.
* ***ElectricityPrice*** (**null**) - electricity price. If single-rate system" `{ "XXX": 0.1 }` or multirate `{ "XXX": { "5": 0.05, "7": 0.1, "18": 0.12, "21.5": 0.03, "23.5": 0.02 } }` where `USD` is any [supported currency](https://api.coinbase.com/v2/exchange-rates?currency=BTC) and in multirate system the key is start hour of new tarif value (higest hour working up to lowest hour, 23.5 is equal 23:30). For show electricity cost must be enabled `DevicesStatus` and working `nvidia-smi` and `overdriven`.
* ***ElectricityConsumption*** [bool] (true|**false**) - includes (substract) in profit the accounting of cost of electricity. Must be enabled `DevicesStatus`, specified `ElectricityPrice` and working `nvidia-smi` and `overdriven`.
* ***MaximumAllowedGrowth*** [decimal] (1.25 - 5, 2) - Maximum possible growth of API pools data values.
* ***DefaultCPUCores*** [int] - the default number of cores for CPU mining (`-t x` param of cpu miner).
* ***DefaultCPUThreads*** [int] - the default number of threads for CPU mining (`-t x` param of cpu miner).

## Algorithms
MindMiner algorithms settings placed in algorithms.txt file into root application folder.

Algorithms settings read on each loop. You may change configuration at any time and it will be applied on the next loop. If you delete algorithms config or change to wrong json format it will be created default on the next loop.

```json
{
    "Difficulty": { "X16r": 48, "X16s": 48, "Phi": 128000 },
    "EnabledAlgorithms": [ "Bitcore", "X17", "X16r" ],
    "DisabledAlgorithms": [ "Blake2s" ],
    "RunBefore": { "Ethash": "fastmem.bat", "X16r": "memminus500.bat" },
    "RunAfter": "normalmem.bat"
}
```

* ***Difficulty*** [key value collection] - algorithms difficulties (as `d=XXX` in miner password parameter).
    * **Key** [string] - algorithm name.
    * **Value** [decimal] - difficulty value.
* ***EnabledAlgorithms*** [string array] - set of enabled (prioritized) algorithms. If the value is null or empty, this means that all algorithms are enabled from the all pools otherwise only the specified algorithms are prioritized on all pools.
* ***DisabledAlgorithms*** [string array] - set of disabled algorithms. Always disables the specified algorithms on all pools.
* ***RunBefore*** - command line to run before start of miner in folder ".\Run". More priority than in the configuration of the miner.
    * or [key value collection]
        * **Key** [string] - algorithm name.
        * **Value** [string] - command line.
	* or [string] - command line for any algorithm.
* ***RunAfter*** - command line to run after end of miner in folder ".\Run". More priority than in the configuration of the miner.
    * or [key value collection]
        * **Key** [string] - algorithm name.
        * **Value** [string] - command line.
	* or [string] - command line for any algorithm.

## Pools
Pools configuration placed in Pools folder and named as pool name and config extension.

Pools settings read on each loop. You may change configuration at any time and it will be applied on the next loop. If you delete pool config or change to wrong json format it will be created default on the next loop after your confirm and answer at console window.

Look like this "PoolName.config.txt".

Any pool has this config (exlude ApiPoolsProxy, see it section):
```json
{
    "AverageProfit": "1 hour 30 min",
    "Enabled": false,
    "EnabledAlgorithms": [ "Bitcore", "X17", "X16r" ],
    "DisabledAlgorithms": null
}
```

* **Enabled** [bool] (true|false) - enable or disable pool for mine.
* **AverageProfit** [string] - averages a profit on the coins at the specified [time interval](https://github.com/Quake4/HumanInterval/blob/master/README.md).
* ***EnabledAlgorithms*** [string array] - set of enabled (prioritized) algorithms. If the value is null or empty, this means that all algorithms are enabled from the pool otherwise only the specified algorithms are prioritized.
* ***DisabledAlgorithms*** [string array] - set of disabled algorithms. Always disables the specified algorithms.

### Specific for MiningPoolHub
* ***APiKey*** [string] - api key for get balance on MiningPoolHub. See "Edit Account" section and "API KEY" value in MPH account.

### Specific for NiceHash
* ***Region*** [string] (eu|usa|hk|jp|in|br) - stratum region replace of region main settings file.

Example, replace main region to usa:
```json
{
    "AverageProfit": "20 min",
    "Enabled": true,
    "Region": "usa"
}
```

### Specific for BlockMasters, ZergPool and ZPool
* ***SpecifiedCoins*** [array] - specifing preferred coin for algo. (Algo as key and sign of coin as value or array of value for several sign of coins) If add "only" to the array of coin signs, only the specified coin will be used (see `X17` algo and `XVG` sign of coin).

Example:
```json
{
    "AverageProfit": "1 hour 30 min",
    "Enabled": true,
    "SpecifiedCoins": { "NeoScrypt": [ "SPK", "GBX"], "Phi": "LUX", "X17": [ "XVG", "only" ] }
}
```

If algo has two or three conis you must specify one coin. If it coin down then MindMiner to be mine just algo without specified coin (example Phi algo need specify only LUX, not need specify together FLM).
This feature give you a very great opportunity to increase profit.
The BlockMaster is support only one coin (`"Phi": "LUX"`).

### Specific for BlockMasters and ZergPool
* ***PartyPassword*** [string] - password for party mode.

Solo mode support if add "solo" to the array of coin signs (as `m=solo` in miner password parameter).

Example:
```json
{
    "AverageProfit": "1 hour",
    "Enabled": true,
    "SpecifiedCoins": { "Argon2-dyn": [ "DYN", "solo" ] }
}
```

Party mode support if add "party" to the array of coin signs (as `m=party.password` in miner password parameter).

Example:
```json
{
    "AverageProfit": "1 hour",
    "Enabled": true,
    "PartyPassword": "password",
    "SpecifiedCoins": { "X16rt": [ "GIN", "party" ] }
}
```

### Specific for ZergPool, ZPool & BlockMasters
* ***Wallet*** [string] - coin short name (example `"LTC"`) to use on the pool (as `c=XXX` in miner password parameter). Wallet address must be specified in main settings file.

Example:
```json
{
    "AverageProfit": "1 hour",
    "Enabled": true,
    "Wallet": "LTC"
}
```

### ApiPoolsProxy (Master/Slave)
If you have more then ten rigs, some pools can block api requests because there will be a lot of requests to prevent ddos attacks. For proper operation MindMiner need to use the api pools proxy. Define at least two rigs (Master) to send (Slave) information about the api pools data.
* Change on Master main configuration by adding `"ApiServer": true` (see `MindMiner config` section) and rerun MindMiner as Administrator.
* Change on Slave ApiPoolsProxy configuration: enable it and write names and/or IPs of Master rigs.

Example:
```json
{
    "Enabled": true,
    "ProxyList": [ "rig1", "rig2", "192.168.0.19" ]
}
```

* **Enabled** [bool] (true|false) - enable or disable use api pools proxy.
* **ProxyList** [string array] - set of rig names or IP addresses where to send a request the api pools data.

The Slave rigs will have settings of pools made on the Master rig. In the absence of a response from one Master rig, Slave rig will be switched for the following Master rig in the proxy list.

### MiningRigRentals (MRR)
You can lease your rig at MiningRigRentals. You must create api key with read-only balance and manage rigs permitions for each rig is must be unique.

Example:
```json
{
    "Enabled": true,
    "Region": "eu",
    "FailoverRegion": "eu-de",
    "Key": "xxx",
    "Secret": "xxx",
    "DisabledAlgorithms": [ "Yescryptr16", "Yespower" ],
    "Wallets": [ "ETH", "LTC" ],
    "Target": 50,
    "TargetByAlgorithm": { "Ethash": 100 },
    "Decrease": 1,
    "Increase": 5,
    "MaxHours": 12,
    "MinHours": 3
}
```

* **Enabled** [bool] (true|false) - enable or disable use MiningRigRentals (1% extra fee).
* ***Region*** [string] (us-east|us-central|us-west|eu|eu-de|eu-ru|ap) - pool region.
* **FailoverRegion** [string] (us-east|us-central|us-west|eu|eu-de|eu-ru|ap) - pool failover region.
* **Key** [string] - api key from https://www.miningrigrentals.com/account/apikey.
* **Secret** [string] api secret from https://www.miningrigrentals.com/account/apikey.
* ***DisabledAlgorithms*** [string array] - set of disabled algorithms. Always disables the specified algorithms.
* ***Wallets*** [string array] (ETH|LTC|DASH|BCH) - set of payment coins.
* ***Target*** [int] (5-899, **50**) - target percent profit greater than rig profit on regular pools.
* ***TargetByAlgorithm*** [key value collection] - target percent profit by algorithm. If specified, the maximum of `Traget` and `TargetByAlgorithm` is used:
    * **Key** [string] - algorithm name.
    * **Value** [int] (5-899) - target percent profit.
* ***Decrease*** [int] (0-25, **1**) - decrease percent rental price every hour if not rented.
* ***Increase*** [int] (0-25, **5**) - increase percent rental price after every rent.
* ***MinHours*** [int] (3-120, **4**) - minimum amount of hours to rent your rig.
* ***MaxHours*** [int] (3-120, **12**) - maximum amount of hours to rent your rig.

## Miners
Miners configuration placed in Miners folder and named as miner name and config extension.

Miners settings read on each loop. You may change configuration at any time and it will be applied on the next loop. If you delete miner config or change to wrong json format it will be created default on the next loop.

Look like this "MinerName.config.txt".

Simple miner config:
```json
{
    "Algorithms": [
                       {
                           "ExtraArgs": null,
                           "BenchmarkSeconds": 0,
                           "Enabled": true,
                           "Algorithm": "cryptonight",
                           "RunBefore": "cn.bat 123 345",
                           "RunAfter": "\"..\\..\\My Downloads\\cn.bat\" 123 123"
                       },
                       {
                           "ExtraArgs": "-lite",
                           "BenchmarkSeconds": 0,
                           "Enabled": true,
                           "Algorithm": "cryptolite"
                       }
                   ],
    "ExtraArgs": null,
    "BenchmarkSeconds": 60,
    "Enabled": true
}
```

 * common:
     * **Enabled** [bool] (true|false) - enable or disable miner.
     * ***ExtraArgs*** [string] - miner extra parameters for all algorithms.
     * ***BenchmarkSeconds*** [int] - default timeout in seconds for benchmark for any algorithm. If not set or zero must be set algorithm BenchmarkSeconds.
 * algorithms miners:
     * **Algorithms** [array] - array of miner algorithms.
         * **Enabled** [bool] (true|false) - enable or disable algorithm.
         * **Algorithm** [string] - pool algorithm and miner algorithm parameter.
         * ***DualAlgorithm*** [string] - pool algorithm and miner algorithm parameter for dual mining (only in claymore dual miner).
         * ***ExtraArgs*** [string] - algorithm extra parameters in additional to common ExtraArgs.
         * ***BenchmarkSeconds*** [int] - default timeout in seconds for benchmark for current algorithm. If not set or zero use common BenchmarkSeconds.
         * ***RunBefore*** [string] - full command line to run before start of miner in folder ".\Run".
         * ***RunAfter*** [string] - full command line to run after end of miner in folder ".\Run".
