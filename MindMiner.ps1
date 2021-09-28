<#
MindMiner  Copyright (C) 2017-2021  Oleg Samsonov aka Quake4
https://github.com/Quake4/MindMiner
License GPL-3.0
#>

. .\Code\Out-Data.ps1

Out-Iam
Write-Host "Loading ..." -ForegroundColor Green

$global:HasConfirm = $false
$global:NeedConfirm = $false
$global:AskPools = $false
$global:HasBenchmark = $false
$global:MRRHour = $false
$global:MRRRented = @()
$global:MRRRentedTypes = @()
$global:API = [hashtable]::Synchronized(@{})
$global:Admin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

. .\Code\Include.ps1

# ctrl+c hook
[Console]::TreatControlCAsInput = $true
[Console]::Title = "MindMiner $([Config]::Version.Replace("v", [string]::Empty)) - $([datetime]::Now.ToString())"

$BinLocation = [IO.Path]::Combine($(Get-Location), [Config]::BinLocation)
New-Item $BinLocation -ItemType Directory -Force | Out-Null
$BinScriptLocation = [scriptblock]::Create("Set-Location('$BinLocation')")
$DownloadJob = $null

# download prerequisites
Get-Prerequisites ([Config]::BinLocation)

# read and validate config
$Config = Get-Config

if (!$Config) { exit }

if ($Config.DevicesStatus) {
	$Devices = Get-Devices ([Config]::ActiveTypes)
}
elseif ([Config]::ActiveTypes -contains [eMinerType]::CPU) {
	$Devices = Get-Devices (@([eMinerType]::CPU))
}

# define cores/threads
if ([Config]::ActiveTypes -contains [eMinerType]::CPU) {
	$cpu = $Devices[[eMinerType]::CPU]
	[nullable[int]] $threads = $null
	if ($Config.DefaultCPUThreads -is [int]) {
		$threads = [math]::Min($Config.DefaultCPUThreads, $cpu.Threads)
	}
	[nullable[int]] $cores = $null
	if ($Config.DefaultCPUCores -is [int]) {
		$cores = [math]::Min($Config.DefaultCPUCores, $cpu.Cores)
	}
	if ($cores -and !$threads) {
		$threads = [int][math]::Min($cores * $cpu.Threads / $cpu.Cores, $cpu.Threads)
	}
	elseif (!$cores -and $threads) {
		$cores = [int][math]::Min($threads * $cpu.Cores / $cpu.Threads, $cpu.Cores)
	}
	if ($threads -and $cores) {
		[Config]::DefaultCPU = [CPUConfig]::new($cores, $threads)
	}
	Remove-Variable threads, cores, cpu
}

[SummaryInfo] $Summary = [SummaryInfo]::new([Config]::RateTimeout)
$Summary.TotalTime.Start()

Clear-Host
Out-Header

$ActiveMiners = [Collections.Generic.Dictionary[string, MinerProcess]]::new()
$KnownAlgos = [Collections.Generic.Dictionary[eMinerType, Collections.Generic.Dictionary[string, SpeedProfitInfo]]]::new()
[Config]::ActiveTypes | ForEach-Object {
	$KnownAlgos.Add($_, [Collections.Generic.Dictionary[string, SpeedProfitInfo]]::new())
}
[StatCache] $Statistics = [StatCache]::Read([Config]::StatsLocation)
if ($Config.ApiServer) {
	if ([Net.HttpListener]::IsSupported) {
		if ($global:Admin) {
			Write-Host "Starting API server at port $([Config]::ApiPort) for Remote access ..." -ForegroundColor Green
		}
		else {
			Write-Host "Starting API server at port $([Config]::ApiPort) for Local access ..." -ForegroundColor Green
			Write-Host "To start API server for remote access run MindMiner as Administrator." -ForegroundColor Yellow
		}
		Start-ApiServer
	}
	else {
		Write-Host "Http listner not supported. Can't start API server." -ForegroundColor Red
	}
}

if ($global:API.Running) {
	$global:API.Worker = $Config.WorkerName
	$global:API.Config = ($Config.Web($global:Admin) | ConvertTo-Html -Fragment).Replace("<tr><th>*</th></tr>", "<tr><th>Region</th></tr>")
	$global:API.Wallets = $Config.Api()
}

