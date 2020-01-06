package discovery

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
)

func pbStringValue(v string) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: v,
		},
	}
}

