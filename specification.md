# Xatum Protocol Definition

All the Xatum connections are secured by TLS. Self-signed certificates are accepted, but the miner
can show the certificate fingerprint to the user or verify it via configuration for protection
against MITM attacks.

Packets are separated by `\n`.
A packet is formed by a "method string", followed by tilde ("~") and then the JSON message.

## Connection operation

1. Client connects via TCP with TLS.

2. Client sends handshake packet with JSON

3. Server validates handshake, and then starts responding with jobs

## Definitions

**C2S**: Client To Server
**S2C**: Server To Client


## Standardized Algorithm Names

`xel/0`: XelisHash, first version

## Xelis structure: the BlockMiner
BlockMiner represents a block to mine, has a fixed length of 112 bytes.
The miner must insert the timestamp and increase the nonce, while the workhash,
extranonce and publickey are given by the pool in the job and miner must not change them.

workhash:	32	bytes 0-31
timestamp:	8	bytes 32-39
nonce:		8	bytes 40-47
extranonce:	32	bytes 48-79
publickey:	32	bytes 80-111

## Packets

### C2S Handshake packet

```json
shake~{
	"addr": "xel:myAddress",	// Address: wallet address
	"work": "x",				// Worker: worker name, by default "x"
	"agent": "xmrig/v0.1.0",	// Useragent: the mining software
	"algos": [					// Algos: list of supported algorithms
		"xel/0", "rx/0"
	]
}
```

### S2C Job packet
```json
job~{
	"algo": "xel/0",	// algorithm of the job
	"diff": 15021,		// difficulty of the job
	
	"blob": "base64",	// xelis blob, which embeds work hash, extra nonce and public key (96 bytes) encoded as base64 string
}
```

### C2S Submit packet
```json
submit~{
	"data": "base64",	// the 112-bytes BlockMiner encoded as base64 string
	"hash": "hex",		// the 32-bytes PoW hash of BlockMiner encoded as hex string
}
```

### S2C success packet
This packet is sent to miner when a share is submit, to tell if operation is successful or not.
```json
success~{
	"msg": "ok" // "ok" if share is good, otherwise msg contains the error message
}
```

### S2C print packet
This packet makes the miner print some information.
Usually used before kicking the client, to send the error message. Can also send warnings.

```json
print~{
	"msg": "example message!",	// this message will be printed in miner console
	"lvl": 1,					// log level, 0: verbose, 1: info, 2: warn, 3: error
}
```

### S2C ping packet
Sent by server, can be used to measure latency and check if connection is alive.
```
ping~{}
```

### C2S pong packet
A reply to the ping packet. If the client does not send the packet, server may interrupt the
connection.
```
pong~{}
```

## Drafts
These features aren't yet implemented into software, but are open to discussion and may, or may not,
be added in a future version.

### [DRAFT] S2C redirect packet
Used to redirect to another server address. Useful for pool-side load balancing.
It works in a similar way to HTTP's Temporary Redirect.

```json
redirect~{
	"to": "IP:PORT" // the client will disconnect from the pool and connect to the new address
}
```