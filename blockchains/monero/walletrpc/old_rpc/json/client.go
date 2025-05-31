// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Copyright 2023 Irem Kuyucu <siren@kernal.eu>. All rights reserved.
// Copyright 2024 Decred developers <TODO:>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"encoding/json"
	"io"
)

type EmptyResponse struct {
}

// EncodeClientRequest encodes parameters for a Daemon RPC client request. If
// args is nil an empty json params is created.
func EncodeClientRequest(args interface{}) ([]byte, error) {
	if args == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(args)
}

// DecodeClientResponse decodes the response body of a client request into
// the interface reply.
func DecodeClientResponse(r io.Reader, reply interface{}) error {
	var j json.RawMessage
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return err
	}
	// fmt.Printf("raw json returned:\n%s\n", string(j))
	return json.Unmarshal(j, reply)
}
