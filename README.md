<h3 align="center">Codec for a Typed Map</h3>
<p align="center">Provides round-trip serialization of typed Go maps.<p>
<p align="center"><a href="https://github.com/bep/typedmapcodec/actions"><img src="https://action-badges.now.sh/bep/typedmapcodec" /></a></p>



## Performance

The implementation is easy to reason aobut, but It's not particulary fast and probably not suited for _big data_. As imple benchmark with a roundtrip marshal/unmarshal is included:

```bash
BenchmarkCodec/JSON_regular-4         	   50000	     27523 ns/op	    6742 B/op	     171 allocs/op
BenchmarkCodec/JSON_typed-4           	   20000	     66644 ns/op	   16234 B/op	     411 allocs/op
```