# FastLoop - variable for benchmark or miner errors - very fast switching to other miner - without ask pools and miners
[bool] $FastLoop = $false 
# exit - var for exit
[bool] $exit = $false
# main loop
while ($true)
{
	if ($Summary.RateTime.IsRunning -eq $false -or $Summary.RateTime.Elapsed.TotalSeconds -ge [Config]::RateTimeout.TotalSeconds) {
		$exit = Update-Miner
		if ($exit -eq $true) {
			$FastLoop = $true
		}
		else {
			$Rates = Get-RateInfo
			if ($Summary.RateTime.IsRunning) {
				$global:MRRHour = $true
			}
		}
		$Summary.RateTime.Reset()
		$Summary.RateTime.Start()
	}
	elseif (!$Rates -or $Rates.Count -eq 0) {
		$Rates = Get-RateInfo
	}

	if (!$FastLoop) {
		# read algorithm mapping
		$AllAlgos = [BaseConfig]::ReadOrCreate("algorithms.txt", @{
			EnabledAlgorithms = $null
			DisabledAlgorithms = $null
			Difficulty = $null
			RunBefore = $null
			RunAfter = $null
		})
		# how to map algorithms
		$AllAlgos.Add("Mapping", [ordered]@{
			"aeternity" = "CuckooCycle"
			"argon2d250" = "Argon2-crds"
			"argon2d-crds" = "Argon2-crds"
			"argon2d500" = "Argon2-dyn"
			"argon2d-dyn" = "Argon2-dyn"
			"argon2d" = "Argon2-dyn"
			"argon2d_dynamic" = "Argon2-dyn"
			"autolykosv2" = "Autolykos2"
			"autolykos" = "Autolykos2"
			"beamhash" = "Beam"
			"beamhashII" = "BeamV2"
			"beamhash2" = "BeamV2"
			"beamhashIII" = "BeamV3"
			"beamhash3" = "BeamV3"
			"beamv2" = "BeamV2"
			"beamv3" = "BeamV3"
			"binarium_hash_v1" = "Binarium-V1"
			"blakecoin" = "Blake"
			"blake256r8" = "Blake"
			"blake2b-btcc" = "Blake2b"
			"bl2bsha3" = "Handshake"
			"blake2bsha3" = "Handshake"
			"blake2skadena" = "Kadena"
			"hns" = "Handshake"
			"trtl_chukwa" = "Chukwa"
			"trtl_chukwa2" = "Chukwa2"
			"argon2/chukwa" = "Chukwa"
			"argon2/chukwav2" = "Chukwa2"
			"argon2dchukwa" = "Chukwa"
			"argon2id_chukwa" = "Chukwa"
			"argon2id_chukwa2" = "Chukwa2"
			"chukwav2" = "Chukwa2"
			"chukwa2" = "Chukwa2"
			"randomkeva" = "RandomKeva"
			"randomx_keva" = "RandomKeva"
			"rx/keva" = "RandomKeva"
			"randomarq" = "RandomARQ"
			"randomx_arqma" = "RandomARQ"
			"randomsfx" = "RandomSFX"
			"randomx_safex" = "RandomSFX"
			"randomv" = "RandomV"
			"randomx" = "RandomX"
			"RandomXmonero" = "RandomX"
			"rx/0" = "RandomX"
			"rx/arq" = "RandomARQ"
			"rx/sfx" = "RandomSFX"
			"rx/v" = "RandomV"
			"cryptonotewow" = "RandomWOW"
			"rx/wow" = "RandomWOW"
			"randomwow" = "RandomWOW"
			"randomx_wow" = "RandomWOW"
			"cryptonightrxl" = "RandomXL"
			"rx/loki" = "RandomXL"
			"randomx_loki" = "RandomXL"
			"randomxl" = "RandomXL"
			"cn/superfast" = "cnSFast"
			"cryptonight_superfast" = "cnSFast"
			"cryptonotefh" = "cnSFast"			
			"cryptonight_v8_reversewaltz" = "cnRWltz"
			"cryptonightrw" = "cnRWltz"
			"cryptonight_rw" = "cnRWltz"
			"cn/rwz" = "cnRWltz"
			"cryptonight_v8_double" = "cnXCash"
			"cryptonotev8d" = "cnXCash"
			"cn/double" = "cnXCash"	
			"cn/r" = "CryptonightR"
			"cn/gpu" = "CryptonightGPU"
			"cngpu" = "CryptonightGPU"
			"cnheavy" = "cnHeavy"
			"cn_saber" = "cnSaber"
			"cnsaber" = "cnSaber"
			"cn-heavy/tube" = "cnSaber"
			"cryptonoteh" = "cnHeavy"
			"cryptonight_haven" = "cnHaven"
			"cryptonight_xhv" = "cnHaven"
			"cryptonote_haven" = "cnHaven"
			"cryptonotehaven" = "cnHaven"
			"cn_haven" = "cnHaven"
			"cnhaven" = "cnHaven"
			"cn-heavy/xhv" = "cnHaven"
			"cryptonight_v8_zelerius" = "cnZls"
			"cryptonotev8zls" = "cnZls"
			"cnzls" = "cnZls"
			"cn/zls" = "cnZls"
			"cryptonight_upx" = "cnUpx"
			"cryptonightupx" = "cnUpx"
			"cnupx" = "cnUpx"
			"cnupx2" = "cnUpx"
			"cn/upx2" = "cnUpx"			
			"cnv8_upx2" = "cnUpx"
			"cryptonotextl" = "cnFastV2"
			"cnfast2" = "cnFastV2"
			"cryptonight_fast" = "cnFastV2"
			"cryptonight_masari" = "cnFast"
			"cryptonotefast" = "cnFast"
			"cn/fast" = "cnFast"
			"cnfast" = "cnFast"
			"cn/ccx" = "cnConceal"
			"cn_conceal" = "cnConceal"
			"cnconceal" = "cnConceal"
			"cryptonotec" = "cnConceal"
			"cryptonight_conceal" = "cnConceal"
			"cryptonight_talleo" = "cnTalleo"
			"cryptonightulv2" = "cnTalleo"
			"cryptonighttlo" = "cnTalleo"
			"cn-pico/tlo" = "cnTalleo"
			"cntlo" = "cnTalleo"
			"cryptonoteturtlev2" = "cnTurtle"
			"cryptonight_turtle" = "cnTurtle"
			"cnturtle" = "cnTurtle"
			"cn-pico" = "cnTurtle"
			"cnv7" = "Cryptonightv7"
			"cnv8" = "Cryptonightv8"
			"cnr" = "CryptonightR"
			"cryptonotegpu" = "CryptonightGPU"
			"cryptonoter" = "CryptonightR"
			"cryptonotev7" = "Cryptonightv7"
			"cryptonotev8" = "Cryptonightv8"
			"cryptonight_hvy" = "cnHeavy"
			"cryptonight_gpu" = "CryptonightGPU"
			"cryptonight_heavy" = "cnHeavy"
			"cryptonight_heavyx" = "cnHeavy"
			"cryptonight_lite_v7" = "Cryptolightv7"
			"cryptonight-monero" = "CryptonightR"
			"cryptonight_v7" = "Cryptonightv7"
			"cryptonight_v8" = "Cryptonightv8"
			"cryptonight_r" = "CryptonightR"
			"cryptonight_saber" = "cnSaber"
			"cryptonightr" = "CryptonightR"
			"cryptonightheavy" = "cnHeavy"
			"cryptonightheavysaber" = "cnSaber"
			"conflux" = "Octopus"
			"cuckoo_ae" = "CuckooCycle"
			"cuckooaeternity" = "CuckooCycle"
			"cuckoocycle" = "CuckooCycle"
			"cuckoocycleo" = "Grin29"
			"cuckoocycle29swap" = "Swap"
			"cuckoocycle31" = "Grin31"
			"cuckaroo_swap" = "Swap"
			"cuckoo24" = "Cuckaroo24"
			"cuckaroom29_qitmeer" = "Qitmeer"
			"cuckaroo" = "Grin29"
			"cuckaroo29" = "Grin29"
			"cuckarood" = "Grind29"
			"cuckaroo29bfc" = "Bfc"
			"cuckaroo29d" = "Grind29"
			"cuckaroo29m" = "Cuckaroom"
			"cuckarood29" = "Grind29"
			"cuckarood29_grin" = "Grind29"
			"cuckaroom29" = "Cuckaroom"
			"cuckaroo29z" = "Cuckarooz"
			"cuckarooz29" = "Cuckarooz"
			"cuckatoo" = "Grin31"
			"cuckatoo31" = "Grin31"
			"cuckatoo31_grin" = "Grin31"
			"cuckatoo32" = "Grin32"
			"cuckoocycle32" = "Grin32"
			"Grin" = "Grin29"
			"GrinCuckaroo29" = "Grin29"
			"GrinCuckarood29" = "Grind29"
			"GrinCuckatoo31" = "Grin31"
			"GrinCuckatoo32" = "Grin32"
			"dagger" = "Ethash"
			"daggerhashimoto" = "Ethash"
			"jackpot" = "JHA"
			"hashimotos" = "Ethash"
			"equihash1254" = "Equihash125"
			"equihash125_4" = "Equihash125"
			"equihash144_5" = "Equihash144"
			"equihash1505" = "BeamV3"
			"equihash1505g" = "Grimm"
			"equihash192_7" = "Equihash192"
			"equihash1927" = "Equihash192"
			"aion" = "Equihash210"
			"equihash210_9" = "Equihash210"
			"equihash2109" = "Equihash210"
			"equihash96_5" = "Equihash96"
			"Equihash-BTG" = "EquihashBTG"
			"equihashBTG" = "EquihashBTG"
			"Equihash-ZCL" = "EquihashZCL"
			"equihashZCL" = "EquihashZCL"
			"ergo" = "Autolykos2"
			"ethereum" = "Ethash"
			"ethereum-classic" = "Ethash"
			"gr" = "Ghostrider"
			"glt-astralhash" = "Astralhash"
			"glt-globalhash" = "Globalhash"
			"glt-jeonghash" = "Jeonghash"
			"glt-padihash" = "Padihash"
			"glt-pawelhash" = "Pawelhash"
			"lyra2rev2" = "Lyra2re2"
			"lyra2r2" = "Lyra2re2"
			"lyra2v2" = "Lyra2re2"
			"lyra2v2-old" = "Lyra2re2"
			"lyra2rev3" = "Lyra2v3"
			"lyra2re3" = "Lyra2v3"
			"lyra2r3" = "Lyra2v3"
			# "monero" = "Cryptonightv7"
			"m7m" = "M7M"
			"m7mv2" = "M7M"
			"mgroestl" = "MyrGr"
			"myriad-groestl" = "MyrGr"
			"myriadgroestl" = "MyrGr"
			"myr-gr" = "MyrGr"
			"neoscrypt" = "NeoScrypt"
			"phi1612" = "Phi"
			"poly" = "Polytimos"
			"progpowere" = "ProgpowEre"
			"progpow-ethercore" = "ProgpowEre"
			"progpowveil" = "ProgpowVeil"
			"progpow-veil" = "ProgpowVeil"
			"randomgrft" = "Graft"
			"rx/graft" = "Graft"
			"raven" = "Kawpow"
			"rfv2" = "RainForest2"
			"scryptn2" = "ScryptN2"
			"nscryptv" = "ScryptN2"
			"sha3" = "Keccak"
			"sib" = "X11Gost"
			"sibcoin" = "X11Gost"
			"sibcoin-mod" = "X11Gost"
			"skeincoin" = "Skein"
			"skunkhash" = "Skunk"
			"timetravel10" = "Bitcore"
			"ubqhash" = "Ubiqhash"
			"vit" = "Vitalium"
			"verus" = "Verushash"
			"x11gost" = "X11Gost"
			"x11evo" = "X11Evo"
			"x13bcd" = "Bcd"
			"x13sm3" = "Hsr"
			"x16rtgin" = "X16rt"
			"yespower2b" = "Power2b"
		})
		# disable asic algorithms
		$AllAlgos.Add("Disabled", @("bcd", "beam", "bitcore", "blake", "blake2b", "blake2s", "handshake", "kadena", "sha256", "sha256t", "sha256asicboost", "sha256-ld", "sha3d", "scrypt", "scrypt-ld", "tensority", "x11", "x11-ld", "x13", "x14", "x15", "quark", "qubit", "myrgr", "lbry", "decred", "sia", "blake", "nist5", "cryptonight", "cryptonightr", "cryptonightv7", "cryptonightv8", "cnheavy", "cnsaber", "x11gost", "groestl", "eaglesong", "equihash", "lyra2re2", "lyra2z", "pascal", "keccak", "keccakc", "skein", "tribus", "c11", "phi", "timetravel", "skunk"))

		# ask needed pools
		if ($global:AskPools -eq $true) {
			$AllPools = Get-PoolInfo ([Config]::PoolsLocation)
			$global:AskPools = $false
		}
		Write-Host "Pool(s) request ..." -ForegroundColor Green
		$AllPools = Get-PoolInfo ([Config]::PoolsLocation)

		# check pool exists
		if (!$AllPools -or $AllPools.Length -eq 0) {
			Write-Host "No Pools!" -ForegroundColor Red
			Get-Confirm
			continue
		}
		
		Write-Host "Miners request ..." -ForegroundColor Green
		$AllMiners = Get-ChildItem ([Config]::MinersLocation) | Where-Object Extension -eq ".ps1" | ForEach-Object {
			Invoke-Expression "$([Config]::MinersLocation)\$($_.Name)"
		}

		# filter by exists hardware
		$AllMiners = $AllMiners | Where-Object { [Config]::ActiveTypes -contains ($_.Type -as [eMinerType]) }

		# download miner
		if ($DownloadJob -and $DownloadJob.State -ne "Running") {
			$DownloadJob | Remove-Job -Force | Out-Null
			$DownloadJob.Dispose()
			$DownloadJob = $null
		}
		$DownloadMiners = $AllMiners | Where-Object { !$_.Exists([Config]::BinLocation) } | Select-Object Name, Path, URI -Unique
		if ($DownloadMiners -and ($DownloadMiners.Length -gt 0 -or $DownloadMiners -is [PSCustomObject])) {
			Write-Host "Download miner(s): $(($DownloadMiners | Select-Object Name -Unique | ForEach-Object { $_.Name }) -Join `", `") ... " -ForegroundColor Green
			if (!$DownloadJob) {
				$PathUri = $DownloadMiners | Select-Object Path, URI, Pass -Unique;
				$DownloadJob = Start-Job -ArgumentList $PathUri -FilePath ".\Code\Downloader.ps1" -InitializationScript $BinScriptLocation
			}
		}

		# check exists miners & update bench timeout by global value
		$AllMiners = $AllMiners | Where-Object { $_.Exists([Config]::BinLocation) } | ForEach-Object {
			if ($Config.BenchmarkSeconds -and $Config.BenchmarkSeconds."$($_.Type)" -gt $_.BenchmarkSeconds) {
				$_.BenchmarkSeconds = $Config.BenchmarkSeconds."$($_.Type)"
			}
			[MinerInfo][MinerProfitInfo]::CopyMinerInfo($_, $Config)
		}
		
		if ($AllMiners.Length -eq 0) {
			Write-Host "No Miners!" -ForegroundColor Red
			Get-Confirm
			continue
		}

		# save speed active miners
		$ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running -and $_.Action -eq [eAction]::Normal } | ForEach-Object {
			$speed = $_.GetSpeed($false)
			if ($speed -gt 0) {
				$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey(), $speed, $Config.AverageHashSpeed, 0.25)
				if (![string]::IsNullOrWhiteSpace($_.Miner.DualAlgorithm)) {
					$speed = $_.GetSpeed($true)
					if ($speed -gt 0) {
						$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey($true), $speed, $Config.AverageHashSpeed, 0.25)
					}
				}
			}
			elseif ($speed -eq 0 -and $_.CurrentTime.Elapsed.TotalSeconds -ge ($_.Miner.BenchmarkSeconds * $(if ($_.Miner.Priority -ge [Priority]::Solo) { 5 } else { 2 }))) {
				# no hasrate stop miner and move to nohashe state while not ended
				$_.Stop($AllAlgos.RunAfter)
			}
		}

		$KnownAlgos.Values | ForEach-Object { $_.Clear() }
		[Config]::SoloParty.Clear()
	}

	# get devices status
	if ($Config.DevicesStatus -and !$FastLoop) {
		$Devices = Get-Devices ([Config]::ActiveTypes) $Devices

		# power draw save
		if (Get-ElectricityPriceCurrency) {
			$Benchs = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running -and ($_.CurrentTime.Elapsed.TotalSeconds * 2) -ge $_.Miner.BenchmarkSeconds } | ForEach-Object {
				$measure = $Devices["$($_.Miner.Type)"] | Measure-Object Power -Sum
				if ($measure) {
					$draw = [decimal]$measure[0].Sum
					if ($draw -gt 0) {
						$_.SetPower($draw)
						$draw = $Statistics.SetValue($_.Miner.GetPowerFilename(), $_.Miner.GetKey(), $draw, $Config.AverageHashSpeed)
					}
					Remove-Variable draw
				}
				Remove-Variable measure
			}
		}
	}

	$Running = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running }

	# stop benchmark by condition: timeout reached and has result or timeout more then twice and no result
	$Benchs = $Running | Where-Object { $_.Action -eq [eAction]::Benchmark }
	if ($Benchs) { Get-Speed $Benchs } # read speed from active miners
	$Benchs | ForEach-Object {
		$speed = $_.GetSpeed($false)
		if (($_.CurrentTime.Elapsed.TotalSeconds -ge $_.Miner.BenchmarkSeconds -and $speed -gt 0) -or
			($_.CurrentTime.Elapsed.TotalSeconds -ge ($_.Miner.BenchmarkSeconds * 2) -and $speed -eq 0)) {
			$_.Stop($AllAlgos.RunAfter)
			if ($speed -eq 0) {
				$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey(), -1)
			}
			else {
				$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey(), $speed, $Config.AverageHashSpeed)
				if (![string]::IsNullOrWhiteSpace($_.Miner.DualAlgorithm)) {
					$speed = $_.GetSpeed($true)
					$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey($true), $speed, $Config.AverageHashSpeed)
				}
			}
		}
	}
	Remove-Variable Benchs
	
	# protection switching between pools
	if (!$FastLoop) {
		$Running = $Running | Where-Object { $_.State -eq [eState]::Running -and (Get-PoolInfoEnabled $_.Miner.PoolKey $_.Miner.Algorithm $_.Miner.DualAlgorithm ) } |
			ForEach-Object { $_.Miner } | Where-Object { 
				$r = $_
				# no resistance between unique
				if ($r.Priority -ge [Priority]::Solo) { $false }
				else {
					$null -ne ($AllMiners | Where-Object {
						$r.PoolKey -ne $_.PoolKey -and
						$r.Priority -eq $_.Priority -and
						$r.Name -eq $_.Name -and
						$r.Algorithm -eq $_.Algorithm -and
						$r.Type -eq $_.Type
					})
				}
			}
		if ($Running -and $Running.Length -gt 0) {
			$AllMiners += $Running
		}
	}
	Remove-Variable Running
	
	# read speed and price of proposed miners
	$AllMiners = $AllMiners | ForEach-Object {
		if (!$FastLoop) {
			$speed = $Statistics.GetValue($_.GetFilename(), $_.GetKey())
			# filter unused
			if ($speed -ge 0) {
				$price = (Get-PoolAlgorithmProfit $_.PoolKey $_.Algorithm $_.DualAlgorithm)
				if ($_.Priority -gt [Priority]::None -or ($_.Priority -eq [Priority]::None -and $price -gt 0 -and $speed -gt 0)) {
					[MinerProfitInfo] $mpi = $null
					if (![string]::IsNullOrWhiteSpace($_.DualAlgorithm)) {
						$mpi = [MinerProfitInfo]::new($_, $Config, $speed, $price[0], $Statistics.GetValue($_.GetFilename(), $_.GetKey($true)), $price[1])
					}
					else {
						$mpi = [MinerProfitInfo]::new($_, $Config, $speed, $price)
					}
					if ($speed -gt 0) {
						if ($_.Priority -eq [Priority]::Solo -and ![Config]::SoloParty.Contains($_.Type)) {
							[Config]::SoloParty.Add($_.Type)
						}
						if (!$KnownAlgos[$_.Type].ContainsKey($_.Algorithm)) {
							$KnownAlgos[$_.Type][$_.Algorithm] = [SpeedProfitInfo]::new()
						}
						$KnownAlgos[$_.Type][$_.Algorithm].SetValue($speed, $mpi.Profit, $_.Priority -eq [Priority]::None -or $_.Priority -eq [Priority]::Unique)
					}
					if ($Config.DevicesStatus -and (Get-ElectricityPriceCurrency)) {
						$mpi.SetPower($Statistics.GetValue($_.GetPowerFilename(), $_.GetKey()), (Get-ElectricityCurrentPrice "BTC"))
					}
					$mpi
				}
				Remove-Variable price
			}
		}
		elseif (!$exit) {
			$speed = $Statistics.GetValue($_.Miner.GetFilename(), $_.Miner.GetKey())
			# filter unused
			if ($speed -ge 0) {
				if (![string]::IsNullOrWhiteSpace($_.Miner.DualAlgorithm)) {
					$_.SetSpeed($speed, $Statistics.GetValue($_.Miner.GetFilename(), $_.Miner.GetKey($true)))
				}
				else {
					$_.SetSpeed($speed)
				}
				if ($Config.DevicesStatus -and (Get-ElectricityPriceCurrency)) {
					$_.SetPower($Statistics.GetValue($_.Miner.GetPowerFilename(), $_.Miner.GetKey()), (Get-ElectricityCurrentPrice "BTC"))
				}
				$_
			}
		}
	} |
	# reorder miners for proper output
	Sort-Object @{ Expression = { $_.Miner.Type } }, @{ Expression = { $_.Profit }; Descending = $true }, @{ Expression = { $_.Miner.GetExKey() } }

	if (!$exit) {
		Remove-Variable speed

		$global:HasBenchmark = $null -ne ($AllMiners | Where-Object { $_.Speed -eq 0 -and (($global:MRRRentedTypes -notcontains ($_.Miner.Type) -and
			[Config]::SoloParty -notcontains ($_.Miner.Type) -and $Summary.Loop -gt 1) -or $_.Miner.Priority -ge [Priority]::Solo) } | Select-Object -First 1)

		if ($global:HasConfirm -and !$global:HasBenchmark) {
			# reset confirm after all bench ends
			$global:HasConfirm = $false
		}

		$FStart = !$global:HasConfirm -and !($global:MRRRentedTypes) -and ($Summary.TotalTime.Elapsed.TotalSeconds / [Config]::Max) -gt ($Summary.FeeTime.Elapsed.TotalSeconds + [Config]::FTimeout)
		$FChange = $false
		if ($FStart -or $Summary.FeeCurTime.IsRunning) {
			if ($global:MRRRentedTypes -or ($Summary.TotalTime.Elapsed.TotalSeconds / [Config]::Max) -le ($Summary.FeeTime.Elapsed.TotalSeconds - [Config]::FTimeout)) {
				$FChange = $true
				$Summary.FStop()
			}
			elseif (!$Summary.FeeCurTime.IsRunning) {
				$FChange = $true
				$Summary.FStart()
			}
		}

		[Config]::MRRDelayUpdate = $global:MRRRentedTypes -or $Summary.FeeCurTime.IsRunning

		# look for run or stop miner
		[Config]::ActiveTypes | ForEach-Object {
			$type = $_

			# variables
			if (!$Summary.FeeCurTime.IsRunning) {
				$allMinersByType = $AllMiners | Where-Object { $_.Miner.Type -eq $type -and $_.Miner.Priority -ge [Priority]::Normal } |
					Sort-Object @{ Expression = { [int]($_.Miner.Priority) }; Descending = $true }, @{ Expression = { $_.Profit }; Descending = $true }, @{ Expression = { $_.Miner.GetExKey() } }
			}
			else {
				$allMinersByType = $AllMiners | Where-Object { $_.Miner.Type -eq $type -and $_.Miner.Priority -ge [Priority]::Normal -and $_.Miner.Pool -match [Config]::Pools } |
					Sort-Object @{ Expression = { $_.Profit }; Descending = $true }, @{ Expression = { $_.Miner.GetExKey() } }
			}
			$activeMinersByType = $ActiveMiners.Values | Where-Object { $_.Miner.Type -eq $type }
			$activeMinerByType = $activeMinersByType | Where-Object { $_.State -eq [eState]::Running }
			$activeMiner = if ($activeMinerByType) { $allMinersByType | Where-Object { $_.Miner.GetUniqueKey() -eq $activeMinerByType.Miner.GetUniqueKey() } } else { $null }

			# update pool info on site and benchmarkseconds for active miner
			if ($activeMiner -and $activeMinerByType -and $activeMiner.Miner.PoolKey -eq $activeMinerByType.Miner.PoolKey) {
				$activeMinerByType.Miner.Pool = $activeMiner.Miner.Pool
				$activeMinerByType.Miner.BenchmarkSeconds = $activeMiner.Miner.BenchmarkSeconds
			}

			# place current bench
			$run = $null
			if ($activeMinerByType -and $activeMinerByType.Action -eq [eAction]::Benchmark) {
				$run = $activeMinerByType
			}

			# find benchmark if not benchmarking
			if (!$run -and !$Summary.FeeCurTime.IsRunning) {
				$run = $allMinersByType | Where-Object { $_.Speed -eq 0 -and ($global:MRRRentedTypes -notcontains ($_.Miner.Type) -and
					[Config]::SoloParty -notcontains ($_.Miner.Type) -or $_.Miner.Priority -ge [Priority]::Solo)} | Select-Object -First 1
				if ($global:HasConfirm -eq $false -and $run) {
					if ($Config.ConfirmBenchmark) {
						$run = $null
						$global:NeedConfirm = $true
					}
					else {
						$global:HasConfirm = $true;
					}
				}
			}

			$lf = Get-ProfitLowerFloor $type

			# nothing benchmarking - get most profitable - exclude failed
			if (!$run) {
				$firstminer = $null
				$miner = $null
				$miners = @()
				$allMinersByType | ForEach-Object {
					if (!$run -and ($_.Profit -gt $lf -or $_.Miner.Priority -ge [Priority]::Solo)) {
						# skip failed or nohash miners
						if ($null -eq $firstminer) {
							$firstminer = $_
						}
						$miner = $_
						if ($miner.Miner.Algorithm -eq $firstminer.Miner.Algorithm -and $miner.Miner.Priority -eq $firstminer.Miner.Priority) {
							$miners += $miner.Miner.GetUniqueKey()
						}
						elseif ($firstminer.Miner.Priority -ge [Priority]::Solo) {
							$activeMinersByType | Where-Object { $miners -contains $_.Miner.GetUniqueKey() } | ForEach-Object {
								$_.ResetFailed()
							}
							$run = $firstminer;
						}
						if (!$run -and ($activeMinersByType | 
							Where-Object { ($_.State -eq [eState]::NoHash -or $_.State -eq [eState]::Failed) -and
								$miner.Miner.GetUniqueKey() -eq $_.Miner.GetUniqueKey() }) -eq $null) {
							$run = $miner
						}
					}
				}
				Remove-Variable firstminer, miner, miners
			}

			if ($run -and ($global:HasConfirm -or $FChange -or !$activeMinerByType -or ($activeMinerByType -and !$activeMiner) -or !$Config.SwitchingResistance.Enabled -or
				($Config.SwitchingResistance.Enabled -and ($run.Miner.Priority -ge [Priority]::Solo -or
					$activeMinerByType.CurrentTime.Elapsed.TotalMinutes -ge $Config.SwitchingResistance.Timeout -or
					($activeMiner.Profit -gt 0 -and ($run.Profit * 100 / $activeMiner.Profit - 100) -gt $Config.SwitchingResistance.Percent))))) {
				$miner = $run.Miner
				if (!$ActiveMiners.ContainsKey($miner.GetUniqueKey())) {
					$ActiveMiners.Add($miner.GetUniqueKey(), [MinerProcess]::new($miner, $Config))
				}
				# stop not choosen
				$activeMinersByType | Where-Object { $_.State -eq [eState]::Running -and ($miner.GetUniqueKey() -ne $_.Miner.GetUniqueKey() -or $FChange) } | ForEach-Object {
					$_.Stop($AllAlgos.RunAfter)
				}
				# run choosen
				$mi = $ActiveMiners[$miner.GetUniqueKey()]
				if ($mi.State -eq $null -or $mi.State -ne [eState]::Running) {
					if ($Statistics.GetValue($mi.Miner.GetFilename(), $mi.Miner.GetKey()) -eq 0 -or $FStart) {
						$mi.Benchmark($FStart, $AllAlgos.RunBefore)
					}
					else {
						$mi.Start($AllAlgos.RunBefore)
					}
					$FastLoop = $false
				}
				Remove-Variable mi, miner
			}
			elseif ($run -and $activeMinerByType -and $activeMiner -and $Config.SwitchingResistance.Enabled -and
				$run.Miner.GetUniqueKey() -ne $activeMinerByType.Miner.GetUniqueKey() -and
				!($activeMinerByType.CurrentTime.Elapsed.TotalMinutes -gt $Config.SwitchingResistance.Timeout -or
				($run.Profit * 100 / $activeMiner.Profit - 100) -gt $Config.SwitchingResistance.Percent)) {
				$run.SwitchingResistance = $true
			}
			elseif (!$run -and $lf) {
				# stop if lower floor
				$activeMinersByType | Where-Object { $_.State -eq [eState]::Running -and $_.Profit -lt $lf } | ForEach-Object {
					$_.Stop($AllAlgos.RunAfter)
				}
			}
			Remove-Variable lf, run, activeMiner, activeMinerByType, activeMinersByType, allMinersByType, type
		}

		if ($global:API.Running) {
			$global:API.MinersRunning = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | Select-Object (Get-FormatActiveMinersWeb) | ConvertTo-Html -Fragment
			$global:API.ActiveMiners = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | Select-Object (Get-FormatActiveMinersApi)
		}

		if (!$FastLoop -and ![string]::IsNullOrWhiteSpace($Config.ApiKey) -and
			(!$Summary.SendApiTime.IsRunning -or $Summary.SendApiTime.Elapsed.TotalSeconds -gt [Config]::ApiSendTimeout)) {
			Write-Host "Send data to online monitoring ..." -ForegroundColor Green
			$json = Get-JsonForMonitoring
			if (![string]::IsNullOrWhiteSpace($json)) {
				$str = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($json))
				$json = Get-Rest "http://api.mindminer.online/?type=setworker&apikey=$($Config.ApiKey)&worker=$($Config.WorkerName)&data=$str"
				if ($json -and $json.error) {
					Write-Host "Error send state to online monitoring: $($json.error)" -ForegroundColor Red
					Start-Sleep -Seconds ($Config.CheckTimeout)
				}
				$Summary.SendApiTime = [Diagnostics.Stopwatch]::StartNew()
				Remove-Variable str
			}
			Remove-Variable json
		}

		$Statistics.Write([Config]::StatsLocation)

		if (!$FastLoop) {
			$Summary.LoopTime.Reset()
			$Summary.LoopTime.Start()
		}

		$verbose = $Config.Verbose -as [eVerbose]

		Clear-Host
		Out-Header ($verbose -ne [eVerbose]::Minimal)

		if ($Config.DevicesStatus) {
			Out-DeviceInfo ($verbose -eq [eVerbose]::Minimal)
		}

		if ($verbose -eq [eVerbose]::Full) {
			Out-PoolInfo
		}
		
		[decimal] $mult = if ($verbose -eq [eVerbose]::Normal) { 0.70 } else { 0.85 }
		$bench = [hashtable]::new()
		$max = $AllMiners | Group-Object { $_.Miner.Type } | ForEach-Object {
			$bench[$_.Name] = ($_.Group | Where-Object { $_.Speed -eq 0 } | Select-Object @{ Name = "BenchmarkSeconds"; Expression = { $_.Miner.BenchmarkSeconds } } |
				Measure-Object BenchmarkSeconds -Sum).Sum
			$prft = ($_.Group | Select-Object -First 1).Profit
			$val = $_.Group | Where-Object { $_.Miner.Priority -gt [Priority]::None } | Select-Object -First 1
			if ($val) { $prft = $val.Profit }
			@{ $_.Name = $mult * $prft }
		}
		Remove-Variable mult
		$alg = [hashtable]::new()
		Out-Table ($AllMiners | Where-Object {
			$uniq =  $_.Miner.GetUniqueKey()
			$type = $_.Miner.Type
			if (!$alg[$type]) { $alg[$type] = [Collections.ArrayList]::new() }
			$_.Speed -eq 0 -or ($_.Profit -ge 0.00000001 -and ($verbose -eq [eVerbose]::Full -or
				($ActiveMiners.Values | Where-Object { $_.State -ne [eState]::Stopped -and $_.Miner.GetUniqueKey() -eq $uniq } | Select-Object -First 1) -or
					(($_.Profit -ge $max."$type" -or $_.Miner.Priority -gt [Priority]::Normal) -and
						$alg[$type] -notcontains "$($_.Miner.Algorithm)$($_.Miner.DualAlgorithm)")))
			$ivar = $alg[$type].Add("$($_.Miner.Algorithm)$($_.Miner.DualAlgorithm)")
			Remove-Variable ivar, type, uniq
		} |
		Format-Table (Get-FormatMiners) -GroupBy @{ Label = "Type"; Expression = {
			$rslt = "$($_.Miner.Type)"
			if ($bench[$_.Miner.Type] -gt 0) {
				$rslt += ", " + $(if ($global:HasConfirm -eq $true) { "Benchmarking" } else { "Need bench" }) + ": " +
					"$([SummaryInfo]::Elapsed([timespan]::FromSeconds($bench[$_.Miner.Type])))"
			}
			$rslt;
		}})
		Write-Host "^ Priority, + Running, - No Hash, ! Failed, % Switching Resistance, _ Low Profit, * Specified Coin, ** Solo|Party"
		Write-Host
		Remove-Variable alg, max, bench

		# display active miners
		if ($verbose -ne [eVerbose]::Minimal) {
			Out-Table ($ActiveMiners.Values | Where-Object { $verbose -eq [eVerbose]::Full -or $_.State -ne [eState]::Stopped } |
				Sort-Object { [int]($_.State -as [eState]), [SummaryInfo]::Elapsed($_.TotalTime.Elapsed) } |
					Format-Table (Get-FormatActiveMiners ($verbose -eq [eVerbose]::Full)) -GroupBy State -Wrap)
		}

		if ($Config.ShowBalance) {
			Out-PoolBalance ($verbose -eq [eVerbose]::Minimal)
		}
		Out-Footer
		if ($DownloadMiners -and ($DownloadMiners.Length -gt 0 -or $DownloadMiners -is [PSCustomObject])) {
			Write-Host "Download miner(s): $(($DownloadMiners | Select-Object Name -Unique | ForEach-Object { $_.Name }) -Join `", `") ... " -ForegroundColor Yellow
		}
		if ($global:HasConfirm) {
			Write-Host "Please observe while the benchmarks are running ..." -ForegroundColor Red
		}
		if ($PSVersionTable.PSVersion -lt [version]::new(5,1)) {
			Write-Host "Please update PowerShell to version 5.1 (https://www.microsoft.com/en-us/download/details.aspx?id=54616)" -ForegroundColor Yellow
		}

		Remove-Variable verbose
	}

	$switching = $Config.Switching -as [eSwitching]

	do {
		$FastLoop = $false

		$start = [Diagnostics.Stopwatch]::new()
		$start.Start()
		do {
			Start-Sleep -Milliseconds ([Config]::SmallTimeout)
			while ([Console]::KeyAvailable -eq $true) {
				[ConsoleKeyInfo] $key = [Console]::ReadKey($true)
				if (($key.Modifiers -match [ConsoleModifiers]::Alt -or $key.Modifiers -match [ConsoleModifiers]::Control) -and $key.Key -eq [ConsoleKey]::S) {
					$items = [enum]::GetValues([eSwitching])
					$index = [array]::IndexOf($items, $Config.Switching -as [eSwitching]) + 1
					$Config.Switching = if ($items.Length -eq $index) { $items[0] } else { $items[$index] }
					Remove-Variable index, items
					Write-Host "Switching mode changed to $($Config.Switching)." -ForegroundColor Green
					Start-Sleep -Milliseconds ([Config]::SmallTimeout * 2)
					$FastLoop = $true
				}
				elseif ($key.Key -eq [ConsoleKey]::V) {
					$items = [enum]::GetValues([eVerbose])
					$index = [array]::IndexOf($items, $Config.Verbose -as [eVerbose]) + 1
					$Config.Verbose = if ($items.Length -eq $index) { $items[0] } else { $items[$index] }
					Remove-Variable index, items
					Write-Host "Verbose level changed to $($Config.Verbose)." -ForegroundColor Green
					Start-Sleep -Milliseconds ([Config]::SmallTimeout * 2)
					$FastLoop = $true
				}
				elseif (($key.Modifiers -match [ConsoleModifiers]::Alt -or $key.Modifiers -match [ConsoleModifiers]::Control) -and
					($key.Key -eq [ConsoleKey]::E -or $key.Key -eq [ConsoleKey]::Q -or $key.Key -eq [ConsoleKey]::X)) {
					$exit = $true
					# for mrr to disable all rigs
					[Config]::ActiveTypes = @()
				}
				elseif (($key.Modifiers -match [ConsoleModifiers]::Alt -or $key.Modifiers -match [ConsoleModifiers]::Control) -and $key.Key -eq [ConsoleKey]::R) {
					New-Item ([IO.Path]::Combine([Config]::BinLocation, ".restart")) -ItemType Directory -Force | Out-Null
					$exit = $true
				}
				elseif ($Config.ShowBalance -and $key.Key -eq [ConsoleKey]::R) {
					$Config.ShowExchangeRate = !$Config.ShowExchangeRate;
					$FastLoop = $true
				}
				elseif ($key.Key -eq [ConsoleKey]::C -and !$global:HasConfirm) {
					Clear-OldMiners ($ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | ForEach-Object { $_.Miner.Name })
				}
				elseif ($key.Key -eq [ConsoleKey]::F -and !$global:HasConfirm) {
					if (Clear-FailedMiners ($ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Failed })) {
						$FastLoop = $true
					}
				}
				elseif ($key.Key -eq [ConsoleKey]::T -and !$global:HasConfirm -and [Config]::ActiveTypesInitial.Length -gt 1) {
					[Config]::ActiveTypes = Select-ActiveTypes ([Config]::ActiveTypesInitial)
					[Config]::ActiveTypesInitial | Where-Object { [Config]::ActiveTypes -notcontains $_ } | ForEach-Object {
						$type = $_
						$ActiveMiners.Values | Where-Object { $_.Miner.Type -eq $type -and $_.State -eq [eState]::Running } | ForEach-Object {
							$_.Stop($AllAlgos.RunAfter)
							$KnownAlgos[$type].Clear()
						}
						Remove-Variable type
					}
					# for normal loop
					$switching = $null
					$FastLoop = $true
				}
				elseif ($key.Key -eq [ConsoleKey]::Y -and $global:HasConfirm -eq $false -and $global:NeedConfirm -eq $true) {
					Write-Host "Thanks. " -ForegroundColor Green -NoNewline
					Write-Host "Please observe while the benchmarks are running ..." -ForegroundColor Red
					Start-Sleep -Milliseconds ([Config]::SmallTimeout * 2)
					$global:HasConfirm = $true
					$FastLoop = $true
				}
				elseif ($key.Key -eq [ConsoleKey]::P -and $global:HasConfirm -eq $false -and $global:NeedConfirm -eq $false -and [Config]::UseApiProxy -eq $false) {
					$global:AskPools = $true
					$FastLoop = $true
				}

				Remove-Variable key
			}
		} while ($start.Elapsed.TotalSeconds -lt $Config.CheckTimeout -and !$exit -and !$FastLoop)
		Remove-Variable start

		# if needed - exit
		if ($exit -eq $true) {
			Write-Host "Exiting ..." -ForegroundColor Green
			if ($global:API.Running) {
				Write-Host "Stoping API server ..." -ForegroundColor Green
				Stop-ApiServer
			}
			$ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | ForEach-Object {
				$_.Stop($AllAlgos.RunAfter)
			}
			# stop mrr
			if (![string]::IsNullOrWhiteSpace($global:MRRFile) -and !([Config]::ActiveTypes)) {
				Invoke-Expression $global:MRRFile | Out-Null
			}
			exit
		}

		if (!$FastLoop) {
			# read speed while run main loop timeout
			if ($ActiveMiners.Values -and $ActiveMiners.Values.Length -gt 0) {
				Get-Speed $ActiveMiners.Values
			}
			# check miners work propertly
			$ActiveMiners.Values | Where-Object { $_.State -ne [eState]::Stopped } | ForEach-Object {
				$prevState = $_.State
				if ($_.Check($AllAlgos.RunAfter) -eq [eState]::Failed -and $prevState -ne [eState]::Failed) {
					# miner failed - run next
					if ($_.Action -eq [eAction]::Benchmark) {
						$speed = $Statistics.SetValue($_.Miner.GetFilename(), $_.Miner.GetKey(), -1)
						Remove-Variable speed
					}
					$FastLoop = $true
				}
				# benchmark time reached - exit from loop
				elseif ($_.Action -eq [eAction]::Benchmark -and $_.State -ne [eState]::Failed) {
					$speed = $_.GetSpeed($false)
					if (($_.CurrentTime.Elapsed.TotalSeconds -ge $_.Miner.BenchmarkSeconds -and $speed -gt 0) -or
						($_.CurrentTime.Elapsed.TotalSeconds -ge ($_.Miner.BenchmarkSeconds * 2) -and $speed -eq 0)) {
						$FastLoop = $true
					}
					Remove-Variable speed
				}
				Remove-Variable prevState
			}
		}
		if ($global:API.Running) {
			$global:API.MinersRunning = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | Select-Object (Get-FormatActiveMinersWeb) | ConvertTo-Html -Fragment
			$global:API.ActiveMiners = $ActiveMiners.Values | Where-Object { $_.State -eq [eState]::Running } | Select-Object (Get-FormatActiveMinersApi)
			$global:API.Info = $Summary | Select-Object ($Summary.Columns()) | ConvertTo-Html -Fragment
			$global:API.Status = $Summary | Select-Object ($Summary.ColumnsApi())
		}
	} while ($Config.LoopTimeout -gt $Summary.LoopTime.Elapsed.TotalSeconds -and !$FastLoop)

	# if timeout reached or askpools or bench or change switching mode - normal loop
	if ($Config.LoopTimeout -le $Summary.LoopTime.Elapsed.TotalSeconds -or $switching -ne $Config.Switching -or
		$global:AskPools -eq $true -or ($global:HasConfirm -eq $true -and $global:NeedConfirm -eq $true)) {
		$FastLoop = $false
	}

	if (!$FastLoop) {
		if ($Summary.RateTime.IsRunning -eq $false -or $Summary.RateTime.Elapsed.TotalSeconds -ge [Config]::RateTimeout.TotalSeconds) {
			Clear-OldMinerStats $AllMiners $Statistics "180 days"
		}
		$global:NeedConfirm = $false
		Remove-Variable AllPools, AllMiners
		[GC]::Collect()
		$Summary.Loop++
	}
}
