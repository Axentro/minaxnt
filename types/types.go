package types

const (
	CoreVersion                    int = 1
	TypeMinerHandshake             int = 0x0001
	TypeMinerHandshakeAccepted     int = 0x0002
	TypeMinerHandshakeRejected     int = 0x0003
	TypeMinerFoundNonce            int = 0x0004
	TypeMinerBlockUpdate           int = 0x0005
	TypeMinerBlockDifficultyAdjust int = 0x0006
	TypeMinerBlockInvalid          int = 0x0007
)

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
	Timestamp      int64         `json:"timestamp"`
	Difficulty     int32         `json:"difficulty"`
	Kind           string        `json:"kind"`
	Address        string        `json:"address"`
}

type MinerNonce struct {
	Mid        string `json:"mid"`
	Value      string `json:"value"`
	Timestamp  int64  `json:"timestamp"`
	Address    string `json:"address"`
	NodeID     string `json:"node_id"`
	Difficulty int32  `json:"difficulty"`
}

func NewMinerNonce() MinerNonce {
	return MinerNonce{
		Mid:        "0",
		Value:      "",
		Timestamp:  0,
		Address:    "0",
		NodeID:     "0",
		Difficulty: 0,
	}
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

type PeerResponseWithReason struct {
	PeerResponse
	Reason string `json:"reason"`
}

func (prwr PeerResponseWithReason) ToPeerResponse() PeerResponse {
	return PeerResponse{
		Version:          prwr.Version,
		Block:            prwr.Block,
		MiningDifficulty: prwr.MiningDifficulty,
	}
}

type PeerRejectedResponse struct {
	Reason string `json:"reason"`
}
