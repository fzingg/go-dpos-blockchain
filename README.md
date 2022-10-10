# Implement a simple blockchain use dpos algorithm
Sample code taken from https://github.com/csunny/dpos for learning and experimenting

## Architecture Design
- Create a P2P Conn-pool
- BlockChain Generate
- Node Manage And Vote
- Pick Node
- Write Block On Blockchain

## Build 
go build -o build/dpos  main/dpos.go

## RUN 
```
git clone git@github.com:fzingg/dpos.git

cd dpos    
go build main/dpos.go
```

connect multi peer 
```
./dpos new --port 3000 --secio
```
## Vote
```
./dpos vote -name QmaxEdbKW4x9mP2vX15zL9fyEsp9b9yV48zwtdrpYddfxe -v 30
```


