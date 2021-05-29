package types

const (
	CoreVersion                    string = "3.0.0"
	TypeMinerHandshake             int    = 0x0001
	TypeMinerHandshakeAccepted     int    = 0x0002
	TypeMinerHandshakeRejected     int    = 0x0003
	TypeMinerFoundNonce            int    = 0x0004
	TypeMinerBlockUpdate           int    = 0x0005
	TypeMinerBlockDifficultyAdjust int    = 0x0006
	TypeMinerBlockInvalid          int    = 0x0007
	TypeMinerExceedRate            int    = 0x0008
	TypeMinerInsufficientDuration  int    = 0x0009
)

type Sender struct {
	Address       string `json:"address"`
	PublicKey     string `json:"public_key"`
	Amount        int64  `json:"amount"`
	Fee           int64  `json:"fee"`
	Signature     string `json:"signature"`
	AssetId       string `json:"asset_id,omitempty"`
	AssetQuantity int32  `json:"asset_quantity,omitempty"`
}

type Recipient struct {
	Address       string `json:"address"`
	Amount        int64  `json:"amount"`
	AssetId       string `json:"asset_id,omitempty"`
	AssetQuantity int32  `json:"asset_quantity,omitempty"`
}

type Asset struct {
	AssetId       string `json:"asset_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	MediaLocation string `json:"media_location"`
	MediaHash     string `json:"media_hash"`
	Quantity      int32  `json:"quantity"`
	Terms         string `json:"terms"`
	Locked        string `json:"locked"`
	Version       int32  `json:"version"`
	Timestamp     int64  `json:"timestamp"`
}

type Module struct {
	ModuleId  string `json:"module_id"`
	Timestamp int64  `json:"timestamp"`
}

type Input struct {
	InputId   string `json:"input_id"`
	Timestamp int64  `json:"timestamp"`
}

type Output struct {
	OutputId  string `json:"output_id"`
	Timestamp int64  `json:"timestamp"`
}

type Transaction struct {
	Id         string      `json:"id"`
	Action     string      `json:"action"`
	Message    string      `json:"message"`
	Token      string      `json:"token"`
	PrevHash   string      `json:"prev_hash"`
	Timestamp  int64       `json:"timestamp"`
	Scaled     int32       `json:"scaled"`
	Kind       string      `json:"kind"`
	Version    string      `json:"version"`
	Senders    []Sender    `json:"senders"`
	Recipients []Recipient `json:"recipients"`
	Assets     []Asset     `json:"assets"`
	Modules    []Module    `json:"modules"`
	Inputs     []Input     `json:"inputs"`
	Outputs    []Output    `json:"outputs"`
	Linked     string      `json:"linked"`
}

type MinerBlock struct {
	Index          int64         `json:"index"`
	Nonce          string        `json:"nonce"`
	PrevHash       string        `json:"prev_hash"`
	MerkleTreeRoot string        `json:"merkle_tree_root"`
	Difficulty     int32         `json:"difficulty"`
	Address        string        `json:"address"`
	PublicKey      string        `json:"public_key"`
	Signature      string        `json:"signature"`
	Hash           string        `json:"hash"`
	Version        string        `json:"version"`
	HashVersion    string        `json:"hash_version"`
	Checkpoint     string        `json:"checkpoint"`
	MiningVersion  string        `json:"mining_version"`
	Transactions   []Transaction `json:"transactions"`
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
	Version          string     `json:"version"`
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
