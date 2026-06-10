//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package config

import _ "embed"

//go:embed dcv-autosession.conf.default
var defaultConfig []byte // Embed the file content as a byte slice
