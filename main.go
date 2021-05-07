// Copyright (c) 2021 MacEwan University. All rights reserved.
//
// This source code is licensed under the MIT-style license found in
// the LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"

	"github.com/macewan-cs/lti/datastore/nonpersistent"
)

func main() {
	np := nonpersistent.New()

	fmt.Printf("np: %v", np)
}
