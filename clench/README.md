# Clench Protocol v0.0.1

Clench just hanle the content of data, which is not in crypto and is stateful. 

Request

| Len(RAND) | Protocol | PORT | Len(DST) | DST | RAND |
|--|--|--|--|--|--|--|--|
| 1 byte | 1 byte (TCP, UDP etc) | 2 bytes |  1 byte | N bytes (Max 2<<7=256bytes) | 0-1<<5 bytes |

Protocol

- 0x00 TCP + IP
- 0x01 TCP + FQDN

Response

| Len(RAND) | StatusMask | RAND |
|--| -- | -- | -- | -- | -- |
| 1 byte | 2 bytes | N |
