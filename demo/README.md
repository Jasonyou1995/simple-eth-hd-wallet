# Sample Outputs

[Notes: make sure that you already installed go and the packages required in your local `$GOPATH`]

Command to run: `go run <demo-name>.go`

### demo1-keygen.go

```
Account address: 0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947
Private key in hex: 63e21d10fd50155dbba0e7d3f7431a400b84b4c2ac1ee38872f82448fe3ecfb9
Public key in hex: 6005c86a6718f66221713a77073c41291cc3abbfcd03aa4955e9b2b50dbf7f9b6672dad0d46ade61e382f79888a73ea7899d9419becf1d6c9ec2087c1188fa18
```

### demo2-seed2addr.go

```
Random Seed:
[ 203 219 159 10 245 15 62 50 246 15 36 134 164 143 238 176 109 219 116 36 40 255 26 198 124 3 22 192 254 75 42 227 232 85 91 237 112 220 249 210 28 221 143 8 199 45 145 241 206 62 55 146 58 90 188 159 207 82 144 213 239 141 39 49 ]

Path: m/44'/60'/0'/0/0
Address: 0x6547D3a563aE70dA57E6eE7edCDDf20E8a97f8ef

Path: m/44'/60'/0'/0/9
Address: 0xC39aA32b1D911F144AE55B3F0412a43480127e52
```

### demo3-sign.go

```
(*types.Transaction)(0xc000106900)({
 data: (types.txdata) {
  AccountNonce: (uint64) 0,
  Price: (*big.Int)(0xc00000c500)(21000000000),
  GasLimit: (uint64) 21000,
  Recipient: (*common.Address)(0xc000018660)((len=20 cap=20) 0x0000000000000000000000000000000000000000),
  Amount: (*big.Int)(0xc00000c4e0)(1000000000000000000),
  Payload: ([]uint8) <nil>,
  V: (*big.Int)(0xc00000c7e0)(27),
  R: (*big.Int)(0xc00000c7a0)(34405166580762396054881948095668280144114812929766777744840143175291345694076),
  S: (*big.Int)(0xc00000c7c0)(35876182893985365337785111608140175101791577417143941897988905965773505549184),
  Hash: (*common.Hash)(<nil>)
 },
 hash: (atomic.Value) {
  v: (interface {}) <nil>
 },
 size: (atomic.Value) {
  v: (interface {}) <nil>
 },
 from: (atomic.Value) {
  v: (types.sigCache) {
   signer: (types.HomesteadSigner) {
    FrontierSigner: (types.FrontierSigner) {
    }
   },
   from: (common.Address) (len=20 cap=20) 0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947
  }
 }
})
```

### demo4-addr-index.go

```
Successfully generated address1 from path1:
	Path:		 m/44'/60'/0'/0/0 
	Address:	 0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947
Successfully generated address2 from path2:
	Path2:		 m/44'/60'/0'/0/9 
	Address2:	 0x2d69B45301b9B3E01c4797C7a48BBc7e7F9b355b
```