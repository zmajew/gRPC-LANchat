# gRPC-LANchat

Command line (shell or cmd) LAN group chat. It finds active nodes on the local network and establish p2p connection between. 

Application discovers computers with activated chat app on the LAN. It uses the host computer arp table to find active IPs on the local network. Then it sends handshake gRPC messages to the IPs to check if there are other active chat nodes on LAN. 

Chat uses port "4041" and it is hardcoded since the port should be unique for all nodes. 
It should work on Linux, Windows and Mac OS.

If there is a problem with establishing chat check if your os firewall or third party firewall is blocking the port "4041".

Please consider that this is simple demonstration GO gRPC script. 

## Usage
Install or build it from shell or cmd:

```sh
go install 
```
or:
```sh
go build 
```
Run the chat application on Linux or Mac from shell:
```sh
gRPC-LANchat
```
or if you build it:
```sh
./gRPC-LANchat
```
on Windows:
```sh
start gRPC-LANchat.exe
```

Run the chat app with message sound volume reduction:
```sh
gRPC-LANchat -volume <VOLUME> //0-100
```
There is Chat.exe and Chat binaries in build folder.
Start the node on the diferent computers on the LAN and start chat in the cli.
