package util

import "fmt"

func Welcome(node string, address string, minerId string, minerProcess int, version string) {
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
	fmt.Println("=>           version:", version)
	fmt.Println("=>        miner name:", minerId)
	fmt.Println("=> connected to node:", node)
	fmt.Println("=>     miner address:", address)
	fmt.Println("=>     miner process:", minerProcess)
}
