module github.com/james-milligan/flagd-proxy-profiling

go 1.20

require (
	buf.build/gen/go/open-feature/flagd/bufbuild/connect-go v1.6.0-20230317150644-afd1cc2ef580.1
	buf.build/gen/go/open-feature/flagd/protocolbuffers/go v1.30.0-20230317150644-afd1cc2ef580.1
	github.com/bufbuild/connect-go v1.6.0
)

require google.golang.org/protobuf v1.30.0 // indirect
