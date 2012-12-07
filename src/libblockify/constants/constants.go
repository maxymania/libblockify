// This package contains the Blockify Constants
package constants

// Number of Bytes in a (Binary) Kilo Byte
const KiB = 1024

// Number of KiB in a Block
const BlockSizeInKiB = 128

// Nunber of Bytes in a Block
const BlockSize = BlockSizeInKiB*KiB

// Size of a Hash
const HashSize = 64

// Maximum Number of Hashes Per Block
const MaxHashes = 2048-1
