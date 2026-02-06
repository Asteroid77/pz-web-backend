package legacy

import (
	"fmt"
	"net/http"
	"sync"

	"pz-web-backend/internal/infra/pzpaths"
	"pz-web-backend/internal/mods"
)

var (
	workshopMu     sync.Mutex
	workshopClient *mods.WorkshopClient
)

// Deprecated: FetchWorkshopInfoLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/modsapp。
func FetchWorkshopInfoLegacy(workshopID string, devMode bool) (mods.ModInfo, error) {
	c, err := getWorkshopClient(devMode)
	if err != nil {
		return mods.ModInfo{}, err
	}
	return c.FetchWorkshopInfo(workshopID)
}

func getWorkshopClient(devMode bool) (*mods.WorkshopClient, error) {
	workshopMu.Lock()
	defer workshopMu.Unlock()

	if workshopClient != nil {
		return workshopClient, nil
	}

	path := pzpaths.WorkshopCachePath(devMode)

	client, err := mods.NewFileCachedWorkshopClient(path, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Workshop cache path: %s\n", path)
	workshopClient = client
	return workshopClient, nil
}
