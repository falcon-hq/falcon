# Falcon

Falcon is a proxy tool which helps transfer anything in simple and secure, which can be seen as an alternative of shadowsocks or v2ray.

WARN: Falcon does not represent for "Americanization", and no politic thing.

## Build

Linux

```bash
make
./bin/falcon --help
```

Windows

```powershell
./Make.ps1
./bin/falcon.exe --help
```


## Usage

Upload or download the `falcon` binary to remote server, then

```bash
./falcon remote -l [Listen Port] -k [Key for Encrypt]

# example
./falcon remote -l 18000 -k 123456
```

Then on your PC or any local device

```bash
./falcon local -l [Local Socks5 Service Port] -r [Remote Server IP or Address] -k [The Key above]

# example
./falcon local -l 10008 -r 11.11.11.11:18000 -k 123456
```
