package main

import "google.golang.org/grpc/resolver"

const Scheme = "grpc"

type serviceResolverBuilder struct{}

func (*serviceResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &serviceResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			"product.core": {
				"localhost:30001",
				"localhost:30002",
			},
			"order.core": {
				"localhost:30003",
				"localhost:30004",
			},
		},
	}
	r.start()
	return r, nil
}

func (*serviceResolverBuilder) Scheme() string { return Scheme }

type serviceResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *serviceResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*serviceResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*serviceResolver) Close()                                  {}
