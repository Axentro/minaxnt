package types

const (
	CORE_VERSION                  int = 1
	TYPE_MINER_HANDSHAKE          int = 0x0001
	TYPE_MINER_HANDSHAKE_ACCEPTED int = 0x0002
	TYPE_MINER_HANDSHAKE_REJECTED int = 0x0003
	TYPE_MINER_FOUND_NONCE        int = 0x0004
	TYPE_MINER_BLOCK_UPDATE       int = 0x0005
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

type Sender struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Amount    int64  `json:"amount"`
	Fee       int64  `json:"fee"`
	Signature string `json:"signature"`
}

type Recipient struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type Transaction struct {
	Id         string      `json:"id"`
	Action     string      `json:"action"`
	Senders    []Sender    `json:"senders"`
	Recipients []Recipient `json:"recipients"`
	Message    string      `json:"message"`
	Token      string      `json:"token"`
	PrevHash   string      `json:"prev_hash"`
	Timestamp  int64       `json:"timestamp"`
	Scaled     int32       `json:"scaled"`
	Kind       string      `json:"kind"`
	Version    string      `json:"version"`
}

type MinerBlock struct {
	Index          int64         `json:"index"`
	Transactions   []Transaction `json:"transactions"`
	Nonce          string        `json:"nonce"`
	PrevHash       string        `json:"prev_hash"`
	MerkleTreeRoot string        `json:"merkle_tree_root"`
	Difficulty     int32         `json:"difficulty"`
	Address        string        `json:"address"`
}

type MinerNonce struct {
	Mid       string `json:"mid"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
	Address   string `json:"address"`
	NodeId    string `json:"node_id"`
}

func NewMinerNonce() MinerNonce {
	mn := MinerNonce{}
	mn.Mid = "0"
	mn.Value = ""
	mn.Timestamp = 0
	mn.Address = "0"
	mn.NodeId = "0"
	return mn
}

type MinerNonceContent struct {
	Nonce MinerNonce `json:"nonce"`
}

type MessageResponse struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

type PeerResponse struct {
	Version          int32      `json:"version"`
	Block            MinerBlock `json:"block"`
	MiningDifficulty int32      `json:"difficulty"`
}

type PeerRejectedResponse struct {
	Reason string `json:"reason"`
}
