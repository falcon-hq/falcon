# Falcon

Falcon is a little proxy tool pair which helps transfer anything in simple and secure, which can be seen as an alternative of shadowsocks or v2ray.

## Usage

Upload or download the `server` binary to remote server, then

```bash
./serer -l [Listen Port] -k [Key for Encrypt]

# example
./serer -l 18000 -k 123456
```

Then on your PC or any local device

```bash
./client -l [Local Socks5 Service Port] -r [Remote Server IP or Address] -k [The Key above]

# example
./client -l 10008 -r 11.11.11.11:18000 -k 123456
```
