package util

import (
	"fmt"
	"minaxnt/miner"
	"minaxnt/types"
)

func Welcome(client *miner.Client) {
	str := `
  __       __ __           ______  __    __ __    __ ________
  |  \     /  \  \         /      \|  \  |  \  \  |  \        \
  | ▓▓\   /  ▓▓\▓▓_______ |  ▓▓▓▓▓▓\ ▓▓  | ▓▓ ▓▓\ | ▓▓\▓▓▓▓▓▓▓▓
  | ▓▓▓\ /  ▓▓▓  \       \| ▓▓__| ▓▓\▓▓\/  ▓▓ ▓▓▓\| ▓▓  | ▓▓
  | ▓▓▓▓\  ▓▓▓▓ ▓▓ ▓▓▓▓▓▓▓\ ▓▓    ▓▓ >▓▓  ▓▓| ▓▓▓▓\ ▓▓  | ▓▓
  | ▓▓\▓▓ ▓▓ ▓▓ ▓▓ ▓▓  | ▓▓ ▓▓▓▓▓▓▓▓/  ▓▓▓▓\| ▓▓\▓▓ ▓▓  | ▓▓
  | ▓▓ \▓▓▓| ▓▓ ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓  ▓▓ \▓▓\ ▓▓ \▓▓▓▓  | ▓▓
  | ▓▓  \▓ | ▓▓ ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓ ▓▓  \▓▓▓  | ▓▓
   \▓▓      \▓▓\▓▓\▓▓   \▓▓\▓▓   \▓▓\▓▓   \▓▓\▓▓   \▓▓   \▓▓
 `
	fmt.Print(str)
	fmt.Println()
	fmt.Println("=>          miner name:", client.ClientName)
	fmt.Println("=>            miner id:", client.MinerID)
	fmt.Println("=> client core version:", types.CoreVersion)
	fmt.Println("=>   connected to node:", client.NodeURL)
	fmt.Println("=>       miner address:", client.Address)
	fmt.Println("=>       miner process:", client.Process)
	fmt.Println("=>           cpu model:", client.CPUModel)
	fmt.Println("=>           cpu cores:", client.CPUCores)
	fmt.Println("=>          cpu caches:", client.CPUCaches)
	fmt.Println("=>        cpu features:", client.CPUFeatures)
}
