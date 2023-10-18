package cctransfer

import (
	"path"
)

// Data store pathes.
const (
	prefix                   = "/f/"                              // f - foundation
	pathBase                 = prefix + "b/"                      // b - base namespace
	pathCrossChannelTransfer = pathBase + "cct/"                  // cct - cross channel transfer
	pathTransferFrom         = pathCrossChannelTransfer + "from/" // f - From + ID
	pathTransferTo           = pathCrossChannelTransfer + "to/"   // t - To + ID
)

// Prefix returns prefix key path.
func Prefix() string {
	return prefix
}

// Base returns the last element of path.
// Trailing slashes are removed before extracting the last element.
func Base(fullPath string) string {
	return path.Base(fullPath)
}

// CCFromTransfers returns path to store key.
func CCFromTransfers() string {
	return pathTransferFrom
}

// CCFromTransfer returns path to store key.
func CCFromTransfer(id string) string {
	return path.Join(CCFromTransfers(), id)
}

// CCToTransfers returns path to store key.
func CCToTransfers() string {
	return pathTransferTo
}

// CCToTransfer returns path to store key.
func CCToTransfer(id string) string {
	return path.Join(CCToTransfers(), id)
}
