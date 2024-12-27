---

date: 2024/12/27

---

# Sirang - Tunneling TCP over QUIC
Sirang is both ways(forward and reverse) TCP tunnel transported over QUIC.

**::** It is written in Rust and primarily built on the [tokio](https://tokio.rs/) runtime and AWS's implementation of QUIC, [s2n-quic](https://github.com/aws/s2n-quic). 

**::** Sirang is inspired by [reverstd](https://github.com/flipt-io/reverst), an HTTP reverse tunnel written in Go, which similarly uses QUIC as the transport layer.

**::** It's simple and only incorporates the usage of a single binary for both forward and reverse tunneling. 

**::** It has minimal overhead and can be easily deployed to a low end VPS to tunnel TCP services. 

Even though it is currently experimental, it is fairly functional and should be reliable if deployed. Further plans for the project include the development of an HTTP extension that would enable reverse tunneling HTTP 1,2 and 3 by dispensing and parsing custom subdomains over a single remote instance, like [ngrok](https://ngrok.com/).

Check it out: https://github.com/icelain/sirang
