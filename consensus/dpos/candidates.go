package dpos

type CandidatesInfo struct {
	Address string
	PeerId  string
}

// initialCandidates the first super node of the block generation cycle.
// The first half is the address of the block, the second half is the id of the block node
var initialCandidates = []CandidatesInfo{
	{
		Address: "UWDR5oWfEjnGcEnQJrknuNW3LxZeWSpZUPJ5",
		PeerId:  "16Uiu2HAm8dGc2gAuQG9WAdPNeXRDG1GU8wQtv4f2jdGEGkrX9Ln1",
	},
	{
		Address: "UWDP1EbJ1mpT4sDD1p6TWhvNXBjxRiqLjW7F",
		PeerId:  "16Uiu2HAm5gMSBTPc1PjcsJJpSNvXWmA3KaMkfLsjxs78BDA4FLDM",
	},
	{
		Address: "UWDGvCpqgfGdjTRYBV2F8WCek6oBWTLiUTQH",
		PeerId:  "16Uiu2HAmTfA1REe8wkYGNn7NzENwfyR9wVbLFLCDe6x3J1aXhTG2",
	},
	{
		Address: "UWDcSaX5hUhutRDC4EQT65DLeRJxBJWBWppX",
		PeerId:  "16Uiu2HAmK2PNwXpTM91ftN9ZF92CNLKf5m7XDApBaBa69cwPDpMv",
	},
	{
		Address: "UWDHR3U2FQDNYdQpgUrHw3LkrMiC23wetsb6",
		PeerId:  "16Uiu2HAmKCs2So8WRqZj5aWQKHAMj6DhLd1eQQBvF8MdjgbajUHL",
	},
	{
		Address: "UWDamE7HaJNvRrRDx3Rnmj9WKqHnDMFzQ6P3",
		PeerId:  "16Uiu2HAkvRSRt5VEeN8vXKg1sxDSeXwJfQzPqYPPneoaTNwYxQuM",
	},
	{
		Address: "UWDRKs3eg7deVcRo5pJVCsRAVQ1umXNM1DiD",
		PeerId:  "16Uiu2HAmABaSg6ZmBpKXoSCV5V89fNS4N3s91wRAbneq7DJPN81y",
	},
	{
		Address: "UWDZqKCCQLV2yTZ7crW6kZn7UFLPCpHtFdGi",
		PeerId:  "16Uiu2HAm9w9gZsnS7ZKGqdLfNgXuyMrfhBgSKHWePVfhHGVzDDxH",
	},
	{
		Address: "UWDKvSoMxcD7MbT1KZ3kX5xjnL4tWK7XXGCd",
		PeerId:  "16Uiu2HAmULDuqSBP9mueYjHCknUuex3yGVQM4QXgBNTHf8qiNFD6",
	},
	{
		Address: "UWDN2rVzEgJRVGyZFUQstL2y2JpSkGVz1kbY",
		PeerId:  "16Uiu2HAkxKA74FNnwY5hAJGW64GyDgfWFeCtkkuFJwXbSLeVKK1t",
	},
	{
		Address: "UWDToaCgPNCZ168ZnvUg38bRMm5U7DjwRVuA",
		PeerId:  "16Uiu2HAmRsZi1iWAx1efSv7o5dvzeYoRXhQusNtjzbhmmuqUo99q",
	},
}
