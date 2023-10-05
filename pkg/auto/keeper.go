package auto

import "os"

func getKeeperTagName() string {
	k, ok := os.LookupEnv("KEEPER_TAG_NAME")
	if !ok {
		k = "com.wulung-triyanto.keeper"
	}

	return k
}
